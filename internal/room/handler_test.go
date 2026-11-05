package room_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kosanku-api/internal/room"
)

func newTestServer() *http.ServeMux {
	mux := http.NewServeMux()
	room.NewHandler(room.NewService(newFakeRepository())).Register(mux)
	return mux
}

func doRequest(mux http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestHandler(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{"create success", http.MethodPost, "/v1/rooms", `{"number":"A01","monthly_price":1500000}`, http.StatusCreated, `"number":"A01"`},
		{"create validation error", http.MethodPost, "/v1/rooms", `{"number":"","monthly_price":0}`, http.StatusBadRequest, `"code":"invalid_argument"`},
		{"create rejects unknown field", http.MethodPost, "/v1/rooms", `{"number":"A01","harga":1}`, http.StatusBadRequest, `"code":"invalid_argument"`},
		{"create rejects empty body", http.MethodPost, "/v1/rooms", ``, http.StatusBadRequest, `body tidak boleh kosong`},
		{"list empty returns array", http.MethodGet, "/v1/rooms", ``, http.StatusOK, `"data":[]`},
		{"get not found", http.MethodGet, "/v1/rooms/tidak-ada", ``, http.StatusNotFound, `"code":"not_found"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(newTestServer(), tt.method, tt.path, tt.body)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("body = %s, want to contain %s", rec.Body, tt.wantBody)
			}
		})
	}
}

func TestHandler_CreateThenGet(t *testing.T) {
	mux := newTestServer()
	doRequest(mux, http.MethodPost, "/v1/rooms", `{"number":"B02","monthly_price":900000}`)

	// fakeRepository memberi ID berurutan: room-1, room-2, ...
	rec := doRequest(mux, http.MethodGet, "/v1/rooms/room-1", "")

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"number":"B02"`) {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body)
	}
}
