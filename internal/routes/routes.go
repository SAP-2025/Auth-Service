// routes/routes.go
package routes

import (
	"github.com/SAP-2025/auth-service/internal/handlers"
	custommiddleware "github.com/SAP-2025/auth-service/internal/middleware"
	"github.com/SAP-2025/auth-service/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func SetupRoutes(authService *services.AuthService) *chi.Mux {
	r := chi.NewRouter()

	// Built-in middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60))

	// Custom middleware
	r.Use(custommiddleware.CORS())

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok","redis":"connected"}`))
	})

	// Auth routes (public)
	r.Route("/auth", func(r chi.Router) {
		r.Get("/login", authHandler.Login)
		r.Get("/callback", authHandler.Callback)
		r.Delete("/cancel", authHandler.CancelLogin)
		r.Get("/session", authHandler.SessionStatus)

		// Protected auth routes
		r.Group(func(r chi.Router) {
			r.Use(custommiddleware.AuthMiddleware(authService))
			r.Get("/profile", authHandler.Profile)
		})
	})

	return r
}
