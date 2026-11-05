package httpx

import (
	"errors"
	"net/http"
	"testing"

	"kosanku-api/internal/apperr"
)

func TestToErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{"invalid", apperr.Invalid("salah", nil), http.StatusBadRequest, "invalid_argument"},
		{"not found", apperr.NotFound("tidak ada"), http.StatusNotFound, "not_found"},
		{"conflict", apperr.Conflict("bentrok"), http.StatusConflict, "conflict"},
		{"unknown error is hidden", errors.New("koneksi db putus"), http.StatusInternalServerError, "internal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := toErrorResponse(tt.err)
			if status != tt.wantStatus {
				t.Errorf("status = %d, want %d", status, tt.wantStatus)
			}
			if body.Code != tt.wantCode {
				t.Errorf("code = %q, want %q", body.Code, tt.wantCode)
			}
		})
	}
}
