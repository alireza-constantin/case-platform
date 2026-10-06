package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/alireza-constantin/case-platform/server/internal/assembly"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://case_platform:local_only@127.0.0.1:54329/case_platform?sslmode=disable"
	}
	origin := os.Getenv("PUBLIC_ORIGIN")
	if origin == "" {
		origin = "http://localhost:5173"
	}
	parsed, err := url.Parse(origin)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.User != nil {
		log.Fatal("PUBLIC_ORIGIN must be an HTTP(S) origin without a path")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal("Invalid database configuration")
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatal("PostgreSQL unavailable; start the local database")
	}
	server := &http.Server{Addr: "127.0.0.1:8080", Handler: assembly.Handler(pool, origin, parsed.Scheme == "https"), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	log.Print("API listening on http://127.0.0.1:8080")
	log.Fatal(server.ListenAndServe())
}
