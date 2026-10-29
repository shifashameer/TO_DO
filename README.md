# TODO List Web App

A Go web app using chi v5, PostgreSQL, GORM, HTML templates, bcrypt, and session authentication. It supports creating, editing, completing, deleting, filtering, and sorting todos. Authenticated users share one todo list; there is no public signup or per-user ownership model.

## Prerequisites

- Go 1.26.4 or newer compatible toolchain (`go version`).
- Docker Desktop with Docker Compose, running (`open -a Docker`).
- A Mac terminal using zsh for the commands below.

## Project structure

```text
main.go                    HTTP server and protected/public routes
cmd/seed/main.go           Initial-user command
internal/models/          Todo and User GORM models
internal/store/           Store interfaces, memory and PostgreSQL implementations
internal/auth/            bcrypt and expiring server-side sessions
internal/handlers/        Form/JSON handlers, rendering, authentication middleware
templates/                Embedded HTML pages
static/                   Embedded stylesheet
docker-compose.yml        Persistent development DB and isolated test DB
.env.example              Configuration example (no working credentials)
docs/audit.md             Initial/final audit and verification evidence
```

## Configure this workspace

The audit created `/Users/shifashameer/todo/.env` with mode 600 using the existing PostgreSQL container settings. It is ignored by Git. The existing `postgres_data` volume was retained. Do not replace its database password with a new random value: PostgreSQL initialization variables do not change passwords in an already initialized volume.

Load the existing configuration in **each new terminal**:

```zsh
cd /Users/shifashameer/todo
set -a
source .env
set +a
export PGPASSWORD="$POSTGRES_PASSWORD"
```

For a fresh checkout without `.env`, copy `.env.example` to `.env`, set its database password, and restrict permissions with `chmod 600 .env`. Use the existing database password for an existing volume. For a new empty database, choose a strong unique password. The app does not automatically load `.env`; the shell commands above do. Compose reads `.env` automatically.

| Variable | Meaning |
|---|---|
| `DATABASE_URL` | PostgreSQL DSN; example uses `127.0.0.1:5432`, `todo_user`, `todo_db`, and a five-second connection timeout |
| `POSTGRES_USER`, `POSTGRES_DB`, `POSTGRES_PORT` | Compose development database settings |
| `POSTGRES_PASSWORD` | Database initialization password, read from private configuration |
| `PGPASSWORD` | Password read by the Go PostgreSQL driver; export from `POSTGRES_PASSWORD` |
| `TEST_DATABASE_URL` | Dedicated `todo_test` database at `127.0.0.1:5433` with `todo_test_user` |
| `ADDR` | Server listen address, default `127.0.0.1:8080` |
| `COOKIE_SECURE` | `false` for local HTTP; set `true` when serving through HTTPS |
| `SEED_USERNAME`, `SEED_PASSWORD` | Only needed by the seed command; password must be 12–72 bytes |

If changing database identifiers or ports, update both the Compose variables and `DATABASE_URL`. For remote PostgreSQL, configure TLS appropriately instead of the local `sslmode=disable` setting. Passwords and hashes are not logged or returned in user JSON.

## Start PostgreSQL, create your login, and run

After loading `.env` as above:

```zsh
open -a Docker
docker compose up -d --wait postgres

read 'SEED_USERNAME?Initial username: '
read -rs 'SEED_PASSWORD?Initial password (12–72 bytes): '
printf '\n'
export SEED_USERNAME SEED_PASSWORD
go run ./cmd/seed
unset SEED_PASSWORD SEED_USERNAME

go run .
```

The seed command creates a user with a bcrypt password hash. It never prints the password and refuses to overwrite an existing username. Run it once for each intended account; if you already have a login, skip seeding. There is no built-in password reset command.

Both the server and seed command run checked GORM migrations for `Todo` and `User` automatically. No separate migration command is needed. A missing database configuration or failed migration stops startup. The production server always constructs the PostgreSQL store; memory storage is used in unit tests only.

In another terminal, open the app:

```zsh
open http://127.0.0.1:8080
```

Log in using the username and password you entered at the prompts. There is no default login password.

## Run tests in a separate terminal

```zsh
cd /Users/shifashameer/todo
set -a
source .env
set +a
export PGPASSWORD="$POSTGRES_PASSWORD"

docker compose --profile test up -d --wait postgres-test
go test ./... -count=1
go vet ./...
go test -race ./... -count=1
```

The test service uses a separate PostgreSQL container on port 5433 and an ephemeral filesystem. Integration tests require `TEST_DATABASE_URL` with a database name ending in `_test`. Each test creates a unique schema inside a transaction and rolls it back. Tests never use `DATABASE_URL` as a fallback. If the configured test database cannot connect, the suite fails. Without `TEST_DATABASE_URL`, database tests explicitly skip; inspect verbose output rather than treating skips as database verification:

```zsh
env -u TEST_DATABASE_URL go test ./... -v -count=1
```

The audit verified Go 1.26.4 on macOS arm64, including `go vet` and the race detector. Inside the Codex sandbox, `GOCACHE=/tmp/todo-go-cache` was needed because the default cache became inaccessible; normal terminal use should not require this override.

## Routes and behavior

| Method | Path | Behavior |
|---|---|---|
| GET | `/healthz` | Public liveness endpoint: `200`, exact body `ok` |
| GET / POST | `/login` | Public login page / form authentication |
| POST | `/logout` | Revoke session and clear cookie; GET is not allowed |
| GET | `/static/style.css` | Public embedded stylesheet |
| GET | `/` | Protected HTML list and add form |
| POST | `/todos` | Create from `title` and optional `description`; redirect to `/` |
| GET | `/todos/{id}/edit` | Protected edit form |
| POST | `/todos/{id}` | Update title/description; redirect to `/` |
| GET | `/api/todos` | Protected JSON array (including `[]` for empty results) |
| PATCH | `/api/todos/{id}/toggle` | Toggle completed state; `200` |
| DELETE | `/api/todos/{id}` | Delete; `204` |

Both list routes accept `status=all|active|completed`, `sort=created_at|title`, and `order=asc|desc`, including combined queries. Defaults are `all`, `created_at`, `asc`. Unsupported values return `400`; SQL sort fields are allowlisted. The HTML controls submit all three options together.

Creation/edit/login accept URL-encoded forms, not JSON. Unsupported body media types return `415`; malformed/oversized forms and blank/whitespace-only todo titles return `400`. Missing todos return `404`. Storage failures return generic `500` responses without database details. Unauthenticated browsers redirect to `/login`; unauthenticated API requests return `401`.

Sessions use cryptographically random 256-bit opaque tokens stored only on the server with a 12-hour lifetime. Cookies are HttpOnly, SameSite=Strict, and optionally Secure. Login rotates an existing token; logout revokes it. Restarting the app invalidates all sessions. A multi-instance deployment would need a shared session store. Go's cross-origin protection rejects unsafe cross-origin browser requests, including login/logout. Use HTTPS and `COOKIE_SECURE=true` beyond local development.

## Stop safely

Press **Ctrl+C** in the terminal running `go run .`. Then:

```zsh
cd /Users/shifashameer/todo
docker compose stop postgres
docker compose --profile test stop postgres-test
```

`stop` retains the development volume. Do not use `docker compose down -v` unless you deliberately want to delete development data. Stopping the ephemeral test service clears its test database; it will initialize again on startup.

## Manual browser checklist

- Visit `/healthz`; confirm `ok`. Visit `/` while logged out; confirm login redirect.
- Try an incorrect password, then log in with your seeded credentials.
- Create a todo with only a title, then one with a description. Reject a blank title.
- Mark a todo complete, then incomplete; edit both its title and description.
- Combine status, sort field, and order; confirm the displayed list changes correctly.
- Delete a todo; refresh and confirm it stays deleted.
- Log out; confirm `/` redirects and `/api/todos` returns `401`.
- Restart the app and log in again; confirm remaining todos persist.

Automated router tests and live HTTP smoke tests cover these core flows. A visual browser inspection and manual interaction checklist were not performed during the audit.
