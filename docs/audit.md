# TODO project audit

Initial inspection: 2026-10-19. Root: `/Users/shifashameer/todo`.
No Git repository was present, so no Git diff/status was available. Existing source, tests, templates, CSS, Compose, module and seed command were inspected.

| Requirement | Initial status | Evidence |
|---|---|---|
| 1 Environment/health | PARTIAL | chi v5 and health route in source; no health test or verified startup |
| 2 Model/memory store | PARTIAL | Model/interface/mutex exist; missing store and JSON handler tests |
| 3 HTML list | PASS | Existing handler test passed (uses working-directory mutation) |
| 4 Create form | PARTIAL | Happy-path test passes; whitespace/malformed forms unchecked |
| 5 Toggle/edit/delete | PARTIAL | Happy paths pass; error and route coverage missing |
| 6 PostgreSQL | FAIL | Baseline suite cannot connect; hardcoded credentials, unchecked migration, destructive tests |
| 7 Filter/sort | PARTIAL | Implementation present; missing validation and combined UI controls |
| 8 Polish/docs | FAIL | Styles exist; README and comprehensive tests absent |
| Authentication | FAIL | Password hashing tests pass; cookie is forgeable user ID, no expiry or server revocation |

Baseline `go test ./...`: auth and handlers passed; store panicked on sandbox-denied PostgreSQL connection. No successful database test was observed. Existing tests containing `DELETE FROM todos` will be isolated before any database-enabled run.

## Final verified audit

| Requirement | Final status | Evidence |
|---|---|---|
| 1 Environment/health | PASS | Go 1.26.4; chi v5; router test checks exact `ok`; actual servers started on 8080 and 18080, live health 200 |
| 2 Model/memory store | PASS | Required model fields and Store interface retained; CRUD/error contract and concurrent access tests pass under race detector |
| 3 HTML list | PASS | Embedded html/template rendering; list test checks title, description, escaping and stylesheet reference; live page available |
| 4 Create form | PASS | Saved data/303 covered; missing/whitespace titles, malformed and oversized forms rejected; live PostgreSQL creation passed |
| 5 Toggle/edit/delete | PASS | Real chi route workflow tests both toggle directions, edit/update/delete, missing IDs and database failures; live HTTP smoke passed |
| 6 PostgreSQL/GORM | PASS | Production uses PostgreSQL only; checked migrations, environment configuration, exact missing-row semantics; full isolated DB contract passed; existing development service started without resetting its volume |
| 7 Filter/sort | PASS | Shared contract tests every status × field × direction; combined JSON response verified; invalid parameters rejected; UI submits combined choices; SQL fields allowlisted |
| 8 Polish/tests/docs | PASS | gofmt clean; full tests/vet/race passed with PostgreSQL enabled; embedded CSS returns 200; README includes configuration, run/seed/test/stop instructions and manual checklist |
| Authentication | PASS | bcrypt salt/check tests; login success/failure and cookie flags; random tokens, expiry, rotation, invalid/forged cookie rejection, API 401, route protection and logout revocation; public health/login verified; live seeded login passed |

PASS refers to the evidence stated above. Visual browser inspection and clicking the browser UI were not performed. HTML, stylesheet delivery, underlying HTTP actions, and template escaping were verified automatically. No public deployment/HTTPS proxy was tested. Secure-cookie behavior was verified in handler tests.

## Changes

- `main.go`, `main_test.go`: testable chi router, public health/assets/login, protected todo routes, POST logout, cross-origin protection, HTTP timeouts and environment-backed PostgreSQL startup; complete authenticated route workflow tests.
- `internal/store/store.go`, `memory.go`, `postgres.go`: shared list-option validation/defaults and not-found errors; configuration without embedded passwords; checked migrations; pooled connections; updates/deletes report missing rows instead of silently inserting/succeeding.
- `internal/store/*_test.go`: shared CRUD/filter/sort contract, concurrent memory access, dedicated test-database guard, per-test transaction/schema isolation, user operations and duplicate username handling. Existing PostgreSQL tests now use the isolated helper rather than the development database.
- `internal/auth/session.go`, auth tests: cryptographically random tokens, expiration, revocation, session locking, salted bcrypt verification and password-length error coverage.
- `internal/handlers/`: safe form parsing and limits, whitespace validation, generic storage errors, embedded rendering, session authentication, secure cookie options and expanded handler tests. Removed tests' process-wide working-directory changes.
- `internal/models/user.go`: prevents hash serialization as JSON.
- `cmd/seed/main.go`: credentials from environment, bcrypt, random user IDs, no overwrite of existing accounts, no printed passwords.
- `templates/`, `static/`: embed assets in binaries, combined filter/sort form, logout control, fetch error feedback, responsive metadata, login styling and keyboard focus styling; removed external font dependency.
- `docker-compose.yml`, `.env.example`, `.gitignore`: environment configuration, loopback database bindings, health checks, separate ephemeral PostgreSQL test service, ignored secrets.
- `.env`: newly created with mode 600 from existing container credentials, without printing them. Existing development volume retained. No test uses the development database.
- `README.md`, this audit: setup, behavior, commands and actual verification evidence.

## Commands and actual results

Commands were run from `/Users/shifashameer/todo` on 2026-10-19 (Asia/Kolkata).

| Command/check | Actual outcome |
|---|---|
| Initial `git status --short` | No Git repository present; no diff/status available |
| Baseline `go test ./...` | Auth/handler packages passed; store failed connecting to PostgreSQL inside sandbox |
| `go test ./internal/store -run TestMemoryStoreContract -count=1` | Initially failed invalid-option cases; passed after validation fix |
| New handler validation/forgery tests | Initially failed whitespace/malformed forms, bad-query status and forged user-ID acceptance; passed after fixes |
| `go test ./internal/auth -run TestSessions -count=1` | Initially missing implementation; passed after session manager added |
| `go test ./internal/handlers -run TestFormContentTypeWithCharset -count=1` | Initially 415 for valid charset form; passed after media-type parser fix |
| `docker compose ps` | Initially sandbox denied, then daemon stopped; started Docker Desktop successfully |
| `docker compose --profile test up -d --wait postgres-test` | Test container healthy on 5433; random temporary test credentials used |
| `go test ./internal/store -count=1 -v` | All PostgreSQL and memory tests passed against dedicated `todo_test`; no integration skips |
| Final `go test ./... -count=1` | All packages passed with PostgreSQL configured |
| Final `go vet ./...` | Exit 0, no diagnostics |
| Final `go test -race ./... -count=1` | All packages passed with PostgreSQL configured; no race reports |
| `gofmt -l main.go main_test.go cmd internal static/assets.go templates/assets.go` | No output (all formatted) |
| `go run ./cmd/seed` against test database | Temporary account created successfully; no password printed |
| `ADDR=127.0.0.1:18080 go run .` against test database | Migrations and startup succeeded |
| `python3 /tmp/todo-http-smoke.py` | Live health/assets/authentication/CRUD/filter/sort/CSRF/logout/replay assertions passed |
| `docker compose up -d --wait postgres` using preserved `.env` | Existing development PostgreSQL healthy; original volume retained |
| `go run .` using development `.env` | Migrations and startup succeeded on `127.0.0.1:8080` |
| `curl --fail --silent http://127.0.0.1:8080/healthz` | `ok` |
| Development login page HTTP check | 200 |
| `.env` permissions check | 600 |

Final Go commands used `GOCACHE=/tmp/todo-go-cache` due to the sandbox cache restriction. Database-enabled commands also set `TEST_DATABASE_URL` and `PGPASSWORD` for the isolated test service. Localhost/Docker access required sandbox escalation and was approved. An intermediate build caught an incorrect `mime.ParseMediaType` return-value count; this was fixed before the final successful full test/vet/race run.

The smoke-test script and temporary credential files were verification artifacts under `/tmp`, not application dependencies. Temporary app servers and test PostgreSQL were stopped after verification. Development PostgreSQL remains available for the documented run commands. No real initial-user password was chosen for the user: run the documented seed command with your own credentials, or use an existing account.

## Remaining limits

- No unresolved implementation or environment blocker for local use.
- Manual browser visual/interaction checklist is provided in README and remains unperformed.
- Sessions are intentionally process-local and reset on app restart; a multi-instance deployment needs shared session storage.
- Todos are a shared authenticated board, as in the original application; per-user ownership was not requested.
- Existing credentials were preserved privately, not rotated. `.env` must stay private.
