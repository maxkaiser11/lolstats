package server

import (
	"html/template"
	"net/http"

	"github.com/maxkaiser11/lolstats/internal/database"
	"github.com/maxkaiser11/lolstats/internal/riot"
	"github.com/maxkaiser11/lolstats/internal/web"
)

type Server struct {
	riot *riot.Client
	tmpl *template.Template
	db   *database.Queries
}

func New(riotClient *riot.Client, db *database.Queries) *Server {
	return &Server{
		riot: riotClient,
		db:   db,
		tmpl: template.Must(template.ParseFS(web.Templates, "templates/*.html")),
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /home", s.handleHome)
	mux.HandleFunc("GET /account/{name}/{tag}", s.handleRiotAccount)
	mux.HandleFunc("GET /matches/{name}/{tag}", s.handleMatchIDs)
	mux.HandleFunc("GET /signup", s.handleSignupPage)
	mux.HandleFunc("POST /signup", s.handleSignup)
	mux.HandleFunc("GET /login", s.handleLoginPage)
	mux.HandleFunc("POST /login", s.handleLogin)
	mux.Handle("GET /static/", http.FileServerFS(web.Static))
	return mux
}
