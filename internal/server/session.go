package server

import (
	"context"
	"net/http"

	"github.com/maxkaiser11/lolstats/internal/database"
)

type ctxKey string

const userKey ctxKey = "user"

func (s *Server) optionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err == nil && cookie.Value != "" {
			user, err := s.db.GetUserBySession(r.Context(), hashToken(cookie.Value))
			if err == nil {
				r = r.WithContext(context.WithValue(r.Context(), userKey, user))
			}
		}
		next(w, r)
	}
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user, err := s.db.GetUserBySession(r.Context(), hashToken(cookie.Value))
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, user)
		next(w, r.WithContext(ctx))
	}
}

func currentUser(r *http.Request) (database.User, bool) {
	user, ok := r.Context().Value(userKey).(database.User)
	return user, ok
}
