# Kosanku API

Backend REST API untuk Kosanku, sistem manajemen kos untuk pemilik kos:
kamar, penyewa, tagihan, arus kas, dan tiket keluhan.

Backend ini dirancang supaya nanti bisa diakses dari tiga "pintu" yang memakai
logika bisnis yang sama: REST API (frontend), MCP server (AI assistant), dan
webhook (integrasi milik user).

---

## Tech stack dan alasannya

| Bagian | Pilihan | Alasan |
|---|---|---|
| Bahasa | Go 1.24+ | Cepat, sederhana, standard library lengkap. Go 1.24 dibutuhkan untuk `go tool`. |
| Router | `net/http` (standard library) | Sejak Go 1.22 sudah mendukung `GET /rooms/{id}`. Tidak perlu framework (KISS). |
| Database | PostgreSQL 17 | Relasional, cocok untuk data tagihan dan pembayaran yang butuh konsistensi. |
| Akses DB | `database/sql` + driver `pgx` | SQL ditulis eksplisit, mudah dipelajari. Bisa diganti ke `sqlc` nanti tanpa mengubah service. |
| Migration | `goose` (dipasang lewat `go tool`) | Versi tercatat di `go.mod`, jadi semua developer memakai versi yang sama. |
| Logging | `log/slog` (standard library) | Structured logging tanpa dependency tambahan. |
| Lint | `golangci-lint` v2 | Menjaga konsistensi gaya dan menangkap bug umum. |

---

## Struktur folder

```
kosanku-api/
├── cmd/
│   └── api/                 # entry point REST API
│       ├── main.go          # composition root: config, DB, server, graceful shutdown
│       └── router.go        # merakit semua modul fitur
├── internal/                # kode yang tidak bisa di-import project lain
│   ├── apperr/              # error aplikasi yang tidak bergantung pada HTTP
│   ├── config/              # baca environment variable
│   ├── health/              # /healthz dan /readyz
│   ├── platform/            # infrastruktur bersama
│   │   ├── database/        # koneksi PostgreSQL + helper error
│   │   └── httpx/           # respons JSON, decode request, middleware
│   └── room/                # modul fitur: kamar (contoh pola untuk modul lain)
├── migrations/              # file SQL migration (goose)
├── docker-compose.yml       # PostgreSQL untuk development
├── Makefile                 # perintah sehari-hari
└── .env.example             # contoh konfigurasi
```

### Kenapa dikelompokkan per fitur (`room/`, nanti `tenant/`, `billing/`)?

Alternatifnya adalah per layer (`handlers/`, `services/`, `repositories/`).
Masalahnya, untuk mengubah satu fitur kamu harus membuka banyak folder.
Dengan per fitur, semua yang berhubungan dengan kamar ada di satu folder.
Ini membuat project tetap mudah dinavigasi saat fiturnya bertambah banyak.

### Susunan file di setiap modul fitur

| File | Tanggung jawab | Boleh bergantung pada |
|---|---|---|
| `room.go` | Entity, aturan validasi, error modul | `apperr` |
| `repository.go` | Interface penyimpanan data | (tidak ada) |
| `service.go` | Use case: alur bisnis | `Repository` (interface) |
| `postgres.go` | Implementasi `Repository` dengan SQL | `database/sql`, `platform/database` |
| `dto.go` | Bentuk JSON request dan respons | entity |
| `handler.go` | Adapter HTTP | `Service`, `platform/httpx` |

**Aturan utama: dependency selalu mengarah ke dalam.**
Handler memanggil Service. Service memanggil Repository lewat interface.
Service tidak pernah tahu soal HTTP atau SQL. Karena itu, nanti MCP server
cukup memanggil `room.Service` yang sama, tanpa menulis ulang logika bisnis.

```
HTTP handler ─┐
              ├──► Service ──► Repository (interface) ◄── PostgresRepository
MCP (nanti) ──┘
```

### Prinsip yang diterapkan

- **Single Responsibility**: tiap file punya satu tugas (lihat tabel di atas).
- **Open/Closed**: menambah modul baru cukup menambah baris di `cmd/api/router.go`.
- **Dependency Inversion**: Service bergantung pada interface `Repository`,
  sehingga unit test memakai fake repository tanpa database.
- **DRY**: format respons, pemetaan error, dan decode JSON hanya ada di `httpx`.
- **KISS / YAGNI**: tanpa framework, tanpa ORM, tanpa fitur yang belum dibutuhkan
  (pagination, auth, cache ditambahkan saat memang perlu).

---

## Menjalankan di laptop

### Prasyarat

- Go 1.24 atau lebih baru
- Docker Desktop (untuk PostgreSQL)
- `make` (Windows: pakai Git Bash/WSL, atau jalankan perintah di Makefile secara manual)
- `golangci-lint` v2 (opsional, untuk `make lint`)

### Langkah pertama kali

```bash
cp .env.example .env     # buat konfigurasi lokal
make setup               # download pgx dan goose, isi go.sum
make db-up               # jalankan PostgreSQL di Docker
make migrate-up          # buat tabel
make run                 # jalankan API di http://localhost:8080
```

### Coba endpoint

```bash
# health check
curl localhost:8080/readyz

# buat kamar
curl -X POST localhost:8080/v1/rooms \
  -H "Content-Type: application/json" \
  -d '{"number":"A01","monthly_price":1500000}'

# daftar kamar
curl localhost:8080/v1/rooms

# ubah sebagian data kamar (ganti {id} dengan id dari respons sebelumnya)
curl -X PATCH localhost:8080/v1/rooms/{id} \
  -H "Content-Type: application/json" \
  -d '{"status":"maintenance"}'
```

### Format respons

Semua endpoint memakai format yang sama:

```json
{ "data": { "id": "…", "number": "A01", "monthly_price": 1500000, "status": "available" } }
```

```json
{
  "error": {
    "code": "invalid_argument",
    "message": "data kamar tidak valid",
    "fields": { "monthly_price": "harus lebih dari 0" }
  }
}
```

| `code` | Status HTTP | Arti |
|---|---|---|
| `invalid_argument` | 400 | Input tidak valid |
| `not_found` | 404 | Data tidak ditemukan |
| `conflict` | 409 | Bentrok dengan data lain, misalnya nomor kamar sudah dipakai |
| `internal` | 500 | Kesalahan server (detail hanya ada di log) |

---

## Perintah sehari-hari

Jalankan `make` tanpa argumen untuk melihat semua perintah.

| Perintah | Fungsi |
|---|---|
| `make test` | Jalankan semua test |
| `make cover` | Test + laporan coverage |
| `make lint` | Jalankan linter |
| `make fmt` | Rapikan format kode |
| `make migrate-new name=create_tenants` | Buat file migration baru |
| `make migrate-status` | Lihat migration yang sudah/belum jalan |

---

## Checklist menambah modul fitur baru

Contoh: modul `tenant`.

1. `make migrate-new name=create_tenants`, lalu tulis SQL `Up` dan `Down`.
2. Buat folder `internal/tenant/` dengan file yang sama seperti `internal/room/`.
3. Tulis aturan bisnis di `tenant.go` dan test-nya dulu di `service_test.go`.
4. Implementasikan `postgres.go`, lalu `handler.go` dan `dto.go`.
5. Daftarkan di `cmd/api/router.go`.
6. Jalankan `make test` dan `make lint` sebelum commit.

---

## Roadmap

- [x] Fondasi project + modul `room`
- [ ] Modul `tenant`, `lease`, `billing` (invoice + payment), `ticket`
- [ ] Autentikasi admin
- [ ] Dokumentasi OpenAPI (kontrak untuk frontend)
- [ ] MCP server lokal (stdio) dengan tool baca
- [ ] MCP server remote + OAuth
- [ ] Public API dengan API key + webhook (outbox, queue, signature HMAC)
