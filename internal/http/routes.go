package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(h *Handler) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestSize(5 << 20)) // 5MB max body
	r.Use(corsMiddleware)
	r.Use(securityHeaders)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/parts", func(r chi.Router) {
			r.Post("/", h.CreatePart)
			r.Get("/", h.ListParts)
			r.Get("/{id}", h.GetPart)
			r.Put("/{id}", h.UpdatePart)
			r.Delete("/{id}", h.DeletePart)
		})

		r.Route("/restock", func(r chi.Router) {
			r.Get("/priorities", h.GetRestockPriorities)
		})
	})

	r.Get("/health", h.HealthCheck)

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Content-Security-Policy", "default-src 'none'")

		next.ServeHTTP(w, r)
	})
}
