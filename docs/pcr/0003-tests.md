# 0003. Tests live next to the code and run on every PR

Status: Accepted

## Convention
- Tests are standard Go tests: `foo_test.go` next to `foo.go`, same package. They run with
  `go test ./...`, in CI on every PR and every merge to `main`.
- Every PR adds or updates tests that show its ticket's behaviour works. Writing that test first is a good
  way to pin down what "done" means.
- Tests never call the real Luma API: use `httptest.Server` or a fake.

## Why
One kind of test that always runs is easier to keep up than several kinds with their own rules.

## How to apply
CI runs `go vet ./...` and `go test -race ./...`. Run the same locally before pushing.
