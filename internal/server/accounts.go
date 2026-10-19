package server

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/maxkaiser11/lolstats/internal/database"
)

type accountsPage struct {
	Title    string
	User     *database.User
	Accounts []database.RiotAccount
	Error    string
}


func (s *Server) handleAccounts(w http.ResponseWriter, r *http.Request) {
	user, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	accounts, err := s.db.ListRiotAccountsByUser(r.Context(), user.ID)
	if err != nil {
		log.Printf("list accounts: %v", err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	page := accountsPage{Title: "My accounts", User: &user, Accounts: accounts}
	if err := s.tmpl.ExecuteTemplate(w, "accounts.html", page); err != nil {
		log.Printf("accounts template: %v", err)
	}
}

func (s *Server) handleAddAccount(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r)
	name := strings.TrimSpace(r.FormValue("name"))
	tag := strings.TrimPrefix(strings.TrimSpace(r.FormValue("tag")), "#")


	acc, err := s.riot.AccountByRiotID("europe", name, tag)
	if err != nil {
		http.Redirect(w, r, "/accounts", http.StatusSeeOther)
	}

	_, err = s.db.AddRiotAccount(context.Background(), database.AddRiotAccountParams{
		UserID: user.ID,
		Puuid: acc.PUUID,
		GameName: acc.GameName,
		TagLine: acc.TagLine,
	})

	if err != nil {
		log.Printf("add account: %v", err)
	}

	http.Redirect(w, r, "/accounts", http.StatusSeeOther)
}

func (s *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	user, _ := currentUser(r)
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if _, err := s.db.DeleteRiotAccount(r.Context(), database.DeleteRiotAccountParams{
		ID: id, UserID: user.ID,
	}); err != nil {
		log.Printf("delete account: %v", err)
	}
	http.Redirect(w, r, "/accounts", http.StatusSeeOther)
}
