package server

import (
	"net/http"

	"github.com/maxkaiser11/lolstats/internal/riot"
)

func (s *Server) handleMatchIDs(w http.ResponseWriter, r *http.Request) {
	acc, err := s.riot.AccountByRiotID("europe", r.PathValue("name"), r.PathValue("tag"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	ids, err := s.riot.MatchIDSByPUUID(acc.PUUID, 5)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
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

	data := struct {
		Name  string
		Games []riot.Participant
	}{
		Name:  r.PathValue("name"),
		Games: games,
	}

	if err := s.tmpl.ExecuteTemplate(w, "matches.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
