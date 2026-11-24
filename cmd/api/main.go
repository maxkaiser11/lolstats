package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/joho/godotenv"
	"github.com/maxkaiser11/lolstats/internal/database"
	"github.com/maxkaiser11/lolstats/internal/riot"
	"github.com/maxkaiser11/lolstats/internal/server"
)

func main() {
	// Load env vars
	_ = godotenv.Load()

	conn, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	if err := conn.Ping(); err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	if err := database.Migrate(conn); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	mux := server.New(riot.NewClient(os.Getenv("RIOT_API_KEY")), database.NewStore(conn))

	champs, err := riot.LoadChampions(context.Background())
	if err != nil {
		log.Printf("load champion data (names and icons disabled): %v", err)
	} else {
		mux.SetChampions(champs)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux.Routes(),
		// Timeouts stop slow or stuck clients from holding connections open.
		// WriteTimeout leaves room for a first visit that fetches up to 21 matches from Riot.
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	log.Println("Listening on port " + port)
	log.Fatal(srv.ListenAndServe())
}
