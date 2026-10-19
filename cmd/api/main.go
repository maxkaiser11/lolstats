package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"

	"github.com/joho/godotenv"
	"github.com/maxkaiser11/lolstats/internal/database"
	"github.com/maxkaiser11/lolstats/internal/riot"
	"github.com/maxkaiser11/lolstats/internal/server"
)


func main() {
	// Load env vars
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	conn, err := sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}

	mux := server.New(riot.NewClient(os.Getenv("RIOT_API_KEY")), database.New(conn))

	log.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", mux.Routes()))
}


