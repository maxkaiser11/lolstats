package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type loginPage struct {
	Error string
	Email string
}

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

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.tmpl.ExecuteTemplate(w, "login.html", invalid)
		return
	}

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}
