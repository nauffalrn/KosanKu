// Command api menjalankan REST API Kosanku.
//
// File ini adalah "composition root": satu-satunya tempat semua dependency
// dirakit (config, database, repository, service, handler). Package lain
// tidak membuat dependency-nya sendiri, melainkan menerimanya dari sini.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // mendaftarkan driver "pgx" untuk database/sql
	"github.com/joho/godotenv"

	"kosanku-api/internal/config"
	"kosanku-api/internal/platform/database"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	// Muat .env jika ada (berguna di Windows yang tidak punya make).
	// Variabel yang sudah di-set di environment tidak ditimpa.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	slog.SetDefault(newLogger(cfg))

	// ctx dibatalkan saat menerima Ctrl+C atau SIGTERM (misalnya dari Docker).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           newRouter(db),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "addr", cfg.HTTPAddr, "env", cfg.Env)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
	}

	// Graceful shutdown: berhenti menerima request baru,
	// lalu tunggu request yang sedang berjalan selesai.
	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

// newLogger memakai format teks yang mudah dibaca saat development,
// dan JSON di environment lain supaya mudah diolah tool monitoring.
func newLogger(cfg config.Config) *slog.Logger {
	if cfg.IsDevelopment() {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
