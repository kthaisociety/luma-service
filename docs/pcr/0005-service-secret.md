# 0005. Callers authenticate with a shared secret

Status: Accepted

## Convention
Everything under `/v1` requires the header `X-Service-Secret`, compared in constant time with
`LUMA_SERVICE_SECRET`. An empty configured secret rejects every request. `/health` is open. The service is
internal only (no public domain): callers reach it on `dokploy-network`.

## Why
Same pattern as onboarding-service. Several services will call this one, and it serves members' personal
data.

## How to apply
Each environment has its own secret, shared by all callers, in OpenBao at
`secret/shared/luma-service-secret/<env>`. Replace this with identity-service tokens once that service
exists.
