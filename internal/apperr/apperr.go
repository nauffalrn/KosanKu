// Package apperr berisi error aplikasi yang TIDAK bergantung pada transport.
//
// Service hanya tahu "jenis" error (tidak valid, tidak ditemukan, konflik).
// Layer transport (HTTP sekarang, MCP nanti) yang menerjemahkan jenis itu
// ke format masing-masing, misalnya status code HTTP. Dengan begitu logika
// bisnis tidak perlu tahu apa-apa soal HTTP.
package apperr

import "errors"

// Kind adalah kategori error yang dipakai transport untuk memilih respons.
type Kind int

// Daftar Kind. KindInternal adalah nilai default untuk error yang tidak dikenal.
const (
	KindInternal Kind = iota
	KindInvalid
	KindNotFound
	KindConflict
)

// Error adalah error aplikasi dengan kategori dan pesan yang aman ditampilkan ke user.
type Error struct {
	Kind    Kind
	Message string
	// Fields berisi detail error per field input, misalnya untuk form validation.
	Fields map[string]string
}

func (e *Error) Error() string { return e.Message }

// Invalid dipakai ketika input dari user tidak valid.
func Invalid(message string, fields map[string]string) *Error {
	return &Error{Kind: KindInvalid, Message: message, Fields: fields}
}

// NotFound dipakai ketika data yang diminta tidak ada.
func NotFound(message string) *Error {
	return &Error{Kind: KindNotFound, Message: message}
}

// Conflict dipakai ketika aksi bertabrakan dengan data yang sudah ada.
func Conflict(message string) *Error {
	return &Error{Kind: KindConflict, Message: message}
}

// As mengambil *Error dari rantai error. ok bernilai false untuk error tak dikenal.
func As(err error) (appErr *Error, ok bool) {
	ok = errors.As(err, &appErr)
	return appErr, ok
}
