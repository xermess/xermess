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
internal/cache          Redis: the cache in front of the store, the rate limit's counts
internal/auth           signing administrators in, and what they may do
internal/oidc           the provider: authorize, tokens, userinfo, logout
internal/api/server.go  the engine, and the table of every route
internal/api/<subject>/ handler.go, request.go, response.go, validation.go
migrations/             the schema; one file today, applied in order
web/console/src/lib/    api/, query/, components/ui/, components/<feature>/
web/console/src/routes/admin/(panel)/  everything behind a session
```

Nothing above `store` writes a query, and nothing below `api` knows about HTTP.
A handler reads a request, asks the store, and answers; the rules live in
`validation.go` and in the model, and return a `respond.Fault` made from a
problem — a status and a code the apps translate (see "An error" below).
`respond.Failure` turns anything else into a 500 and logs it, so a database
error never reaches a browser.

## Adding things

**An endpoint.** Write the method in the package under `internal/api/` it
belongs to, keeping that package's four files, then mount it in
`registerRoutes` in `internal/api/server.go` — the one place that says which
paths exist, and on which of the two servers. Put it behind `session.Require`
and the permission it needs. Add it to the API table in `README.md`.

**A model.** A struct in `internal/model`, embedding `Base`, registered in
`model.All()` — and bump the count in `TestAllListsEveryModel`, which exists to
catch a model that never got a table.

While `migrations/` holds only the schema migration, that is all a new model
needs: it builds from `model.All()`, so the table follows, and a development
database is rebuilt rather than stepped forward (`make db-reset`). Add the
`down` entry for it, in an order that drops children first. Once a database
exists that cannot be rebuilt, the schema file stops being editable and a new
migration is written instead — `AutoMigrate` on the new models in `up`, the
undo in `down` — and from then on no migration that has run anywhere is ever
edited.

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

**A provider users can sign in with.** The kinds live in `model.SocialSpecs`,
one entry per provider: its endpoints, the scopes to ask for, where the
identity is in the answer, and the quirks (Apple posts its answer and signs
its own secret; Yandex reads `OAuth` rather than `Bearer`; VK puts the profile
under `response.0` and the address beside the token). Adding one is an entry
there — the flow, the panel and the sign-in pages all read it. Nothing about a
provider belongs in the database except what that installation was given:
credentials, and the addresses of a provider this server does not know.

**A setting that decides how people sign in.** Two exist to copy: the
organisation's, and `admin_security`, which says whether administrators need a
second factor. Both are a singleton row, seeded from the configuration on the
first start and owned by the panel afterwards, and both are read where they
matter rather than held in a field from startup — so a change takes effect on
the next request instead of the next restart.

**A step a login flow can name.** The catalog is `model.LoginStepSpecs`, one
entry per step: what it is called, what it does, and `Implemented` — whether
this server runs it yet. The panel builds its step picker from the catalog and
marks the rest "not run yet", so a step can be offered as a plan before it is
built. Implementing one is flipping that flag and writing the step; nothing
else reads a hard-coded list of steps.

**An error.** Never a sentence in Go: define the problem next to the handler
that returns it, `var taken = respond.Define(http.StatusConflict,
"language_code_taken", respond.Admin)`, answer with `respond.Fail(c, taken)`
or return `taken.With("code", code)`, and add `error.language_code_taken` to
`locales/<app>/*.json` for every app it is for (`Public` → `id`, `Admin` →
`console`, `Both`). The server's English is read from `en.json`, and the apps
show the key in the reader's language with `messageOf(err, t)`.
`TestErrorCodesMatchTheCatalogs` fails until both sides agree.

**A cached read.** Only in the store, with `cached(ctx, s, group, field, load)`,
and only for something every page asks for and almost nothing changes — never
users, sessions, tokens or administrators. Every store method that writes it
calls `s.forget(ctx, group)` after the write commits; a new group is a
constant in `internal/cache`. Cache a projection rather than a model that
carries a secret, and add the type to `TestCachedTypesSurviveJSON`. Redis is
optional: a nil cache is an empty one, so everything has to work without it.

**A language.** An installation adds its own on the Languages page: the text
lives in the database (`languages`, and a `translations` row per app) and the
apps fetch it while rendering, so nothing is rebuilt. To *ship* one with the
server, add a JSON file under `locales/id/` — and under `locales/console/`
only for the panel's languages, `locales.PanelLanguages` (English and
Russian) — named after the language tag,
copied from `en.json` with the values translated and `$name`/`$native` naming
the language. The first start imports every shipped file
(`store.EnsureLanguages`), and later starts copy in keys a release added
without touching anything an administrator wrote.
`TestTranslationsAreComplete` fails when a shipped language falls behind `en`,
so a key added to the base has to be added to every file before it ships.

**Text in the apps.** Never a literal in the markup: `const t = useTranslator()`
at the top of the component and `t('area.thing')` where the words go, with the
key added to every file under `locales/<app>/` — nested under its screen,
`{"area": {"thing": "…"}}`. Parameters are `{braces}` in
the text and an object at the call — `t('login.subtitle_app', { app: name })`.
The sign-in pages are fully moved over; the panel's shell is, and its other
pages are not yet.

**A panel page.** A `+page.server.ts` that checks the permission
(`requirePermission` / `requireAnywhere`) and fetches with `apiGet`, a
`+page.svelte` that seeds a TanStack query from what the server rendered, query
options in `lib/query/` with their key in `lib/query/keys.ts`, a typed client
call in `lib/api/admin.ts`, and the feature's components in
`lib/components/<feature>/`. Add the route to `sections.ts` with an `allowed`
check.

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
