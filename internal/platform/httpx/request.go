package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"kosanku-api/internal/apperr"
)

// maxBodyBytes membatasi ukuran body supaya server tidak bisa dibanjiri request besar.
const maxBodyBytes = 1 << 20 // 1 MB

// DecodeJSON membaca body JSON ke dst dengan aturan ketat:
// ukuran dibatasi, field yang tidak dikenal ditolak, dan body harus berisi satu objek.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return apperr.Invalid(decodeErrorMessage(err), nil)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return apperr.Invalid("body hanya boleh berisi satu objek JSON", nil)
	}
	return nil
}

func decodeErrorMessage(err error) string {
	var maxBytesErr *http.MaxBytesError
	var typeErr *json.UnmarshalTypeError

	switch {
	case errors.Is(err, io.EOF):
		return "body tidak boleh kosong"
	case errors.As(err, &maxBytesErr):
		return "ukuran body terlalu besar"
	case errors.As(err, &typeErr):
		return "tipe data field " + typeErr.Field + " tidak sesuai"
	default:
		return "format JSON tidak valid: " + err.Error()
	}
}
