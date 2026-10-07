# 0003. The Community tier is set by environment

Status: Accepted

## Context
Luma API keys belong to one calendar, so the key alone decides which calendar is read. The growth team
needs the members of one tier, "Community".

## Decision
`LUMA_COMMUNITY_TIER_ID` names that tier, and `/v1/community/members` serves it. Callers don't pass tier IDs.

## Consequences
Callers don't need to know Luma IDs. Serving another tier means a new endpoint or a generic
`/v1/tiers/:id/members`; add it when a caller needs one.
