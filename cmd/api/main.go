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
	_ = godotenv.Load()

	conn, err := sql.Open("postgres", os.Getenv("DB_URL"))
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
	log.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", mux.Routes()))
}


