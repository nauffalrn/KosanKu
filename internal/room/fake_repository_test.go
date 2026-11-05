package room_test

import (
	"context"
	"fmt"
	"sort"
	"time"

	"kosanku-api/internal/room"
)

// fakeRepository menyimpan data di memori. Dipakai test supaya cepat dan
// tidak butuh database. Perilakunya meniru PostgresRepository.
type fakeRepository struct {
	rooms  map[string]room.Room
	nextID int
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{rooms: make(map[string]room.Room)}
}

func (f *fakeRepository) Create(_ context.Context, r room.Room) (room.Room, error) {
	if f.numberTaken(r.Number, "") {
		return room.Room{}, room.ErrNumberTaken
	}
	f.nextID++
	now := time.Now()
	r.ID = fmt.Sprintf("room-%d", f.nextID)
	r.CreatedAt, r.UpdatedAt = now, now
	f.rooms[r.ID] = r
	return r, nil
}

func (f *fakeRepository) List(_ context.Context) ([]room.Room, error) {
	out := make([]room.Room, 0, len(f.rooms))
	for _, r := range f.rooms {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Number < out[j].Number })
	return out, nil
}

func (f *fakeRepository) GetByID(_ context.Context, id string) (room.Room, error) {
	r, ok := f.rooms[id]
	if !ok {
		return room.Room{}, room.ErrNotFound
	}
	return r, nil
}

func (f *fakeRepository) Update(_ context.Context, r room.Room) (room.Room, error) {
	if _, ok := f.rooms[r.ID]; !ok {
		return room.Room{}, room.ErrNotFound
	}
	if f.numberTaken(r.Number, r.ID) {
		return room.Room{}, room.ErrNumberTaken
	}
	r.UpdatedAt = time.Now()
	f.rooms[r.ID] = r
	return r, nil
}

func (f *fakeRepository) numberTaken(number, exceptID string) bool {
	for id, r := range f.rooms {
		if r.Number == number && id != exceptID {
			return true
		}
	}
	return false
}
