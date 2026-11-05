package room

import (
	"net/http"

	"kosanku-api/internal/platform/httpx"
)

// Handler adalah adapter HTTP. Tugasnya hanya: baca request, panggil service,
// tulis respons. Tidak ada logika bisnis di sini.
type Handler struct {
	svc *Service
}

// NewHandler membuat Handler HTTP untuk modul room.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Register mendaftarkan route modul room. Prefix /v1 disiapkan untuk public API.
func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/rooms", h.create)
	mux.HandleFunc("GET /v1/rooms", h.list)
	mux.HandleFunc("GET /v1/rooms/{id}", h.get)
	mux.HandleFunc("PATCH /v1/rooms/{id}", h.update)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	room, err := h.svc.Create(r.Context(), req.toInput())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, toResponse(room))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.svc.List(r.Context())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toResponses(rooms))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	room, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toResponse(room))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	var req updateRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.WriteError(w, r, err)
		return
	}

	room, err := h.svc.Update(r.Context(), r.PathValue("id"), req.toInput())
	if err != nil {
		httpx.WriteError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, toResponse(room))
}
