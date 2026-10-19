package server

import "net/http"

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Title string
	}{
		Title: "Home",
	}

	if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
