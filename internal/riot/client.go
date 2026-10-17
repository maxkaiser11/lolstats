package riot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	key    string
	region string
	http   *http.Client
}

type Account struct {
	PUUID    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

// https://europe.api.riotgames.com/lol/match/v5/matches/{matchId}
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
		GameDuration int `json:"gameDuration"`
		QueueID int `json:"queueID"`
		Participants []Participant `json:"participants"`
	} `json:"info"`
}

func NewClient(key string) *Client {
	return &Client{key: key, region: "europe", http: &http.Client{Timeout: 10 * time.Second}}
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

func (c *Client) MatchIDSByPUUID(puuid string, count int) ([]string, error) {
	u := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/by-puuid/%s/ids?start=0&count=%d",
		c.region, puuid, count)

	var matchIds []string
	if err := c.get(u, &matchIds); err != nil {
		return nil, err
	}
	return matchIds, nil
}


func (c *Client) MatchByID(id string) (*Match, error) {
	u := fmt.Sprintf("https://%s.api.riotgames.com/lol/match/v5/matches/%s", c.region, id)

	var m Match

	if err := c.get(u, &m); err != nil {
		return nil, err
	}
	return &m, nil

}
