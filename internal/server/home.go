package server

import (
	"net/http"

	"github.com/maxkaiser11/lolstats/internal/database"
)

type homePage struct {
	Title string
	User  *database.User
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	page := homePage{}

	if user, ok := currentUser(r); ok {
		page.User = &user
	}

	s.tmpl.ExecuteTemplate(w, "index.html", page)
}
