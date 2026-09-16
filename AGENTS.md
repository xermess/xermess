# Working in this repository

xermess is an OAuth 2.0 / OpenID Connect server in Go, with two SvelteKit apps
under `web/`: `console` (the admin panel) and `id` (sign-in pages and a user's
own account). `README.md` is the long-form documentation and is kept current —
read the part that covers what you are changing, and update it when your change
makes a sentence in it untrue.

## Commands

| What                       | Command                                              |
| -------------------------- | ---------------------------------------------------- |
| Format, vet and test Go    | `make check`                                         |
| Go tests only              | `make test`                                          |
| Tests that need Postgres   | `make test-integration` (runs every `TestLive…`)     |
| Lint and type-check apps   | `make web-check` (or `bun run lint` / `bun run check` in `web/<app>`) |
| Run everything             | `make dev` — API on :8080/:8081, id :5173, console :5174 |
| Migrations                 | `make migrate-up`, `migrate-down`, `migrate-status`, `migrate-new name=x` |

Before saying a change is done, run `make check` and, for anything the apps
touch, `make web-check`. Run `make test-integration` when you changed a
migration, the store, or a route table. Formatting is enforced: Go through
`gofmt`, the apps through Prettier — `bunx prettier --write <path>` fixes what
`bun run lint` refuses.

The apps log to `.logs/<app>.log`. `.env` holds the secret key and is not for
reading; `.env.example` documents every variable.

## Layout

```
cmd/xermess/main.go     startup, in order, in one function
internal/config         reads .env
internal/model          one file per table, all listed in model.All
internal/store          every query in the project, one file per subject
internal/auth           signing administrators in, and what they may do
internal/oidc           the provider: authorize, tokens, userinfo, logout
internal/api/server.go  the engine, and the table of every route
internal/api/<subject>/ handler.go, request.go, response.go, validation.go
migrations/             one Go file per migration, applied in order
web/console/src/lib/    api/, query/, components/ui/, components/<feature>/
web/console/src/routes/admin/(panel)/  everything behind a session
```

Nothing above `store` writes a query, and nothing below `api` knows about HTTP.
A handler reads a request, asks the store, and answers; the rules live in
`validation.go` and in the model, and return a `respond.Fault` carrying the
status. `respond.Failure` turns anything else into a 500 and logs it, so a
database error never reaches a browser.

## Adding things

**An endpoint.** Write the method in the package under `internal/api/` it
belongs to, keeping that package's four files, then mount it in
`registerRoutes` in `internal/api/server.go` — the one place that says which
paths exist, and on which of the two servers. Put it behind `session.Require`
and the permission it needs. Add it to the API table in `README.md`.

**A model.** A struct in `internal/model`, embedding `Base`, registered in
`model.All()` — and bump the count in `TestAllListsEveryModel`, which exists to
catch a model that never got a table. Add a migration for it: a Go file in
`migrations/` whose `up` calls `AutoMigrate` on the new models and whose `down`
undoes it. Never edit a migration that has already run.

**A permission.** A constant and a catalog entry in
`internal/model/admin_permission.go`, guarding the routes with `session.Can` or
`session.CanAnywhere`, and added to the `AdminPermissionName` union in
`web/console/src/lib/api/types.ts`. The panel builds its own permission UI from
the catalog the API serves, so nothing else needs changing.

**Something written to the activity log.** `audit.Record` or `RecordWith` from
the handler, then give the action a sentence, an icon and a category in
`web/console/src/lib/components/activity/actions.ts`, or the log shows its raw
name. If the target is a kind of record the log can name, say who may see that
name in `maySeeName` in `internal/api/activity/handler.go`.

**A setting the outside world sees.** The organisation's settings
(`internal/api/organization/`) are the example: the panel writes them, the
discovery document publishes the agreements (`internal/oidc/provider.go`), and
the sign-in pages read the rest from `GET /api/v1/account/organization`, which
takes no session — so only what a stranger may see goes in `PublicOrganization`.
A setting nothing reads is a note, not a setting; wire it somewhere or leave it
out.

**A panel page.** A `+page.server.ts` that checks the permission
(`requirePermission` / `requireAnywhere`) and fetches with `apiGet`, a
`+page.svelte` that seeds a TanStack query from what the server rendered, query
options in `lib/query/` with their key in `lib/query/keys.ts`, a typed client
call in `lib/api/admin.ts`, and the feature's components in
`lib/components/<feature>/`. Add the route to `sections.ts` with an `allowed`
check. A section that stops being a placeholder loses its `status: 'preview'`
there, its block in `lib/data/demo.ts`, and its mention in `README.md`.

## Style

The code is written to be read. Comments say *why* something is the way it is,
not what the next line does, and the prose is plain — look at a neighbouring
file and match it rather than inventing a second voice. Names are spelled out.
A file that needs a paragraph to explain itself has it at the top.

In the panel, a page never writes a colour, a height or a hover state: it names
one. Everything is built from `$lib/components/ui` — `Panel`, `List`,
`PageContainer`, `Input`, `Select`, `Button` — and `lib/components/ui/README.md`
says what each is for and where every kind of styling lives. Pages render on the
server, so data is fetched in `+page.server.ts`, never in `onMount`.

Tests are table-driven, named for what they prove — "a slug with spaces", not
"TestCase3". The ones that need Postgres are named `TestLive…` and skip
themselves without `XERMESS_TEST_DB_DSN`.

## Seeing a change in the app

`make dev` is the normal way. If a stack is already running (check ports 8080,
8081, 5173, 5174 before starting one), do not restart it — `/preview` starts a
throwaway instance on spare ports and its own database, and tears it down after.

## What is in .claude

```
.claude/settings.json      the permission allowlist, and the formatting hook
.claude/hooks/format.sh    what that hook runs: gofmt, or the app's Prettier
.claude/commands/check.md      /check      — every check this repository has
.claude/commands/preview.md    /preview    — the app in a browser, on spare ports
.claude/commands/endpoint.md   /endpoint   — add an admin API endpoint
.claude/commands/migration.md  /migration  — write a migration
```

The hook formats each file as it is written, so an edit cannot fail the lint
on spacing alone. It is not a substitute for running the checks.
`.claude/settings.local.json` is for personal overrides and is not committed.
