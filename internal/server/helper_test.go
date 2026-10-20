package server

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/maxkaiser11/lolstats/internal/database"
	"github.com/maxkaiser11/lolstats/internal/riot"
)

// fakeRiot returns whatever data the test gives it, without any network calls.
type fakeRiot struct {
	account    *riot.Account
	matchIDs   []string
	matches    map[string]*riot.Match
	err        error
	matchCalls int
}

func (f *fakeRiot) AccountByRiotID(region, name, tag string) (*riot.Account, error) {
	return f.account, f.err
}

func (f *fakeRiot) MatchIDSByPUUID(puuid string, count int) ([]string, error) {
	return f.matchIDs, f.err
}

func (f *fakeRiot) MatchByID(id string) (*riot.Match, error) {
	f.matchCalls++
	return f.matches[id], f.err
}

// testDB connects to the test database, runs migrations and empties every table.
// It skips the test if TEST_DB_URL isn't set, so `go test` still works without Docker.
func testDB(t *testing.T) *sql.DB {
	t.Helper()
	_ = godotenv.Load("../../.env") // tests run from internal/server/
	dsn := os.Getenv("TEST_DB_URL")
	if dsn == "" {
		t.Skip("TEST_DB_URL not set")
	}

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	if err := database.Migrate(conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := conn.Exec("TRUNCATE users, sessions, riot_accounts, matches, match_participants CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return conn
}

// get sends a GET request through the handler, with an optional session cookie.
func get(h http.Handler, path string, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// postForm sends a form POST through the handler, with an optional session cookie.
func postForm(h http.Handler, path string, form url.Values, cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest("POST", path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// signupAndLogin creates a user, logs in and returns their session cookie.
func signupAndLogin(t *testing.T, h http.Handler, email string) *http.Cookie {
	t.Helper()
	form := url.Values{"email": {email}, "password": {"password123"}}

	if rec := postForm(h, "/signup", form, nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("signup %s: status %d", email, rec.Code)
	}
	rec := postForm(h, "/login", form, nil)
	for _, c := range rec.Result().Cookies() {
		if c.Name == "session_token" {
			return c
		}
	}
	t.Fatalf("login %s: no session cookie", email)
	return nil
}

// accountsFor returns the saved Riot accounts of the user with this email.
func accountsFor(t *testing.T, db *database.Store, email string) []database.RiotAccount {
	t.Helper()
	ctx := context.Background()
	user, err := db.GetUserByEmail(ctx, email)
	if err != nil {
		t.Fatalf("get user %s: %v", email, err)
	}
	accounts, err := db.ListRiotAccountsByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("list accounts for %s: %v", email, err)
	}
	return accounts
}
