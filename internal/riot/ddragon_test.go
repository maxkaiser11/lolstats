package riot

import "testing"

func TestChampions(t *testing.T) {
	c := &Champions{
		Version: "15.1.1",
		byID: map[string]champion{
			"monkeyking":   {ID: "MonkeyKing", Name: "Wukong"},
			"fiddlesticks": {ID: "Fiddlesticks", Name: "Fiddlesticks"},
		},
	}

	tests := []struct {
		id, wantName, wantIcon string
	}{
		{"MonkeyKing", "Wukong", "https://ddragon.leagueoflegends.com/cdn/15.1.1/img/champion/MonkeyKing.png"},
		{"FiddleSticks", "Fiddlesticks", "https://ddragon.leagueoflegends.com/cdn/15.1.1/img/champion/Fiddlesticks.png"},
		{"BrandNewChamp", "BrandNewChamp", ""}, // unknown: raw ID, no icon
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := c.Name(tt.id); got != tt.wantName {
				t.Errorf("Name = %q, want %q", got, tt.wantName)
			}
			if got := c.IconURL(tt.id); got != tt.wantIcon {
				t.Errorf("IconURL = %q, want %q", got, tt.wantIcon)
			}
		})
	}

	var none *Champions // not loaded
	if got := none.Name("Ahri"); got != "Ahri" {
		t.Errorf("nil Champions: Name = %q, want Ahri", got)
	}
}
