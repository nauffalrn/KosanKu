package room

import "context"

// Service berisi use case kamar. Ini satu-satunya tempat logika bisnis kamar.
// HTTP handler sekarang, dan MCP server nanti, sama-sama memanggil Service ini.
type Service struct {
	repo Repository
}

// NewService membuat Service dengan repository yang diberikan (dependency injection).
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create membuat kamar baru dengan status awal available.
func (s *Service) Create(ctx context.Context, in CreateInput) (Room, error) {
	room := Room{
		Number:       normalizeNumber(in.Number),
		MonthlyPrice: in.MonthlyPrice,
		Status:       StatusAvailable,
	}
	if err := room.Validate(); err != nil {
		return Room{}, err
	}
	return s.repo.Create(ctx, room)
}

// List mengembalikan semua kamar.
func (s *Service) List(ctx context.Context) ([]Room, error) {
	return s.repo.List(ctx)
}

// Get mengembalikan satu kamar berdasarkan ID.
func (s *Service) Get(ctx context.Context, id string) (Room, error) {
	return s.repo.GetByID(ctx, id)
}

// Update mengubah sebagian data kamar, lalu memvalidasi hasil akhirnya.
func (s *Service) Update(ctx context.Context, id string, in UpdateInput) (Room, error) {
	room, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Room{}, err
	}

	room.apply(in)
	if err := room.Validate(); err != nil {
		return Room{}, err
	}
	return s.repo.Update(ctx, room)
}
