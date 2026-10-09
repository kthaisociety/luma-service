# luma-service

The one KTHAIS service that talks to the [Luma API](https://docs.luma.com). Other services call it instead
of holding their own Luma keys. Read-only for now.

## API

Everything under `/v1` needs the header `X-Service-Secret: <LUMA_SERVICE_SECRET>`.

| Route | Returns |
|---|---|
| `GET /health` | `{"service":"luma-service","status":"healthy"}`, or 503 if the database is unreachable. No secret needed. |
| `GET /v1/calendar` | the calendar the API key belongs to |
| `GET /v1/events?after=&before=` | `{"events":[…]}`; `after`/`before` optional, RFC 3339 |
| `GET /v1/tiers` | `{"tiers":[…]}`, each with its registration questions |
| `GET /v1/community/members` | `{"tier_id":"…","members":[…]}`: the Community tier's approved members and their answers |

Luma failures return 502 with no detail; the detail is in the log.

## Running locally

Needs Go 1.26 and Docker.

```sh
docker compose up -d          # Postgres on 127.0.0.1:5434
cp .env.example .env          # then fill in LUMA_API_KEY, LUMA_COMMUNITY_TIER_ID, LUMA_SERVICE_SECRET
go run ./cmd/api
curl -H "X-Service-Secret: $(grep ^LUMA_SERVICE_SECRET .env | cut -d= -f2-)" localhost:8001/v1/tiers
```

Tests: `make test` (`go vet ./...` and `go test -race ./...`; database tests need `docker compose up -d`). They never call the real Luma API.

The schema is SQL files in `internal/database/migrations/`, applied by the server at startup
([PCR 0006](docs/pcr/0006-sql-migrations.md)).

## Configuration

All from env ([PCR 0001](docs/pcr/0001-config-from-env.md)); see `.env.example`.

| Variable | Secret | |
|---|---|---|
| `DATABASE_URL` | yes | Postgres connection string |
| `LUMA_API_KEY` | yes | Luma calendar API key |
| `LUMA_SERVICE_SECRET` | yes | shared with every caller |
| `LUMA_COMMUNITY_TIER_ID` | no | the Community tier |
| `HOST`, `PORT` | no | default all interfaces, `8000` |

## Delivery

Like every KTHAIS app ([infrastructure/docs/app-delivery.md](https://github.com/kthaisociety/infrastructure/blob/main/docs/app-delivery.md)):
PRs need a Conventional Commit title; every merge to `main` builds `ghcr.io/kthaisociety/luma-service:sha-<7>`
and deploys it to staging; merging the release PR releases `X.Y.Z` for production.

## Docs

- [`CONTEXT.md`](CONTEXT.md): vocabulary
- [`docs/adr/`](docs/adr/): decisions about one part of the code
- [`docs/pcr/`](docs/pcr/): conventions for the whole project
