package main

import (
	"log"
	"net/http"
	"os"

	"splendor/backend/internal/app"
)

func main() {
	addr := getEnv("APP_ADDR", ":8080")

	a := app.New()
	server := &http.Server{
		Addr:    addr,
		Handler: a.Routes(),
	}

	log.Printf("backend listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
