package server

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/maxkaiser11/lolstats/internal/database"
	"github.com/maxkaiser11/lolstats/internal/riot"
)

func TestLoginGivesAccess(t *testing.T) {
	conn := testDB(t)
	h := New(&fakeRiot{}, database.New(conn)).Routes()

	cookie := signupAndLogin(t, h, "max@example.com")

	if rec := get(h, "/accounts", cookie); rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
}

func TestLoginWrongCredentials(t *testing.T) {
	conn := testDB(t)
	h := New(&fakeRiot{}, database.New(conn)).Routes()
	signupAndLogin(t, h, "max@example.com")

	tests := []struct {
		name, email, password string
	}{
		{"wrong password", "max@example.com", "wrongpassword"},
		{"unknown email", "nobody@example.com", "password123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{"email": {tt.email}, "password": {tt.password}}
			rec := postForm(h, "/login", form, nil)

			// Both cases must show the same message, so attackers can't
			// find out which emails are registered.
			if !strings.Contains(rec.Body.String(), "Invalid email or password") {
				t.Error("missing error message")
			}
			for _, c := range rec.Result().Cookies() {
				if c.Name == "session_token" {
					t.Error("failed login still set a session cookie")
				}
			}
		})
	}
}

func TestLogoutDeletesSession(t *testing.T) {
	conn := testDB(t)
	h := New(&fakeRiot{}, database.New(conn)).Routes()

	cookie := signupAndLogin(t, h, "max@example.com")
	postForm(h, "/logout", nil, cookie)

	// Reuse the OLD cookie. If logout only cleared it in the browser,
	// the session would still be valid and this would return 200.
	if rec := get(h, "/accounts", cookie); rec.Code != http.StatusSeeOther {
		t.Fatalf("old session still works after logout: status %d", rec.Code)
	}
}

func TestInvalidRiotIDSavesNothing(t *testing.T) {
	conn := testDB(t)
	db := database.New(conn)
	h := New(&fakeRiot{err: errors.New("account not found")}, db).Routes()

	cookie := signupAndLogin(t, h, "max@example.com")
	postForm(h, "/accounts", url.Values{"name": {"Nobody"}, "tag": {"EUW"}}, cookie)

	if accounts := accountsFor(t, db, "max@example.com"); len(accounts) != 0 {
		t.Fatalf("saved %d accounts, want 0", len(accounts))
	}
}

func TestCannotDeleteOtherUsersAccount(t *testing.T) {
	conn := testDB(t)
	db := database.New(conn)
	fake := &fakeRiot{account: &riot.Account{PUUID: "puuid-a", GameName: "PlayerA", TagLine: "EUW"}}
	h := New(fake, db).Routes()

	alice := signupAndLogin(t, h, "alice@example.com")
	bob := signupAndLogin(t, h, "bob@example.com")

	postForm(h, "/accounts", url.Values{"name": {"PlayerA"}, "tag": {"EUW"}}, alice)
	accounts := accountsFor(t, db, "alice@example.com")
	if len(accounts) != 1 {
		t.Fatalf("setup: alice has %d accounts, want 1", len(accounts))
	}

	// Bob tries to delete Alice's account.
	postForm(h, "/accounts/"+accounts[0].ID.String()+"/delete", nil, bob)

	if accounts := accountsFor(t, db, "alice@example.com"); len(accounts) != 1 {
		t.Fatal("bob deleted alice's account")
	}
}
