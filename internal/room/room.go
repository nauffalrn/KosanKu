// Package room mengelola data kamar kos.
//
// Susunan file di setiap modul fitur selalu sama:
//
//	room.go       entity, aturan bisnis, error milik modul
//	repository.go kontrak penyimpanan data (interface)
//	service.go    use case: orkestrasi aturan bisnis + repository
//	postgres.go   implementasi repository untuk PostgreSQL
//	handler.go    adapter HTTP: request JSON -> service -> respons JSON
//	dto.go        bentuk JSON request/respons
package room

import (
	"strings"
	"time"
	"unicode/utf8"

	"kosanku-api/internal/apperr"
)

// Status kamar.
type Status string

// Daftar status kamar yang valid.
const (
	StatusAvailable   Status = "available"
	StatusOccupied    Status = "occupied"
	StatusMaintenance Status = "maintenance"
)

// Valid bernilai true jika status termasuk status yang dikenal.
func (s Status) Valid() bool {
	switch s {
	case StatusAvailable, StatusOccupied, StatusMaintenance:
		return true
	default:
		return false
	}
}

const maxNumberLength = 20

// Error milik modul room. Repository dan service mengembalikan error ini,
// test cukup mengecek dengan errors.Is.
var (
	ErrNotFound    = apperr.NotFound("kamar tidak ditemukan")
	ErrNumberTaken = apperr.Conflict("nomor kamar sudah dipakai")
)

// Room adalah satu kamar di kos.
type Room struct {
	ID     string
	Number string
	// MonthlyPrice dalam rupiah. Uang selalu disimpan sebagai bilangan bulat,
	// jangan float, karena float tidak presisi untuk perhitungan uang.
	MonthlyPrice int64
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CreateInput adalah data yang dibutuhkan untuk membuat kamar baru.
type CreateInput struct {
	Number       string
	MonthlyPrice int64
}

// UpdateInput memakai pointer: nil berarti field tersebut tidak diubah.
type UpdateInput struct {
	Number       *string
	MonthlyPrice *int64
	Status       *Status
}

// Validate mengecek seluruh aturan bisnis kamar dan mengumpulkan semua
// kesalahan sekaligus, supaya frontend bisa menampilkan semua error form sekaligus.
func (r Room) Validate() error {
	fields := make(map[string]string)

	switch {
	case r.Number == "":
		fields["number"] = "wajib diisi"
	case utf8.RuneCountInString(r.Number) > maxNumberLength:
		fields["number"] = "maksimal 20 karakter"
	}
	if r.MonthlyPrice <= 0 {
		fields["monthly_price"] = "harus lebih dari 0"
	}
	if !r.Status.Valid() {
		fields["status"] = "harus salah satu dari: available, occupied, maintenance"
	}

	if len(fields) > 0 {
		return apperr.Invalid("data kamar tidak valid", fields)
	}
	return nil
}

// apply menerapkan perubahan sebagian dari UpdateInput.
func (r *Room) apply(in UpdateInput) {
	if in.Number != nil {
		r.Number = normalizeNumber(*in.Number)
	}
	if in.MonthlyPrice != nil {
		r.MonthlyPrice = *in.MonthlyPrice
	}
	if in.Status != nil {
		r.Status = *in.Status
	}
}

// normalizeNumber menyeragamkan nomor kamar, misalnya " a01 " menjadi "A01".
func normalizeNumber(number string) string {
	return strings.ToUpper(strings.TrimSpace(number))
}
