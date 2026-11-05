// Package health menyediakan endpoint untuk memantau kondisi server.
package health

import (
	"context"
	"net/http"
	"time"

	"kosanku-api/internal/platform/httpx"
)

// Pinger cukup satu method, jadi handler ini tidak bergantung langsung pada *sql.DB.
type Pinger interface {
	PingContext(ctx context.Context) error
}

// Handler melayani endpoint health check.
type Handler struct {
	db Pinger
}

// NewHandler membuat Handler health check.
func NewHandler(db Pinger) *Handler { return &Handler{db: db} }

// Register mendaftarkan route health check.
//   - /healthz: proses server hidup (liveness).
//   - /readyz: server siap menerima request, termasuk database (readiness).
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", h.liveness)
	mux.HandleFunc("GET /readyz", h.readiness)
}

func (h *Handler) liveness(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "database unavailable"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
