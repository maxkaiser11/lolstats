package server

import (
	"net/http"
	"testing"
)

func TestHomeAndUnknownRoutes(t *testing.T) {
	// No session cookie and no Riot calls, so no database is needed.
	h := New(&fakeRiot{}, nil).Routes()

	if rec := get(h, "/", nil); rec.Code != http.StatusOK {
		t.Errorf("GET / = %d, want 200", rec.Code)
	}
	if rec := get(h, "/does-not-exist", nil); rec.Code != http.StatusNotFound {
		t.Errorf("GET /does-not-exist = %d, want 404", rec.Code)
	}

	rec := get(h, "/home", nil)
	if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != "/" {
		t.Errorf("GET /home = %d to %q, want 301 to /", rec.Code, rec.Header().Get("Location"))
	}
}
