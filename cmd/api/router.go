package main

import (
	"database/sql"
	"net/http"

	"kosanku-api/internal/health"
	"kosanku-api/internal/platform/httpx"
	"kosanku-api/internal/room"
)

// newRouter merakit semua modul. Menambah fitur baru cukup menambah
// beberapa baris di sini, tanpa mengubah modul lain (Open/Closed Principle).
func newRouter(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	health.NewHandler(db).Register(mux)

	roomService := room.NewService(room.NewPostgresRepository(db))
	room.NewHandler(roomService).Register(mux)

	return httpx.Chain(mux, httpx.RequestLogger, httpx.Recoverer)
}
