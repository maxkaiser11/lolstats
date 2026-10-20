package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)


func TestAuthRequired(t *testing.T) {
	s := New(&fakeRiot{}, nil)

	tests := []struct {
		method, path string
	}{
		{"GET", "/accounts"},
		{"POST", "/accounts"},
		{"POST", "/logout"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			s.Routes().ServeHTTP(rec, req)

			if rec.Code != http.StatusSeeOther {
				t.Fatalf("status = %d, want 303", rec.Code)
			}
			if loc := rec.Header().Get("Location"); loc != "/login" {
				t.Errorf("redirect = %q, want /login", loc)
			}
		})
	}
}

func TestCrossOriginPostBlocked(t *testing.T) {
	s := New(&fakeRiot{}, nil)

	req := httptest.NewRequest("POST", "/login", nil)
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	rec := httptest.NewRecorder()
	s.Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}
