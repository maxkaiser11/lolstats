package riot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// AccountRegion is the routing value used for Riot account lookups. The
// account-v1 API serves every player from any regional cluster.
const AccountRegion = "europe"

// DefaultMatchRegion is used for match-v5 calls when a player's home region
// can't be determined.
const DefaultMatchRegion = "europe"

type Client struct {
	key  string
	http *http.Client
}

type Account struct {
	PUUID    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

type Participant struct {
	PUUID        string `json:"puuid"`
	ChampionName string `json:"championName"`
	Kills        int    `json:"kills"`
	Deaths       int    `json:"deaths"`
	Assists      int    `json:"assists"`
	Win          bool   `json:"win"`
}

type Match struct {
	Metadata struct {
		MatchID string `json:"matchId"`
	} `json:"metadata"`
	Info struct {
		GameCreation int64         `json:"gameCreation"`
		GameDuration int           `json:"gameDuration"`
		QueueID      int           `json:"queueId"`
		Participants []Participant `json:"participants"`
	} `json:"info"`
}

func NewClient(key string) *Client {
	return &Client{key: key, http: &http.Client{Timeout: 10 * time.Second}}
}

// platformToRegion maps a League platform (the server a player plays on) to
// the regional cluster that serves its match-v5 data.
var platformToRegion = map[string]string{
	"na1": "americas", "br1": "americas", "la1": "americas", "la2": "americas",
	"kr": "asia", "jp1": "asia",
	"euw1": "europe", "eun1": "europe", "tr1": "europe", "ru": "europe", "me1": "europe",
	"oc1": "sea", "sg2": "sea", "tw2": "sea", "vn2": "sea", "ph2": "sea", "th2": "sea",
}

// MatchRegion returns the match-v5 routing cluster for a platform such as
// "euw1" or "kr". Unknown platforms fall back to DefaultMatchRegion.
func MatchRegion(platform string) string {
	if r, ok := platformToRegion[strings.ToLower(platform)]; ok {
		return r
	}
	return DefaultMatchRegion
}

func (c *Client) get(u string, out any) error {
	for {
		req, _ := http.NewRequest("GET", u, nil)
		req.Header.Set("X-Riot-Token", c.key)

		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			wait, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
			resp.Body.Close()
			time.Sleep(time.Duration(wait+1) * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("riot api %s: %s", u, resp.Status)
		}
		return json.NewDecoder(resp.Body).Decode(out)
	}
}

func (c *Client) AccountByRiotID(region, name, tag string) (*Account, error) {
	u := fmt.Sprintf("https://%s.api.riotgames.com/riot/account/v1/accounts/by-riot-id/%s/%s",
		region, url.PathEscape(name), url.PathEscape(tag))

	var account Account
	return &account, c.get(u, &account)
}

// MatchRegionByPUUID looks up which platform the player plays League on and
// returns the regional cluster to use for their match history.
func (c *Client) MatchRegionByPUUID(puuid string) (string, error) {
	u := fmt.Sprintf("https://%s.api.riotgames.com/riot/account/v1/region/by-game/lol/by-puuid/%s",
		AccountRegion, url.PathEscape(puuid))

	var res struct {
		Region string `json:"region"` // a platform, e.g. "euw1" or "kr"
	}
	if err := c.get(u, &res); err != nil {
		return "", err
	}
	return MatchRegion(res.Region), nil
}

func (c *Client) MatchIDSByPUUID(region, puuid string, count int) ([]string, error) {
	u := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/by-puuid/%s/ids?start=0&count=%d",
		region, url.PathEscape(puuid), count)

	var matchIds []string
	if err := c.get(u, &matchIds); err != nil {
		return nil, err
	}
	return matchIds, nil
}

func (c *Client) MatchByID(region, id string) (*Match, error) {
	u := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/%s", region, url.PathEscape(id))

	var m Match

	if err := c.get(u, &m); err != nil {
		return nil, err
	}
	return &m, nil
}
