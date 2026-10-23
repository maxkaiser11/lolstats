# lolstats

A Go + Postgres web app for looking up League of Legends players, with match caching, per-champion stats, user accounts and a tested codebase.

**Live demo:** https://lolstats-f20d.onrender.com
*(Hosted on Render's free tier, so the first load after a quiet period can take up to a minute while the server wakes up.)*

![lolstats player page](docs/screenshot.png)

## What it does

Search any player by Riot ID (`name#tag`) to see their 20 most recent matches and per-champion stats: games played, win rate and average K/D/A. The Riot API doesn't provide these aggregates directly, so lolstats stores every match it fetches in Postgres and computes them with SQL across all stored games. Signed-in users can save their own Riot accounts for quick access.

## Highlights

- **Fetches only what's missing.** A repeat visit usually costs 1 Riot API call instead of 21, because matches already in the database are never re-fetched.
- **No half-saved data.** Each match is written with all 10 participants in a single transaction.
- **Degrades gracefully.** If Riot fails part-way through, the page still shows what's already stored. If Data Dragon is unreachable, the app still starts and shows raw champion IDs.
- **Security-minded auth.** bcrypt password hashing, cookie sessions where only a SHA-256 hash of each token is stored, cross-origin POST protection, and identical errors for wrong passwords and unknown emails.
- **Testable by design.** The Riot API sits behind a Go interface, so tests run against a fake client and a real Postgres test database.

## How it works

When a player's page is opened:

1. The app asks Riot for the player's 20 most recent match IDs (one API call).
2. One query checks which of those matches are already in the database.
3. Only the missing matches are fetched from Riot and saved, each in its own transaction.
4. The page is built from the database: recent games, plus champion stats aggregated with SQL across every stored game.

Champion display names and icons come from [Data Dragon](https://developer.riotgames.com/docs/lol#data-dragon), Riot's static data CDN. The champion list is loaded once at startup, and icons load in the browser directly from Riot's CDN.

## Stack

- **Go** standard library `net/http` (method and wildcard route patterns, `http.CrossOriginProtection`)
- **PostgreSQL** with [goose](https://github.com/pressly/goose) migrations (embedded and run on startup), [sqlc](https://sqlc.dev/) for type-safe queries, and transactions for multi-row writes
- **html/template** server-rendered pages styled with Tailwind CSS
- **bcrypt** password hashing and cookie-based sessions
- **Docker** multi-stage build and Docker Compose for the app and database
- **Riot API** client with rate-limit retry, plus **Data Dragon** for champion data
- **Tests** with `net/http/httptest`, a fake Riot client and a real Postgres test database
- **Deployed** on [Render](https://render.com/)

## Quick start

Requires Docker and a Riot API key from [developer.riotgames.com](https://developer.riotgames.com/). Development keys expire every 24 hours.

```
echo "RIOT_API_KEY=RGAPI-xxxx" > .env
docker compose up --build
```

The app runs at `http://localhost:8080`. Database migrations run automatically when the app starts.

| Command                        | What it does                        |
|--------------------------------|-------------------------------------|
| `docker compose up --build -d` | Build and start in the background   |
| `docker compose logs -f app`   | Follow the app's logs               |
| `docker compose down`          | Stop the containers (data is kept)  |
| `docker compose down -v`       | Stop and delete the database volume |

## Local development

To run the app with `go run` instead, start only the database in Docker:

```
docker compose up -d db
```

Then add the database URL to `.env`:

```
RIOT_API_KEY=RGAPI-xxxx
DB_URL=postgres://postgres:postgres@localhost:5432/lolstats?sslmode=disable
```

and run:

```
go run ./cmd/api
```

Requires Go 1.27+. If you change anything in `internal/database/queries/` or `internal/database/schema/`, regenerate the query code with [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html):

```
sqlc generate
```

## Running tests

Create a separate test database once (the tests empty its tables on every run):

```
docker compose up -d db
docker compose exec db createdb -U postgres lolstats_test
```

Add it to `.env`:

```
TEST_DB_URL=postgres://postgres:postgres@localhost:5432/lolstats_test?sslmode=disable
```

Then run:

```
go test ./...
```

Tests that need the database are skipped if `TEST_DB_URL` isn't set, so the rest still run without Docker. The suite covers:

- the matches page and champion stats, using a fake Riot client instead of the real API
- that a second visit serves stored matches without fetching them from Riot again
- champion name and icon lookups, including unknown champions and missing data
- auth redirects on protected routes and blocking of cross-origin POSTs
- signup, login and logout, including that logout invalidates the session server-side
- that failed logins give the same error for wrong passwords and unknown emails
- that an invalid Riot ID saves nothing, and that users can't delete each other's saved accounts

## Deployment

The live demo runs on Render as a Docker web service with a managed Postgres database. Configuration comes from environment variables (`RIOT_API_KEY`, `DB_URL`), and migrations run automatically on startup, so a deploy is just a push to `main`.

## Project structure

```
lolstats/
├── cmd/api/main.go               # entry point: config, DB connection, migrations, champion data
├── internal/
│   ├── server/                   # routes, handlers, auth middleware, match syncing, tests
│   ├── riot/
│   │   ├── client.go             # Riot API client (requests, rate-limit retry, response structs)
│   │   └── ddragon.go            # Data Dragon champion names and icon URLs
│   ├── database/
│   │   ├── schema/               # goose migrations (also read by sqlc)
│   │   ├── queries/              # SQL queries for sqlc
│   │   ├── migrate.go            # embeds and runs the migrations
│   │   ├── store.go              # sqlc queries plus transactional writes
│   │   └── *.go                  # sqlc-generated code — don't edit by hand
│   └── web/                      # embedded templates and static files
├── Dockerfile
├── docker-compose.yaml
└── sqlc.yaml
```

## Routes

| Method   | Path                    | Auth     | Description                                   |
|----------|-------------------------|----------|-----------------------------------------------|
| GET      | `/`                     | optional | Landing page with player search               |
| GET      | `/account/{name}/{tag}` | —        | Look up an account by Riot ID                 |
| GET      | `/matches/{name}/{tag}` | optional | Champion stats and the 20 most recent matches |
| GET/POST | `/signup`               | —        | Create an account                             |
| GET/POST | `/login`                | —        | Sign in                                       |
| POST     | `/logout`               | required | Sign out and delete the session               |
| GET      | `/accounts`             | required | Your saved Riot accounts                      |
| POST     | `/accounts`             | required | Save a Riot account                           |
| POST     | `/accounts/{id}/delete` | required | Remove a saved account                        |

Riot IDs are written `name#tag`. In URLs, leave out the `#` and URL-encode spaces, e.g. `/matches/Faker/KR1`.

---

lolstats isn't endorsed by Riot Games and doesn't reflect the views or opinions of Riot Games or anyone officially involved in producing or managing Riot Games properties.
