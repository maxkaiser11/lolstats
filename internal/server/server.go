package server

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/maxkaiser11/lolstats/internal/database"
	"github.com/maxkaiser11/lolstats/internal/riot"
	"github.com/maxkaiser11/lolstats/internal/web"
)

type RiotAPI interface {
	AccountByRiotID(region, name, tag string) (*riot.Account, error)
	MatchIDSByPUUID(puuid string, count int) ([]string, error)
	MatchByID(id string) (*riot.Match, error)
}

var _ RiotAPI = (*riot.Client)(nil)

type Server struct {
	riot   RiotAPI
	champs *riot.Champions
	tmpl   *template.Template
	db     *database.Store
}

func New(riotClient RiotAPI, db *database.Store) *Server {
	s := &Server{riot: riotClient, db: db}
	funcs := template.FuncMap{
		"duration": func(sec int32) string {
			return fmt.Sprintf("%d:%02d", sec/60, sec%60)
		},
		"winrate": func(wins, games int64) int64 {
			return wins * 100 / games
		},
		"champ":     func(id string) string { return s.champs.Name(id) },
		"champIcon": func(id string) string { return s.champs.IconURL(id) },
	}
	s.tmpl = template.Must(template.New("").Funcs(funcs).ParseFS(web.Templates, "templates/*.html"))
	return s
}

func (s *Server) SetChampions(c *riot.Champions) {
	s.champs = c
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.optionalAuth(s.handleHome))
	mux.HandleFunc("GET /account/{name}/{tag}", s.handleRiotAccount)
	mux.HandleFunc("GET /matches/{name}/{tag}", s.optionalAuth(s.handleMatches))
	mux.HandleFunc("GET /signup", s.handleSignupPage)
	mux.HandleFunc("POST /signup", s.handleSignup)
	mux.HandleFunc("GET /login", s.handleLoginPage)
	mux.HandleFunc("POST /login", s.handleLogin)
	mux.HandleFunc("POST /logout", s.requireAuth(s.handleLogout))
	mux.HandleFunc("GET /accounts", s.requireAuth(s.handleAccounts))
	mux.HandleFunc("POST /accounts", s.requireAuth(s.handleAddAccount))
	mux.HandleFunc("POST /accounts/{id}/delete", s.requireAuth(s.handleDeleteAccount))

	mux.Handle("GET /static/", http.FileServerFS(web.Static))

	handler := http.NewCrossOriginProtection().Handler(mux)
	return handler
}
