// Package database berisi helper koneksi dan error PostgreSQL.
//
// Package ini hanya memakai database/sql (standard library). Driver pgx
// didaftarkan di cmd/api/main.go lewat blank import, sehingga pilihan
// driver ada di composition root, bukan tersebar di package lain.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// DriverName harus sama dengan nama driver yang didaftarkan di main.go.
const DriverName = "pgx"

// Kode error PostgreSQL. Daftar lengkap:
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	codeUniqueViolation           = "23505"
	codeInvalidTextRepresentation = "22P02"
)

// Open membuka connection pool dan memastikan database bisa dihubungi.
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open(DriverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}

	// Batasi pool supaya aplikasi tidak menghabiskan koneksi PostgreSQL.
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}
	return db, nil
}

// IsUniqueViolation bernilai true jika err berasal dari constraint UNIQUE.
func IsUniqueViolation(err error) bool { return hasSQLState(err, codeUniqueViolation) }

// IsInvalidText bernilai true jika format input salah, misalnya UUID tidak valid.
func IsInvalidText(err error) bool { return hasSQLState(err, codeInvalidTextRepresentation) }

// hasSQLState mengecek kode error tanpa import pgx secara langsung.
// Error dari pgx (pgconn.PgError) punya method SQLState().
func hasSQLState(err error, code string) bool {
	var stateErr interface{ SQLState() string }
	return errors.As(err, &stateErr) && stateErr.SQLState() == code
}
