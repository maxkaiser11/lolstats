package server

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/maxkaiser11/lolstats/internal/database"
	"github.com/maxkaiser11/lolstats/internal/riot"
)

// fakeWithOneMatch returns a fake Riot API where Faker has played one game.
func fakeWithOneMatch() *fakeRiot {
	m := &riot.Match{}
	m.Metadata.MatchID = "EUW1_123"
	m.Info.GameCreation = 1760000000000 // Unix milliseconds
	m.Info.GameDuration = 1805          // 30:05
	m.Info.QueueID = 420
	m.Info.Participants = []riot.Participant{
		{PUUID: "other-player", ChampionName: "Zed"},
		{PUUID: "faker-puuid", ChampionName: "Ahri", Kills: 10, Deaths: 2, Assists: 7, Win: true},
	}

	return &fakeRiot{
		account:  &riot.Account{PUUID: "faker-puuid", GameName: "Faker", TagLine: "KR1"},
		matchIDs: []string{"EUW1_123"},
		matches:  map[string]*riot.Match{"EUW1_123": m},
	}
}

func TestMatchesPage(t *testing.T) {
	conn := testDB(t)
	h := New(fakeWithOneMatch(), database.NewStore(conn)).Routes()

	rec := get(h, "/matches/Faker/KR1", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	body := rec.Body.String()
	for _, want := range []string{"Faker", "Ahri", "Victory", "30:05", "100%"} {
		if !strings.Contains(body, want) {
			t.Errorf("page is missing %q", want)
		}
	}
	if strings.Contains(body, "Zed") {
		t.Error("page shows another player's champion")
	}
}

func TestMatchesAreCached(t *testing.T) {
	conn := testDB(t)
	fake := fakeWithOneMatch()
	h := New(fake, database.NewStore(conn)).Routes()

	get(h, "/matches/Faker/KR1", nil)
	get(h, "/matches/Faker/KR1", nil)

	if fake.matchCalls != 1 {
		t.Fatalf("MatchByID called %d times, want 1 (second visit should use the database)", fake.matchCalls)
	}
}

func TestMatchesPageRiotError(t *testing.T) {
	// The account lookup fails before the database is touched, so nil is fine here.
	h := New(&fakeRiot{err: errors.New("riot is down")}, nil).Routes()

	if rec := get(h, "/matches/Faker/KR1", nil); rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}
