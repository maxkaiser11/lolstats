package riot

import "testing"

func TestMatchRegion(t *testing.T) {
	tests := map[string]string{
		"euw1":    "europe",
		"EUN1":    "europe",
		"tr1":     "europe",
		"kr":      "asia",
		"jp1":     "asia",
		"na1":     "americas",
		"br1":     "americas",
		"la2":     "americas",
		"oc1":     "sea",
		"sg2":     "sea",
		"unknown": DefaultMatchRegion,
		"":        DefaultMatchRegion,
	}
	for platform, want := range tests {
		if got := MatchRegion(platform); got != want {
			t.Errorf("MatchRegion(%q) = %q, want %q", platform, got, want)
		}
	}
}
