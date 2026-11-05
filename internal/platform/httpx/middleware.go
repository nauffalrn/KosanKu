package httpx

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// Middleware membungkus handler untuk menambah perilaku lintas fitur.
type Middleware func(http.Handler) http.Handler

// Chain memasang middleware. Middleware pertama di daftar menjadi lapisan terluar.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// Recoverer menangkap panic supaya satu request bermasalah tidak mematikan server.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.ErrorContext(r.Context(), "panic recovered", "panic", rec, "stack", string(debug.Stack()))
				write(w, http.StatusInternalServerError, envelope{
					Error: &errorBody{Code: "internal", Message: "terjadi kesalahan pada server"},
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequestLogger mencatat setiap request beserta status dan durasinya.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(recorder, r)

		slog.InfoContext(r.Context(), "http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

// statusRecorder menyimpan status code yang ditulis handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Unwrap membuat http.ResponseController tetap bisa mengakses writer asli.
func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }
