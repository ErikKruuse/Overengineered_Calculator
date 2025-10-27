package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"erikkruuse/calculator/internal/api"
	service "erikkruuse/calculator/internal/services"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	// Respect Cloud Run port environment variable.
	port := getenv("PORT", "8080")
	addr := ":" + port

	// Create service layer (calculator + history)
	svc := service.NewCalculatorService(service.WithMaxHistory(100))

	// Wire up the API layer
	handler := http.NewServeMux()
	api.New(svc).RegisterRoutes(handler)

	// Basic request logging middleware
	logged := loggingMiddleware(handler)

	// Startup banner
	log.Printf("Starting Calculator API on %s ...", addr)
	log.Printf("Health check: http://localhost:%s/health", port)

	// Start HTTP server
	srv := &http.Server{
		Addr:              addr,
		Handler:           logged,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

// loggingMiddleware wraps an http.Handler to log simple request summaries.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		dur := time.Since(start)
		log.Printf("%s %s %s (%v)", r.RemoteAddr, r.Method, r.URL.Path, dur)
	})
}
