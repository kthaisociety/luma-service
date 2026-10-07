# 0001. Storage strategy: open

Status: Proposed

## Context
The service has its own Postgres, but it is not decided yet what it holds. The two candidates:

- **Periodic sync.** A background job copies events, tiers and members from Luma into Postgres on an
  interval; callers read from Postgres. Fast, gentle on Luma's rate limit (200 requests/minute per
  calendar key), and keeps history (who was a member when). More code: a sync job, a schema, staleness.
- **Read-through cache.** Every request goes to Luma; responses are kept in Postgres for a TTL. Less code,
  but slower, closer to the rate limit, and no history.

## Decision
Not decided. Until it is, the service connects to Postgres at boot (and `/health` checks it) but stores
nothing: every `/v1` request reads straight from Luma.

## Consequences
- Each `/v1/community/members` call costs one Luma request per 100 members. Fine for a proof of concept.
- When this is decided, write a new ADR that supersedes this one.
