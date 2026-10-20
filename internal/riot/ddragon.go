package riot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const ddragonURL = "https://ddragon.leagueoflegends.com"

type champion struct {
	ID   string // Riot's internal ID, e.g. "MonkeyKing"
	Name string // display name, e.g. "Wukong"
}

// Champions holds static champion data from Data Dragon, Riot's public CDN.
// It needs no API key.
type Champions struct {
	Version string
	byID    map[string]champion // keyed by lowercase ID
}

// LoadChampions downloads the champion list for the latest game version.
func LoadChampions(ctx context.Context) (*Champions, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	var versions []string
	if err := getJSON(ctx, client, ddragonURL+"/api/versions.json", &versions); err != nil {
		return nil, fmt.Errorf("versions: %w", err)
	}
	if len(versions) == 0 {
		return nil, errors.New("no versions returned")
	}

	var file struct {
		Data map[string]struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	url := fmt.Sprintf("%s/cdn/%s/data/en_US/champion.json", ddragonURL, versions[0])
	if err := getJSON(ctx, client, url, &file); err != nil {
		return nil, fmt.Errorf("champions: %w", err)
	}

	c := &Champions{Version: versions[0], byID: make(map[string]champion, len(file.Data))}
	for _, ch := range file.Data {
		c.byID[strings.ToLower(ch.ID)] = champion{ID: ch.ID, Name: ch.Name}
	}
	return c, nil
}

// Name returns the display name for a champion ID, or the ID itself if unknown.
func (c *Champions) Name(id string) string {
	if c != nil {
		if ch, ok := c.byID[strings.ToLower(id)]; ok {
			return ch.Name
		}
	}
	return id
}

// IconURL returns the champion's square icon, or "" if unknown.
func (c *Champions) IconURL(id string) string {
	if c == nil {
		return ""
	}
	ch, ok := c.byID[strings.ToLower(id)]
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s/cdn/%s/img/champion/%s.png", ddragonURL, c.Version, ch.ID)
}

func getJSON(ctx context.Context, client *http.Client, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", url, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
