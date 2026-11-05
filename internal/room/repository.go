package room

import "context"

// Repository adalah kontrak penyimpanan data kamar.
//
// Interface ini didefinisikan di sisi pemakai (service), bukan di sisi
// implementasi. Service hanya tahu kontrak ini, sehingga:
//   - unit test bisa memakai fake repository tanpa database;
//   - implementasi bisa diganti (misalnya ke sqlc) tanpa mengubah service.
type Repository interface {
	Create(ctx context.Context, room Room) (Room, error)
	List(ctx context.Context) ([]Room, error)
	// GetByID mengembalikan ErrNotFound jika kamar tidak ada.
	GetByID(ctx context.Context, id string) (Room, error)
	// Update mengembalikan ErrNotFound atau ErrNumberTaken.
	Update(ctx context.Context, room Room) (Room, error)
}
