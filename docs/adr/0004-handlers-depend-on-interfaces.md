# 0004. Handlers depend on interfaces

Status: Accepted

## Context
Handler tests shouldn't need a Luma API key or the network.

## Decision
`internal/handlers` declares the interface it needs (`LumaReader`) and receives an implementation in
`Deps`. `cmd/api` wires in the real `*luma.Client`; tests pass a fake.

## Consequences
New dependencies of the HTTP layer (the database, once ADR 0001 is decided) are added the same way: an
interface in `handlers`, the implementation wired in `cmd/api`.
