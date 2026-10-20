package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/maxkaiser11/lolstats/internal/database"
)

type signupPage struct {
	Error string
	Email string
}

func (s *Server) handleSignupPage(w http.ResponseWriter, r *http.Request) {
	s.tmpl.ExecuteTemplate(w, "signup.html", signupPage{})
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	if len(password) < 8 {
		s.tmpl.ExecuteTemplate(w, "signup.html", signupPage{Error: "Password must be at least 8 characters", Email: email})
		return
	}

	hash, err := hashPassword(password)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	_, err = s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	})
	if err != nil {
		log.Printf("create user failed: %v", err)
		s.tmpl.ExecuteTemplate(w, "signup.html", signupPage{Error: "That email is already registered", Email: email})
		return
	}

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}
