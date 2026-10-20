# lolstats

A Go web app for looking up League of Legends players and their recent matches via the [Riot Games API](https://developer.riotgames.com/). Anyone can search for a player; signed-in users can save their own Riot accounts for quick access.

## Stack

- **Go** standard library `net/http` (method and wildcard route patterns, `http.CrossOriginProtection`)
- **PostgreSQL** with [goose](https://github.com/pressly/goose) migrations (embedded and run on startup) and [sqlc](https://sqlc.dev/) for type-safe queries
- **html/template** server-rendered pages styled with Tailwind CSS
- **bcrypt** password hashing and cookie-based sessions (only a SHA-256 hash of each session token is stored)
- **Docker** multi-stage build and Docker Compose for the app and database
- **Tests** with `net/http/httptest`, a fake Riot client, and a real Postgres test database

## Quick start

Requires Docker and a Riot API key from [developer.riotgames.com](https://developer.riotgames.com/). Development keys expire every 24 hours.

```bash
echo "RIOT_API_KEY=RGAPI-xxxx" > .env
docker compose up --build
```

The app runs at `http://localhost:8080`. Database migrations run automatically when the app starts.

| Command | What it does |
|---|---|
| `docker compose up --build -d` | Build and start in the background |
| `docker compose logs -f app` | Follow the app's logs |
| `docker compose down` | Stop the containers (data is kept) |
| `docker compose down -v` | Stop and delete the database volume |

## Local development

To run the app with `go run` instead, start only the database in Docker:

```bash
docker compose up -d db
```

Then add the database URL to `.env`:

```
RIOT_API_KEY=RGAPI-xxxx
DB_URL=postgres://postgres:postgres@localhost:5432/lolstats?sslmode=disable
```

and run:

```bash
go run ./cmd/api
```

Requires Go 1.27+. If you change anything in `internal/database/queries/` or `internal/database/schema/`, regenerate the query code with [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html):

```bash
sqlc generate
```

## Running tests

Create a separate test database once (the tests empty its tables on every run):

```bash
docker compose up -d db
docker compose exec db createdb -U postgres lolstats_test
```

Add it to `.env`:

```
TEST_DB_URL=postgres://postgres:postgres@localhost:5432/lolstats_test?sslmode=disable
```

Then run:

```bash
go test ./...
```

Tests that need the database are skipped if `TEST_DB_URL` isn't set, so the rest still run without Docker. The suite covers:

- the matches page, using a fake Riot client instead of the real API
- auth redirects on protected routes and blocking of cross-origin POSTs
- signup, login and logout, including that logout invalidates the session server-side
- that failed logins give the same error for wrong passwords and unknown emails
- that an invalid Riot ID saves nothing, and that users can't delete each other's saved accounts

## Project structure

```
lolstats/
├── cmd/api/main.go               # entry point: config, DB connection, migrations
├── internal/
│   ├── server/                   # routes, handlers, auth middleware, tests
│   ├── riot/                     # Riot API client (requests, rate-limit retry, response structs)
│   ├── database/
│   │   ├── schema/               # goose migrations (also read by sqlc)
│   │   ├── queries/              # SQL queries for sqlc
│   │   ├── migrate.go            # embeds and runs the migrations
│   │   └── *.go                  # sqlc-generated code — don't edit by hand
│   └── web/                      # embedded templates and static files
├── Dockerfile
├── docker-compose.yaml
└── sqlc.yaml
```

## Routes

| Method | Path | Auth | Description |
|---|---|---|---|
| GET | `/` | optional | Landing page with player search |
| GET | `/account/{name}/{tag}` | — | Look up an account by Riot ID |
| GET | `/matches/{name}/{tag}` | optional | Recent matches: champion, K/D/A, result, duration |
| GET/POST | `/signup` | — | Create an account |
| GET/POST | `/login` | — | Sign in |
| POST | `/logout` | required | Sign out and delete the session |
| GET | `/accounts` | required | Your saved Riot accounts |
| POST | `/accounts` | required | Save a Riot account |
| POST | `/accounts/{id}/delete` | required | Remove a saved account |

Riot IDs are written `name#tag`. In URLs, leave out the `#` and URL-encode spaces, e.g. `/matches/Faker/KR1`.

---

lolstats isn't endorsed by Riot Games and doesn't reflect the views or opinions of Riot Games or anyone officially involved in producing or managing Riot Games properties.
