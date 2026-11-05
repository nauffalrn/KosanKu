# Model Data Kosanku

Dokumen ini menjelaskan tabel inti dan alasan di balik setiap keputusan desain.

```
rooms ──< leases >── tenants
             │
             └──< invoices ──< invoice_items
                     │
                     └──< payments

settings (satu baris)
```

## Ringkasan tabel

| Tabel | Isi | Catatan penting |
|---|---|---|
| `rooms` | Kamar | Nomor kamar unik. |
| `tenants` | Penyewa | No. HP unik. KTP opsional dan sensitif. |
| `leases` | Kontrak sewa kamar–penyewa | Siklus tagihan fleksibel, harga di-snapshot, maks. 1 kontrak aktif per kamar. |
| `invoices` | Tagihan per periode | Maks. 1 tagihan per kontrak per periode. |
| `invoice_items` | Rincian tagihan | Total tagihan = jumlah item (dynamic billing). |
| `payments` | Pembayaran | Boleh cicil, immutable. |
| `settings` | Pengaturan kos | Tabel singleton (satu baris). |

## Keputusan desain

### 1. Siklus tagihan fleksibel: `billing_unit` + `billing_interval`

Kos bisa disewa mingguan, bulanan, beberapa bulan sekaligus, atau tahunan.
Dua kolom cukup untuk semua kasus itu:

| Kasus | `billing_unit` | `billing_interval` |
|---|---|---|
| Mingguan | `week` | 1 |
| Bulanan | `month` | 1 |
| Bayar tiap 3 bulan | `month` | 3 |
| Tahunan dibayar dimuka | `year` | 1 |

Jatuh tempo mengikuti tanggal mulai kontrak. Penyewa yang masuk tanggal 12
ditagih setiap tanggal 12.

**Kasus tepi:** kontrak mulai tanggal 31. Bulan yang tidak punya tanggal 31
memakai tanggal terakhir bulan itu. Aturan ini diimplementasikan dan di-test
di modul billing, bukan di database.

### 2. Harga di-snapshot ke kontrak (`leases.price_per_period`)

Harga di `rooms` adalah harga katalog. Saat kontrak dibuat, harga disalin ke
kontrak. Jika harga kamar naik bulan depan, penyewa lama tetap membayar sesuai
kontrak. Untuk sewa bulanan, nilai default diambil dari `rooms.monthly_price`;
untuk mingguan atau tahunan, admin mengisi harganya saat membuat kontrak.

### 3. Satu kamar, satu kontrak aktif

Dijaga oleh partial unique index `leases_one_active_per_room`. Aturan bisnis
penting harus dijaga database, bukan hanya di kode: dua request bersamaan
tetap tidak bisa membuat dua kontrak aktif untuk kamar yang sama.

### 4. Tagihan = header + item

`invoices` hanya menyimpan periode, jatuh tempo, dan status. Nominalnya ada di
`invoice_items`: sewa, deposit, listrik, air, denda, lain-lain. Karena itu:

- listrik bisa berbeda setiap bulan (diinput manual sesuai pemakaian);
- deposit masuk sebagai item di tagihan pertama, sehingga pembayarannya tercatat
  di `payments` seperti biasa;
- jenis biaya baru cukup menambah nilai `kind`, tanpa mengubah struktur tabel.

### 5. Periode setengah terbuka `[period_start, period_end)`

`period_end` adalah tanggal awal periode berikutnya. Contoh: 12 Nov – 12 Des
berarti mulai 12 Nov sampai sebelum 12 Des. Cara ini membuat periode tidak
tumpang tindih dan menghindari hitungan "tanggal terakhir bulan" yang rawan bug.

### 6. Pembayaran immutable dan bisa dicicil

Satu tagihan bisa punya banyak pembayaran. Pembayaran tidak pernah di-update:
salah catat dikoreksi dengan pencatatan baru. Ini praktik umum di sistem
keuangan supaya jejak audit utuh. Apakah cicilan diizinkan diatur oleh
`settings.allow_partial_payment`.

Status tagihan (`unpaid`, `partially_paid`, `paid`) diperbarui di dalam
transaksi yang sama dengan pencatatan pembayaran.

### 7. Listrik dan air: default dari settings, bisa diubah per kontrak

`settings.default_electricity_included` (default `false`) dan
`settings.default_water_included` (default `true`) dipakai sebagai nilai awal
saat membuat kontrak. Nilainya lalu disimpan di kontrak, sehingga mengubah
settings tidak mengubah kontrak yang sudah berjalan.

### 8. Deposit

`leases.deposit_amount` adalah deposit yang disepakati. Saat kontrak berakhir,
`deposit_refunded_amount` mencatat jumlah yang dikembalikan (boleh sebagian
jika ada potongan kerusakan). Database menolak pengembalian melebihi deposit.

## Yang sengaja belum dibuat (YAGNI)

| Fitur | Kapan ditambahkan |
|---|---|
| Diskon | Saat benar-benar dibutuhkan; bisa sebagai `kind` baru di `invoice_items`. |
| Tarif per kamar untuk mingguan/tahunan (`room_rates`) | Jika admin sering menyewakan non-bulanan dengan harga tetap. |
| Akun login penyewa | Bersama modul autentikasi. |
| Tabel webhook dan outbox | Saat mengerjakan fitur webhook. |
