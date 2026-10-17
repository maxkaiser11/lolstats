package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/maxkaiser11/lolstats/internal/riot"
)

type Server struct {
	riot *riot.Client
} 


func main() {
	// Load env vars
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	s := &Server{riot: riot.NewClient(os.Getenv("RIOT_API_KEY"))}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /home", handlerHome)
	mux.HandleFunc("GET /account/{name}/{tag}", s.handleAccount)
	mux.HandleFunc("GET /matches/{name}/{tag}", s.handleMatchIDs)

	log.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func handlerHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to Goobs's a way to analyze your league data")
}

func (s *Server) handleAccount(w http.ResponseWriter, r *http.Request) {
	acc, err := s.riot.AccountByRiotID("europe", r.PathValue("name"), r.PathValue("tag"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	data ,err := json.Marshal(&acc)	
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (s *Server) handleMatchIDs(w http.ResponseWriter, r *http.Request) {
	acc, err := s.riot.AccountByRiotID("europe", r.PathValue("name"), r.PathValue("tag"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
	ids, err := s.riot.MatchIDSByPUUID(acc.PUUID, 5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
	
	var games []riot.Participant
	for _, id := range ids {
		m, err := s.riot.MatchByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		for _, p := range m.Info.Participants {
			if p.PUUID == acc.PUUID {
				games = append(games, p)
				break
			}
		}
	}

	data, err := json.Marshal(games)
	if err != nil {
		http.Error(w, "failed to marshal games", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
