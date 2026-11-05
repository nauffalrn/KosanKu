package room_test

import (
	"context"
	"errors"
	"testing"

	"kosanku-api/internal/apperr"
	"kosanku-api/internal/room"
)

func newService() *room.Service {
	return room.NewService(newFakeRepository())
}

func TestCreate_Success(t *testing.T) {
	svc := newService()

	got, err := svc.Create(context.Background(), room.CreateInput{Number: " a01 ", MonthlyPrice: 1_500_000})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Number != "A01" {
		t.Errorf("Number = %q, want normalized %q", got.Number, "A01")
	}
	if got.Status != room.StatusAvailable {
		t.Errorf("Status = %q, want %q", got.Status, room.StatusAvailable)
	}
}

func TestCreate_InvalidInputReturnsAllFieldErrors(t *testing.T) {
	svc := newService()

	_, err := svc.Create(context.Background(), room.CreateInput{Number: "", MonthlyPrice: 0})

	appErr, ok := apperr.As(err)
	if !ok || appErr.Kind != apperr.KindInvalid {
		t.Fatalf("err = %v, want invalid error", err)
	}
	for _, field := range []string{"number", "monthly_price"} {
		if _, exists := appErr.Fields[field]; !exists {
			t.Errorf("missing field error for %q", field)
		}
	}
}

func TestCreate_DuplicateNumber(t *testing.T) {
	svc := newService()
	ctx := context.Background()
	in := room.CreateInput{Number: "A01", MonthlyPrice: 1_000_000}

	if _, err := svc.Create(ctx, in); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if _, err := svc.Create(ctx, in); !errors.Is(err, room.ErrNumberTaken) {
		t.Fatalf("err = %v, want ErrNumberTaken", err)
	}
}

func TestUpdate_PartialChangeKeepsOtherFields(t *testing.T) {
	svc := newService()
	ctx := context.Background()
	created, _ := svc.Create(ctx, room.CreateInput{Number: "A01", MonthlyPrice: 1_000_000})

	newPrice := int64(1_250_000)
	got, err := svc.Update(ctx, created.ID, room.UpdateInput{MonthlyPrice: &newPrice})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.MonthlyPrice != newPrice {
		t.Errorf("MonthlyPrice = %d, want %d", got.MonthlyPrice, newPrice)
	}
	if got.Number != "A01" {
		t.Errorf("Number changed to %q, want unchanged", got.Number)
	}
}

func TestUpdate_InvalidStatus(t *testing.T) {
	svc := newService()
	ctx := context.Background()
	created, _ := svc.Create(ctx, room.CreateInput{Number: "A01", MonthlyPrice: 1_000_000})

	status := room.Status("rusak")
	_, err := svc.Update(ctx, created.ID, room.UpdateInput{Status: &status})

	if appErr, ok := apperr.As(err); !ok || appErr.Kind != apperr.KindInvalid {
		t.Fatalf("err = %v, want invalid error", err)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	svc := newService()

	_, err := svc.Update(context.Background(), "tidak-ada", room.UpdateInput{})
	if !errors.Is(err, room.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
