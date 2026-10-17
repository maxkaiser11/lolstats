# lolstats

A small Go HTTP server that wraps the [Riot Games API](https://developer.riotgames.com/) to look up League of Legends players and their recent matches.

## Project structure

```
riotapp/
├── go.mod
├── cmd/server/main.go       # HTTP server and handlers
└── internal/riot/client.go  # Riot API client (requests, rate-limit retry, response structs)
```

## Requirements

- Go 1.22+ (uses method and wildcard route patterns)
- A Riot API key from [developer.riotgames.com](https://developer.riotgames.com/) (development keys expire every 24 hours)

## Running

```bash
export RIOT_API_KEY=RGAPI-xxxx
go run ./cmd/server
```

The server listens on `http://localhost:8080`.

## Endpoints

| Method | Path | Description |
|---|---|---|
| GET | `/account/{name}/{tag}` | Look up an account by Riot ID |
| GET | `/matches/{name}/{tag}` | Recent matches for a player: champion, K/D/A, win |

Riot IDs are written `name#tag`. Leave out the `#` and URL-encode spaces:

```bash
curl "http://localhost:8080/account/name/tag"
curl "http://localhost:8080/matches/name/tag"
```

## Notes

- Uses the `europe` regional routing host (account-v1 and match-v5).
- Requests that hit the rate limit (HTTP 429) are retried after Riot's `Retry-After` delay.
- Each `/matches` request makes 2 + N Riot calls, so keep the match count low on a development key.
