package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/maxkaiser11/lolstats/internal/riot"
)

type fakeRiot struct {
	account  *riot.Account
	matchIDs []string
	matches  map[string]*riot.Match
	err      error
}

func (f *fakeRiot) AccountByRiotID(region, name, tag string) (*riot.Account, error) {
	return f.account, f.err
}

func (f *fakeRiot) MatchIDSByPUUID(puuid string, count int) ([]string, error) {
	return f.matchIDs, f.err
}

func (f *fakeRiot) MatchByID(id string) (*riot.Match, error) {
	return f.matches[id], f.err
}

func TestMatchesPage(t *testing.T) {
	m := &riot.Match{}
	m.Info.GameDuration = 1805
	m.Info.Participants = []riot.Participant{
		{PUUID: "other-player", ChampionName: "Zed"},
		{PUUID: "faker-puuid", ChampionName: "Ahri", Kills: 10, Deaths: 2, Assists: 7, Win: true},
	}
	fake := &fakeRiot{
		account:  &riot.Account{PUUID: "faker-puuid", GameName: "Faker", TagLine: "KR1"},
		matchIDs: []string{"EUW1_123"},
		matches:  map[string]*riot.Match{"EUW1_123": m},
	}

	s := New(fake, nil)

	req := httptest.NewRequest("GET", "/matches/Faker/KR1", nil)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	body := rec.Body.String()
	for _, want := range []string{"Faker", "Ahri", "Victory", "30:05"} {
		if !strings.Contains(body, want) {
			t.Errorf("page is missing %q", want)
		}
	}
	if strings.Contains(body, "Zed") {
		t.Error("page shows another player's champion")
	}
}

func TestMatchesPageRiotError(t *testing.T) {
	s := New(&fakeRiot{err: errors.New("riot is down")}, nil)

	req := httptest.NewRequest("GET", "/matches/Faker/KR1", nil)
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}
