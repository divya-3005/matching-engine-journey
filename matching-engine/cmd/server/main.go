// server starts the matching engine HTTP server.
//
// Usage:
//
//	go run ./cmd/server
//
// The server listens on :8080.
// Press Ctrl+C or send SIGINT to trigger graceful shutdown.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/divya-3005/matching-engine/internal/engine"
)

func main() {
	// capacity 1: the HTTP layer submits one order then immediately calls
	// ProcessNext while holding the mutex, so the ring buffer never holds
	// more than one order at a time.
	e := engine.New(1)
	s := &Server{engine: e}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /orders", s.submitOrder)
	mux.HandleFunc("DELETE /orders/{id}", s.cancelOrder)
	mux.HandleFunc("GET /book", s.getBook)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// signal.NotifyContext cancels ctx on the first SIGINT (Ctrl+C).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		fmt.Println("matching engine listening on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server: %v\n", err)
			os.Exit(1)
		}
	}()

	// Block until interrupted.
	<-ctx.Done()
	fmt.Println("\nshutting down...")

	// Give in-flight requests up to 5 seconds to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "shutdown: %v\n", err)
	}

	fmt.Println("done")
}
