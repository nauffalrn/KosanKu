package room

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"kosanku-api/internal/platform/database"
)

// PostgresRepository adalah implementasi Repository untuk PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository membuat repository kamar berbasis PostgreSQL.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Pastikan saat compile bahwa PostgresRepository memenuhi kontrak Repository.
var _ Repository = (*PostgresRepository)(nil)

const selectColumns = "id, number, monthly_price, status, created_at, updated_at"

// Create menyimpan kamar baru. ID dan timestamp dibuat oleh database.
func (r *PostgresRepository) Create(ctx context.Context, room Room) (Room, error) {
	const query = `
		INSERT INTO rooms (number, monthly_price, status)
		VALUES ($1, $2, $3)
		RETURNING ` + selectColumns

	created, err := scanRoom(r.db.QueryRowContext(ctx, query, room.Number, room.MonthlyPrice, string(room.Status)))
	if err != nil {
		return Room{}, mapError("create", err)
	}
	return created, nil
}

// List mengembalikan semua kamar, diurutkan berdasarkan nomor kamar.
func (r *PostgresRepository) List(ctx context.Context) ([]Room, error) {
	const query = `SELECT ` + selectColumns + ` FROM rooms ORDER BY number`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("room: list: %w", err)
	}
	defer rows.Close()

	rooms := make([]Room, 0)
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, fmt.Errorf("room: list scan: %w", err)
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("room: list rows: %w", err)
	}
	return rooms, nil
}

// GetByID mengambil satu kamar. Mengembalikan ErrNotFound jika tidak ada.
func (r *PostgresRepository) GetByID(ctx context.Context, id string) (Room, error) {
	const query = `SELECT ` + selectColumns + ` FROM rooms WHERE id = $1`

	room, err := scanRoom(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		return Room{}, mapError("get", err)
	}
	return room, nil
}

// Update menyimpan perubahan kamar dan memperbarui updated_at.
func (r *PostgresRepository) Update(ctx context.Context, room Room) (Room, error) {
	const query = `
		UPDATE rooms
		SET number = $2, monthly_price = $3, status = $4, updated_at = now()
		WHERE id = $1
		RETURNING ` + selectColumns

	updated, err := scanRoom(r.db.QueryRowContext(ctx, query,
		room.ID, room.Number, room.MonthlyPrice, string(room.Status)))
	if err != nil {
		return Room{}, mapError("update", err)
	}
	return updated, nil
}

// scanner dipenuhi oleh *sql.Row dan *sql.Rows, jadi scanRoom bisa dipakai keduanya (DRY).
type scanner interface {
	Scan(dest ...any) error
}

func scanRoom(s scanner) (Room, error) {
	var room Room
	err := s.Scan(&room.ID, &room.Number, &room.MonthlyPrice, &room.Status, &room.CreatedAt, &room.UpdatedAt)
	return room, err
}

// mapError menerjemahkan error database menjadi error milik modul room.
// Detail PostgreSQL berhenti di sini dan tidak bocor ke service.
func mapError(op string, err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows), database.IsInvalidText(err):
		// ID dengan format UUID yang salah juga dianggap "tidak ditemukan".
		return ErrNotFound
	case database.IsUniqueViolation(err):
		return ErrNumberTaken
	default:
		return fmt.Errorf("room: %s: %w", op, err)
	}
}
