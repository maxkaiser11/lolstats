# lolstats

A Go web app for looking up League of Legends players and their recent matches via the [Riot Games API](https://developer.riotgames.com/). Anyone can search for a player; signed-in users can save their own Riot accounts for quick access.

## Stack

- **Go** standard library `net/http` (method and wildcard route patterns, `http.CrossOriginProtection`)
- **PostgreSQL** with [goose](https://github.com/pressly/goose) migrations and [sqlc](https://sqlc.dev/) for type-safe queries
- **html/template** server-rendered pages styled with Tailwind CSS
- **bcrypt** password hashing and cookie-based sessions

## Project structure

```
lolstats/
├── cmd/api/main.go               # entry point
├── internal/
│   ├── server/                   # routes, handlers, auth middleware
│   ├── riot/                     # Riot API client (requests, rate-limit retry, response structs)
│   └── database/
│       ├── schema/               # goose migrations (also read by sqlc)
│       ├── queries/              # SQL queries for sqlc
│       └── *.go                  # sqlc-generated code — don't edit by hand
├── web/                          # templates and static files
└── sqlc.yaml
```

## Requirements

- Go 1.25+
- PostgreSQL (e.g. via Docker)
- [goose](https://github.com/pressly/goose) and [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) CLIs
- A Riot API key from [developer.riotgames.com](https://developer.riotgames.com/) (development keys expire every 24 hours)

## Setup

Start Postgres:

```bash
docker run -d --name postgres18 -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:18
docker exec -it postgres18 createdb -U postgres lolstats
```

Set environment variables:

```bash
export RIOT_API_KEY=RGAPI-xxxx
export DB_URL="postgres://postgres:postgres@localhost:5432/lolstats?sslmode=disable"
```

Run migrations and generate the query code:

```bash
goose -dir internal/database/schema postgres "$DB_URL" up
sqlc generate
```

Start the server:

```bash
go run ./cmd/api
```

The app runs at `http://localhost:8080`.

## Routes

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/home` | optional | Landing page with player search |
| GET | `/account/{name}/{tag}` | — | Look up an account by Riot ID |
| GET | `/matches/{name}/{tag}` | optional | Recent matches: champion, K/D/A, win |
| GET/POST | `/signup` | — | Create an account |
| GET/POST | `/login` | — | Sign in |
| POST | `/logout` | required | Sign out and delete the session |
| GET | `/accounts` | required | Your saved Riot accounts |
| POST | `/accounts` | required | Save a Riot account |
| POST | `/accounts/{id}/delete` | required | Remove a saved account |

Riot IDs are written `name#tag`. In URLs, leave out the `#` and URL-encode spaces, e.g. `/matches/Faker/KR1`.

## Authentication

- Passwords are hashed with bcrypt.
- On login, a random 32-byte token is set as an `HttpOnly`, `SameSite=Lax` cookie. Only its SHA-256 hash is stored in the `sessions` table, so a database leak doesn't expose usable sessions.
- Sessions expire after 30 days. Logging out deletes the session row.
- Cross-site POSTs are blocked by `http.CrossOriginProtection`; all state-changing routes are POST-only.

## Database

| Table | Purpose |
|---|---|
| `users` | Email and password hash |
| `sessions` | Hashed session tokens with expiry, one row per login |
| `riot_accounts` | Riot accounts saved by a user, keyed by PUUID |

Saved accounts store the Riot **PUUID** as well as the name and tag, since Riot IDs can change but PUUIDs can't.

Schema changes always go in a new migration file in `internal/database/schema/`. Run `sqlc generate` after changing the schema or queries.

## Notes

- Uses the `europe` regional routing host (account-v1 and match-v5).
- Requests that hit the rate limit (HTTP 429) are retried after Riot's `Retry-After` delay.
- Each `/matches` request makes 2 + N Riot calls, so keep the match count low on a development key.

---

lolstats isn't endorsed by Riot Games and doesn't reflect the views or opinions of Riot Games or anyone officially involved in producing or managing Riot Games properties.
