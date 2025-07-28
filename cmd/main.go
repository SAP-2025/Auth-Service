// main.go
package main

import (
	"github.com/SAP-2025/auth-service/internal/config"
	"github.com/SAP-2025/auth-service/internal/routes"
	"github.com/SAP-2025/auth-service/internal/services"
	"github.com/SAP-2025/auth-service/pkg"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Initialize Redis
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	redisClient := pkg.NewRedisClient(cfg)
	log.Println("Connected to Redis successfully")

	// Initialize stores and services
	pkceStore := services.NewPKCEStore(redisClient)
	authService := services.NewAuthService(pkceStore, cfg)

	// Setup routes
	router := routes.SetupRoutes(authService)

	// Create server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")
	// Add cleanup logic here if needed

	log.Println("Server stopped")
}
