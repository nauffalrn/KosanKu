// Package httpx berisi helper HTTP yang dipakai semua modul:
// format respons JSON, pemetaan error, decode request, dan middleware.
//
// Semua respons memakai envelope yang sama supaya frontend dan
// pengguna public API cukup belajar satu format:
//
//	sukses: {"data": ...}
//	gagal:  {"error": {"code": "...", "message": "...", "fields": {...}}}
package httpx

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"kosanku-api/internal/apperr"
)

type envelope struct {
	Data  any        `json:"data,omitempty"`
	Error *errorBody `json:"error,omitempty"`
}

type errorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// WriteJSON menulis respons sukses dengan envelope {"data": ...}.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	write(w, status, envelope{Data: data})
}

// WriteError menerjemahkan error aplikasi menjadi respons HTTP.
// Error yang tidak dikenal selalu menjadi 500 dan detailnya hanya masuk log,
// tidak dikirim ke client, supaya informasi internal tidak bocor.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	status, body := toErrorResponse(err)
	if status >= http.StatusInternalServerError {
		slog.ErrorContext(r.Context(), "request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	}
	write(w, status, envelope{Error: &body})
}

func toErrorResponse(err error) (int, errorBody) {
	appErr, ok := apperr.As(err)
	if !ok {
		return http.StatusInternalServerError, errorBody{Code: "internal", Message: "terjadi kesalahan pada server"}
	}

	body := errorBody{Message: appErr.Message, Fields: appErr.Fields}
	switch appErr.Kind {
	case apperr.KindInvalid:
		body.Code = "invalid_argument"
		return http.StatusBadRequest, body
	case apperr.KindNotFound:
		body.Code = "not_found"
		return http.StatusNotFound, body
	case apperr.KindConflict:
		body.Code = "conflict"
		return http.StatusConflict, body
	default:
		return http.StatusInternalServerError, errorBody{Code: "internal", Message: "terjadi kesalahan pada server"}
	}
}

func write(w http.ResponseWriter, status int, body envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		// Header sudah terkirim, jadi yang bisa dilakukan hanya mencatat log.
		slog.Error("encode response failed", "err", err)
	}
}
