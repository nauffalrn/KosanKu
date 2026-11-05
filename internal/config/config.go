// Package config membaca konfigurasi aplikasi dari environment variable.
//
// Semua konfigurasi dibaca di satu tempat saat startup. Package lain
// menerima nilai yang sudah jadi, bukan membaca os.Getenv sendiri-sendiri.
package config

import (
	"errors"
	"os"
	"time"
)

// Config adalah seluruh konfigurasi yang dibutuhkan aplikasi.
type Config struct {
	Env             string
	HTTPAddr        string
	DatabaseURL     string
	ShutdownTimeout time.Duration
}

// IsDevelopment bernilai true saat aplikasi berjalan di laptop developer.
func (c Config) IsDevelopment() bool { return c.Env == "development" }

// Load membaca konfigurasi dan langsung gagal jika ada nilai wajib yang kosong.
// Gagal saat startup jauh lebih baik daripada gagal di tengah request.
func Load() (Config, error) {
	cfg := Config{
		Env:             getEnv("APP_ENV", "development"),
		HTTPAddr:        getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		ShutdownTimeout: 10 * time.Second,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("config: DATABASE_URL wajib diisi")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
