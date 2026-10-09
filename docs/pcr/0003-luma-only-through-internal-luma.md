# 0003. Luma is called only through internal/luma

Status: Accepted

## Convention
Only `internal/luma` makes HTTP requests to Luma. It maps Luma's JSON to this service's own types
(`luma.Member`, `luma.Event`, …), and nothing outside it decodes Luma's wire format.

## Why
One place to handle the API key, pagination, timeouts, the rate limit and changes in Luma's API. The
service's own JSON stays stable when Luma's changes.

## How to apply
A new Luma endpoint means a method on `luma.Client` that returns this service's types, plus a test against
an `httptest.Server`.
