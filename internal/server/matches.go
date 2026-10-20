package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/maxkaiser11/lolstats/internal/database"
	"github.com/maxkaiser11/lolstats/internal/riot"
)

// How many recent matches to sync from Riot and show on the page.
const matchesToShow = 20

type matchesPage struct {
	Name  string
	Tag   string
	User  *database.User
	Games []database.RecentGamesByPUUIDRow
	Stats []database.ChampionStatsByPUUIDRow
}

func (s *Server) handleMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	acc, err := s.riot.AccountByRiotID("europe", r.PathValue("name"), r.PathValue("tag"))
	if err != nil {
		log.Printf("lookup account: %v", err)
		http.Error(w, "Couldn't find that player", http.StatusBadGateway)
		return
	}

	// If Riot fails part-way through, still show whatever is already stored.
	if err := s.syncMatches(ctx, acc.PUUID); err != nil {
		log.Printf("sync matches for %s: %v", acc.PUUID, err)
	}

	games, err := s.db.RecentGamesByPUUID(ctx, database.RecentGamesByPUUIDParams{
		Puuid:    acc.PUUID,
		MaxGames: matchesToShow,
	})
	if err != nil {
		log.Printf("recent games: %v", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	stats, err := s.db.ChampionStatsByPUUID(ctx, acc.PUUID)
	if err != nil {
		log.Printf("champion stats: %v", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	page := matchesPage{Name: acc.GameName, Tag: acc.TagLine, Games: games, Stats: stats}
	if user, ok := currentUser(r); ok {
		page.User = &user
	}
	if err := s.tmpl.ExecuteTemplate(w, "matches.html", page); err != nil {
		log.Printf("matches template: %v", err)
	}
}

// syncMatches fetches the player's recent match IDs and stores any matches
// that aren't in the database yet.
func (s *Server) syncMatches(ctx context.Context, puuid string) error {
	ids, err := s.riot.MatchIDSByPUUID(puuid, matchesToShow)
	if err != nil {
		return fmt.Errorf("match ids: %w", err)
	}

	have, err := s.db.ExistingMatchIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("existing ids: %w", err)
	}
	known := make(map[string]bool, len(have))
	for _, id := range have {
		known[id] = true
	}

	for _, id := range ids {
		if known[id] {
			continue // already stored, no Riot call needed
		}
		m, err := s.riot.MatchByID(id)
		if err != nil {
			return fmt.Errorf("fetch match %s: %w", id, err)
		}
		if err := s.db.SaveMatch(ctx, matchParams(m), participantParams(m)); err != nil {
			return fmt.Errorf("save match %s: %w", id, err)
		}
	}
	return nil
}

// matchParams converts a Riot match into the row for the matches table.
func matchParams(m *riot.Match) database.InsertMatchParams {
	return database.InsertMatchParams{
		MatchID:      m.Metadata.MatchID,
		QueueID:      int32(m.Info.QueueID),
		GameDuration: int32(m.Info.GameDuration),
		GameCreation: time.UnixMilli(m.Info.GameCreation),
	}
}

// participantParams converts all 10 players into match_participants rows.
func participantParams(m *riot.Match) []database.InsertParticipantParams {
	players := make([]database.InsertParticipantParams, 0, len(m.Info.Participants))
	for _, p := range m.Info.Participants {
		players = append(players, database.InsertParticipantParams{
			MatchID:      m.Metadata.MatchID,
			Puuid:        p.PUUID,
			ChampionName: p.ChampionName,
			Kills:        int32(p.Kills),
			Deaths:       int32(p.Deaths),
			Assists:      int32(p.Assists),
			Win:          p.Win,
		})
	}
	return players
}
