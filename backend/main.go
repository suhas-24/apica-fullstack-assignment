package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/suhas-24/apica-fullstack-assignment/backend/api"
	"github.com/suhas-24/apica-fullstack-assignment/backend/cache"
)

func main() {
	// Create a new LRU cache with capacity 100 and max memory 10MB
	lruCache := cache.NewLRUCache(5, 2*1024*1024)
	handler := api.NewHandler(lruCache)

	r := mux.NewRouter()
	r.HandleFunc("/api/cache/{key}", handler.GetHandler).Methods("GET")
	r.HandleFunc("/api/cache", handler.SetHandler).Methods("POST")
	r.HandleFunc("/api/cache/{key}", handler.DeleteHandler).Methods("DELETE")
	r.HandleFunc("/api/cache", handler.GetAllHandler).Methods("GET")
	r.HandleFunc("/ws", handler.HandleWebSocket)

	// Create a new CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // Allow requests from your React app
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            true, // Enable Debugging for testing, consider disabling in production
	})

	// Wrap the router with the CORS handler
	corsRouter := c.Handler(r)

	// Create a server with a reasonable timeout
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      corsRouter,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server in a goroutine
	go func() {
		log.Println("Server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}