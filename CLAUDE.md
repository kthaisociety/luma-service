# CLAUDE.md

luma-service: the one KTHAIS service that talks to the Luma API. Read-only for now: calendar, events, tiers
and the Community tier's members. Go, Gin, GORM, Postgres. Internal only: other services call it over
`dokploy-network` with a shared secret.

## Commands

- Local Postgres: `docker compose up -d`
- Config: `cp .env.example .env`, then fill in the secrets
- Run: `go run ./cmd/api` (listens on `127.0.0.1:8001` with the example env)
- Test: `go vet ./... && go test -race ./...`
- Acceptance tests (ticket branches only): `go test -tags acceptance ./test/acceptance/...`
- Image: `docker build -t luma-service .`

## Layout

- `cmd/api`: wiring and startup
- `internal/config`: env, validated at boot
- `internal/luma`: the Luma API client and this service's types; the only code that calls Luma
- `internal/handlers`: Gin routes; depend on interfaces, not on clients
- `internal/database`: Postgres connection
- `test/acceptance`: per-ticket acceptance tests

## Rules

- Read `CONTEXT.md` for vocabulary, `docs/adr/` for decisions and `docs/pcr/` for conventions before
  changing a part they cover. Record a new decision as a new ADR or PCR, never by editing an accepted one.
- Only `internal/luma` calls Luma (PCR 0004). Tests never call the real Luma API.
- What Postgres holds is undecided (ADR 0001): don't add models or a sync job without a new ADR.
- Never log or return API keys, secrets or members' personal data in errors.
- PR titles and commits are Conventional Commits (`feat:`, `fix:`, `docs:`, …): they drive releases.
