package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/maxkaiser11/lolstats/internal/database"
)

type loginPage struct {
	Error string
	Email string
}

const sessionDuration = 30 * 24 * time.Hour

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	s.tmpl.ExecuteTemplate(w, "login.html", loginPage{})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	invalid := loginPage{Error: "Invalid email or password", Email: email}

	user, err := s.db.GetUserByEmail(context.Background(), email)
	if errors.Is(err, sql.ErrNoRows) {
		s.tmpl.ExecuteTemplate(w, "login.html", invalid)
		return
	}
	if err != nil {
		s.tmpl.ExecuteTemplate(w, "login.html", invalid)
		return
	}

	if !checkPasswordHash(password, string(user.PasswordHash)) {
		s.tmpl.ExecuteTemplate(w, "login.html", invalid)
		return
	}

	sessionToken := generateToken(32)
	expires := time.Now().Add(sessionDuration)

	s.db.CreateSession(context.Background(), database.CreateSessionParams{
		TokenHash: hashToken(sessionToken),
		UserID: user.ID,
		ExpiresAt: time.Now().Add(sessionDuration),
	})

	// Set Session cookie
	http.SetCookie(w, &http.Cookie{
		Name: "session_token",
		Value: sessionToken,
		Path: "/",
		Expires: expires,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})


	http.Redirect(w, r, "/home", http.StatusSeeOther)
}
