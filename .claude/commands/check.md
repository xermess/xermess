---
description: Run every check this repository has — Go and both apps
allowed-tools: Bash(make check), Bash(make web-check), Bash(make test-integration), Bash(gofmt:*), Bash(bunx prettier:*), Read, Edit, Glob, Grep
---

Run the checks, in this order, and fix what they find:

1. `make check` — gofmt, go vet and the Go tests.
2. `make web-check` — lint and type-check `web/console` and `web/id`.
3. `make test-integration` — the `TestLive…` tests, on throwaway Postgres
   databases. Run it when the change touched a migration, the store, or the
   route table in `internal/api/server.go`; say so if you skip it.

Formatting is not a judgement call: `gofmt -w <file>` and, inside the app the
file belongs to, `bunx prettier --write <file>`.

Report what passed and what did not. If something fails, fix the cause rather
than the test, unless the test is the thing that is wrong — and say which you
decided it was.
