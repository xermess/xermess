# xermess

An authentication server written in Go, with its frontends under `web/`:
`web/console`, the admin panel, and `web/id`, where users sign in and
manage their accounts.

An OAuth 2.0 authorization server and OpenID Connect provider: authorization
code with PKCE, refresh tokens, client credentials, userinfo, logout,
revocation and introspection, with hosted sign-in pages for users and an admin
panel for everything else.

## Requirements

- Go 1.27+
- PostgreSQL 14+
- Redis 6.2+ — optional, but configured by default (`brew install redis`,
  then `brew services start redis`); see [Redis](#redis)
- [Bun](https://bun.sh) and Node.js 24+ — for the apps under `web/`

## Getting started

```sh
make setup              # .env with a secret key, the database, dependencies
make full-start dev     # or: make dev
```

That starts the API and both apps. The API's logs stay in the terminal;
the apps log to `.logs/<app>.log` (`tail -f .logs/id.log`). Ctrl-C stops all of
them.

| What        | URL                                  |
| ----------- | ------------------------------------ |
| id          | http://localhost:5173                |
| console     | http://localhost:5174/admin/login    |
| API         | :8080 public, :8081 admin            |

`make full-start prod` (or `make prod`, or `make full-start -- --prod`) builds
everything for production and serves the builds on the same URLs, so `.env`
fits both modes. A bare `make full-start --prod` cannot work: make reads
`--prod` as an option of its own. `scripts/start.sh --dev|--prod` is the same
thing without make.

Each app serves the API paths it needs on its own origin, so the browser never
calls the API directly: Vite's proxy does the routing in dev, `web/serve.js` in
local prod, Caddy in a deployment (`deploy/README.md`).

## Make targets

Run `make` for the list:

| Group    | Targets                                                                  |
| -------- | ------------------------------------------------------------------------ |
| Setup    | `setup`                                                                  |
| Run      | `full-start dev\|prod`, `dev`, `prod`, `run` (API only), `build`          |
| Quality  | `check`, `test`, `test-integration`, `web-check`, `web-build`            |
| Database | `migrate-up`, `migrate-down`, `migrate-status`, `migrate-new name=x`, `db-create`, `db-reset`, `db-psql` |
| Deploy   | `deploy-build`, `deploy-up`, `deploy-down`, `deploy-logs`, `clean`       |

## Layout

```
cmd/xermess/main.go            startup, in order, in one function
cmd/migrate/main.go            runs migrations by hand: up, down, status

i18n/                       the sign-in pages' translations, grouped by app, language,
                               and subject, imported into the database on its first start;
                               i18n/id/en/ is the key contract, and i18n/console/en/
                               the admin API's English for its errors

internal/config/config.go      reads .env
internal/database/database.go  opens the connection
internal/database/migrate.go   applies migrations
internal/cache/                Redis: the cache in front of the store, and the rate limit's counts
internal/store/                every query in the project, one file per subject
internal/auth/auth.go          signs administrators in and out, records what they do
internal/oidc/                 the OAuth 2.0 / OpenID Connect provider: authorize, tokens,
                               userinfo, logout, sessions, registration, password resets,
                               and signing in with an account somewhere else
internal/jose/                 signing and checking JWTs, publishing keys, sealing them
internal/mail/                 sending email over SMTP, or into the log; the settings are
                               read per message, from the Mail page
internal/api/server.go         the engine, and the table of every route
internal/api/auth/             signing in and out, and who is signed in
internal/api/oauth/            the provider endpoints: /oauth2/* and /.well-known/*
internal/api/account/          what the id app calls: login, register, reset, the account
internal/api/setup/            creating the first administrator
internal/api/mfa/              an administrator's authenticator and recovery codes
internal/api/keys/             listing and rotating token signing keys
internal/api/apis/             the APIs (resource servers) tokens are issued for
internal/api/users/            the users an organisation manages
internal/api/fields/           the fields a user record is made of
internal/api/applications/     OAuth 2.0 / OIDC clients: settings, secrets
internal/api/roles/            the roles users hold in each application
internal/api/admins/           administrators, managed by a super admin
internal/api/adminroles/       admin roles and the permissions they grant
internal/api/organization/     the organisation this installation belongs to
internal/api/social/           the providers users may sign in with
internal/api/flows/            the login flows applications sign their users in with
internal/api/sessions/         everyone signed in, and signing them out
internal/api/languages/        the languages, their text, and the panel's own
internal/api/mail/             the mail server, a test message, and the words of every email
internal/api/otp/              how the emailed one-time codes behave
internal/api/activity/         the dashboard counts and the log
internal/api/middleware/       request logging, recovery, and their order
internal/api/cors/             which browser origins may call the API
internal/api/csrf/             changes only from the app's own origin, and only as JSON
internal/api/ratelimit/        per-address limits on passwords and emails
internal/api/query/            paging and filter parameters shared by the lists
internal/api/validate/         request validation rules
internal/api/session/          the cookie, the session check, the current admin
internal/api/respond/          how an error is written, once for every endpoint
internal/api/audit/            recording what an administrator did
internal/model/                one file per table, listed in model.All

migrations/                    the schema, as Go files applied in order

AGENTS.md                      how to work in this repository, for people and agents
.claude/                       Claude Code's settings, formatting hook and commands

scripts/setup.sh               prepare a fresh checkout
scripts/start.sh               run the API and the apps, --dev or --prod
scripts/db.sh                  create, reset or open the database
scripts/lib.sh                 what the scripts share: .env, checks

deploy/compose.yaml            the production stack: Postgres, API, apps, Caddy
deploy/Caddyfile               the edge: which paths go to the API, staff-only admin
deploy/docker/                 api.Dockerfile, web.Dockerfile (any app), their ignore files

web/console/                   the admin panel (SvelteKit)
web/id/                        the users' app: sign-in pages and account management (SvelteKit)
web/serve.js                   serves an app's build with its API paths in front, for make prod

web/console/src/lib/api/               the typed client for this API
web/console/src/lib/query/             the query cache: its client, its keys, its options
web/console/src/lib/components/ui/     the design system: Button, Drawer, DataTable, Panel, List, Tag…
web/console/src/lib/components/layout/ the panel's frame: header, sidebar, account menu
web/console/src/lib/components/users/  the users feature: table, drawers, field inputs
web/console/src/lib/components/roles/  roles, their mappings and pickers
web/console/src/lib/components/applications/ applications, their API access, token preview
web/console/src/lib/components/apis/   APIs, their scopes and settings
web/console/src/lib/components/admins/ administrators and admin roles
web/console/src/lib/components/organization/ the organisation's settings
web/console/src/lib/components/social/ the providers users sign in with, and what each kind needs
web/console/src/lib/components/activity/ the dashboard: chart, sign-ins, feed, what the log's actions mean
web/console/src/lib/components/profile/  the account drawer: its two-factor, sessions and signing out
web/console/src/lib/state/             what the panel remembers: the theme, the sidebar's
                                       width and which of its sections are folded
web/console/src/lib/utils/             how values are shown
web/console/src/lib/data/demo.ts       placeholder rows for the sections with no backend
web/console/src/lib/server/api.ts      calling the API from a server load, with the session
web/console/src/lib/constants.ts       the names both sides agree on: the cookies
web/console/src/lib/styles/            fonts.css, tokens.css, base.css, fields.css, ark.css
web/console/src/routes/admin/login/    the sign-in page
web/console/src/routes/admin/(panel)/  everything behind a session

web/id/src/routes/favicon.ico/ the icon for the paths that answer with something other than a page
web/id/src/routes/(auth)/     the sign-in pages users reach from an application
web/id/src/routes/(account)/  a signed-in user's profile, security and connected apps
web/id/src/lib/components/    its own small component set, on Svelte alone
```

### The API packages

Handlers are grouped by subject, and every group is the same four files, so a
package you have never opened is laid out like the last one you did:

```
internal/api/users/handler.go     the endpoints: what happens, in order
internal/api/users/request.go     the bodies and query strings it accepts
internal/api/users/response.go    the shapes it answers with
internal/api/users/validation.go  the rules a request has to keep
```

A handler reads a request, asks the store, and answers. Nothing else: the
rules live in `validation.go` and return a `respond.Fault` carrying the status
to answer with, and `respond.Failure` turns that into the answer — or logs
anything that is not a Fault and says only that something went wrong.

### Data in the panel

A page arrives already rendered: `+page.server.ts` asks the API with the
session cookie, so the first paint costs no request from the browser. What
that load returned then seeds a TanStack Query cache — `usersOptions(params,
data.page)` — and everything after the first paint goes through the cache
instead of through the page.

That is what a save does: the drawer runs a `createMutation`, and on success
invalidates `keys.users.all`. The list refills itself without a navigation,
the URL does not change, and nothing else on the page is thrown away. The
refresh button is the same invalidation by hand, and the cache refills on its
own when someone comes back to the tab.

Signing out calls `queryClient.clear()`: what one administrator saw is not for
whoever signs in next on that browser.

```
web/console/src/lib/query/client.ts    the client, and what its defaults mean
web/console/src/lib/query/keys.ts      the names the cache knows things by
web/console/src/lib/query/users.ts     the options for the user list and the fields
```

### The store

`internal/store` is the only package that writes queries. A handler asks it
for what it needs — `store.Users`, `store.UserField`, `store.WriteAudit` — and
gets models back, so the handlers stay about HTTP and the queries stay in one
place to read and change. It has its own errors, `store.ErrNotFound` and
`store.ErrDuplicate`, which is why nothing above it imports GORM.

A few of its reads go through Redis first (`store.WithCache`); see
[Redis](#redis). Which ones is decided here too, so a handler cannot tell a
cached answer from a fresh one.

```
internal/store/store.go        the Store type, and the errors it returns
internal/store/users.go        listing, searching and writing users
internal/store/user_fields.go  the field definitions
internal/store/user_roles.go   roles users hold, and their mappings
internal/store/applications.go OAuth clients and what they may reach
internal/store/apis.go         APIs and their scopes
internal/store/admins.go       administrators, for signing in
internal/store/admin_roles.go  admin roles and their permissions
internal/store/organizations.go the organisation's settings, the one row of them
internal/store/social.go       the providers, the identities held at them, the sign-ins away at one
internal/store/login_flows.go  the login flows, and which applications hold each
internal/store/languages.go    the languages and their text, and importing the shipped ones
internal/store/user_sessions.go the Sessions page's list, and signing a user out everywhere
internal/store/sweep.go        deleting what has expired, every hour
internal/store/admin_security.go how administrators are made to sign in
internal/store/sessions.go     sessions: start, find, revoke, list
internal/store/mfa.go          authenticators and recovery codes
internal/store/oauth.go        codes, refresh tokens, user sessions, signing keys
internal/store/audit.go        the activity log, and the dashboard counts
```

Each package has one job: `config` reads settings, `database` opens the
connection, `model` describes the tables, `store` queries them, `auth` decides
who may sign in, `api` answers requests. Nothing imports `api` except `main`,
and `model` imports nothing of ours at all.

Startup is `run` in `cmd/xermess/main.go`, top to bottom: read the
configuration, open the database, apply migrations, build the store and the
provider, serve.

`api.NewPublic` and `api.NewAdmin` build the two Gin engines, and `serve` runs
them on their listeners. A path exists on one of them only: the admin API is
never on the public listener, which is what keeps it off the internet however
the proxy in front is configured. Ctrl-C or SIGTERM stops both and lets
requests under way finish, for up to 15 seconds.

Administrators can sign in with a second factor — a TOTP authenticator app,
with single-use recovery codes. It is **off by default**, and requiring it is
a super admin's decision on the Administrators page rather than each
administrator's own: an administrator without one is then taken to
`/admin/mfa-setup` to enrol before they may do anything else, and the same
page lets a signed-in one set an authenticator up before it is compulsory.
Whoever already has one shows as such on the Administrators page, where a
super admin can also reset it.

Whether it is compulsory for everybody is a setting on the Administrators
page, not a line in a file: `XERMESS_ADMIN_MFA` is only what a fresh
installation starts with, written to `admin_security` on the first start, and
a super admin owns it from then on. Off is the default there too — the
alternative decides for people who have not been asked, and would send the
first administrator to set up an authenticator before they had seen the panel.
Turning it on takes effect at once: an administrator without one can then do
nothing but set one up, which is what the panel warns about before making the
change, since that includes whoever is making it. A
session whose password was right but whose code is still to come can do nothing
but give the code; one that has to set an authenticator up can do nothing but
that (`internal/auth/mfa.go`, `internal/totp`). Token signing keys rotate every
`XERMESS_KEY_ROTATION_DAYS`, a new one published a day before it signs.

**Nobody is asked to allow an application.** A user already signed in here is
sent straight back to the application that asked, with a code, and every
application registered in the panel can have tokens for them. That is the
right default for the server an organisation runs for its own applications —
Keycloak does the same with consent off — and it is the whole reason
registering an application is a panel permission (`applications.write`) rather
than something a client can do for itself: the decision about which
applications may act for a user is made once, by an administrator, instead of
by each user at each sign-in. An installation that means to hand client
registration out more widely needs the consent step first; it is in the
catalog as a plan and is not run yet (see **Login flows**).

Every cookie-authenticated API takes changes only from its app's origin
(`internal/api/csrf`): another origin gets 403, a body that is not JSON 415.
Sign-in, registration, setup and password resets are rate limited per address
(`internal/api/ratelimit`), and so are the token, revocation and introspection
endpoints — on a budget of their own, ten times looser, because a sign-in comes
from one person's browser while a token request may come from one backend
exchanging codes for a whole company. With Redis the count is shared by every
server process and survives a restart; without it, each process counts on its
own.

## API

### Errors

Every error either server answers with has the same shape:

```json
{ "error": "Too many attempts. Try again in 42 seconds.", "code": "rate_limited", "params": { "seconds": 42 } }
```

`code` is what the apps use: they show `error.<code>` from their own catalog,
in the reader's language, with `params` filled in — a field name among them
is said the way the form labels it — so a Russian reader sees Russian from the
server as much as from the page. `error` is the same sentence in English for
whoever calls the API directly, and it is not written in Go: it is the base
catalog's text for the key, so the English exists once.

Each problem is defined next to the code that returns it —
`respond.Define(status, code, apps)` — and answered with `respond.Fail` or
returned as `problem.With("name", value)`. The validator answers
`validation.<rule>` with `field` and the rule's own limit; the sign-in
service's refusals (`oidc.Problems`) become public problems on their own.
`TestErrorCodesMatchTheCatalogs` holds the two sides together: every problem
has its sentence in each app it is for, the same English wherever it is said
twice, and every `error.*` key in a catalog is one the server can answer with.

The reset email is written the same way, in the language the page asked in
(`email.reset.*`), and the default language when that one is not offered.

The public API's errors all have codes, and so do the ones every endpoint
shares — the validator, sessions, permissions, the CSRF check, the rate
limit — and the Languages page's. Some admin endpoints still answer with
English alone, as their pages are still English in the markup; giving one a
code is defining its problem and adding the sentence to
`i18n/console/en/server.json`.

### Endpoints

| Method | Path                          | Needs a session | Description                    |
| ------ | ----------------------------- | --------------- | ------------------------------ |
| `GET`  | `/healthz`                    | no              | The server is up               |
| `POST` | `/api/v1/admin/auth/login`    | no              | Sign in, sets the session cookie |
| `POST` | `/api/v1/admin/auth/logout`   | yes             | Sign out, revokes the session  |
| `GET`  | `/api/v1/admin/me`            | yes             | The signed-in administrator    |
| `PATCH`| `/api/v1/admin/me`            | yes             | Change your own name and email; a new email takes your current password |
| `POST` | `/api/v1/admin/me/password`   | yes             | Change your own password; signs out your other sessions |
| `GET`  | `/api/v1/admin/overview`      | yes             | Counts and recent activity     |
| `GET`  | `/api/v1/admin/logs`          | yes             | The activity log (`?limit=`)   |
| `GET`  | `/api/v1/admin/users`         | yes             | List users (`?search=&verified=&limit=&offset=`) |
| `POST` | `/api/v1/admin/users`         | yes             | Create a user                  |
| `GET`  | `/api/v1/admin/users/:id`     | yes             | One user                       |
| `PATCH`| `/api/v1/admin/users/:id`     | yes             | Update a user                  |
| `DELETE`| `/api/v1/admin/users/:id`    | yes             | Delete a user                  |
| `GET`  | `/api/v1/admin/user-fields`   | yes             | The fields a user record has   |
| `POST` | `/api/v1/admin/user-fields`   | yes             | Add a field                    |
| `DELETE`| `/api/v1/admin/user-fields/:id` | yes          | Remove a field                 |
| `GET`  | `/api/v1/admin/organization`  | yes             | The organisation               |
| `GET`  | `/api/v1/admin/security`      | yes             | How administrators are made to sign in |
| `PATCH`| `/api/v1/admin/security`      | yes             | Change it (a super admin's)    |
| `GET`  | `/api/v1/admin/social-providers` | yes          | The providers, and the kinds one may be |
| `POST` | `/api/v1/admin/social-providers` | yes          | Register a provider            |
| `GET`  | `/api/v1/admin/sso-connections` | yes           | The organisations' identity providers |
| `POST` | `/api/v1/admin/sso-connections` | yes           | Connect one (OIDC or SAML); SAML gets its own signing key |
| `POST` | `/api/v1/admin/sso-connections/test` | yes      | Try an issuer's discovery, or SAML metadata, before saving |
| `GET`  | `/api/v1/admin/sso-connections/:id` | yes       | One connection, with what to give its provider |
| `PATCH`| `/api/v1/admin/sso-connections/:id` | yes       | Change it                      |
| `DELETE`| `/api/v1/admin/sso-connections/:id` | yes      | Remove it and the identities held at it |
| `POST` | `/api/v1/admin/sso-connections/:id/refresh-metadata` | yes | Read a SAML provider's metadata again |
| `GET`  | `/api/v1/admin/login-flows`   | yes             | The login flows, and the steps one can be made of |
| `POST` | `/api/v1/admin/login-flows`   | yes             | Write a flow                   |
| `PATCH`| `/api/v1/admin/login-flows/:id` | yes           | Change a flow                  |
| `DELETE`| `/api/v1/admin/login-flows/:id` | yes          | Remove a flow                  |
| `GET`  | `/api/v1/admin/languages`     | yes             | The languages, and how much of each is translated |
| `POST` | `/api/v1/admin/languages`     | yes             | Add one, empty or copied from another |
| `PATCH`| `/api/v1/admin/languages/:code` | yes           | Rename it, offer it, or make it the default |
| `DELETE`| `/api/v1/admin/languages/:code` | yes          | Remove it and its text         |
| `GET`  | `/api/v1/admin/languages/:code/translations/:app` | yes | Its text for `id`, beside the English |
| `PUT`  | `/api/v1/admin/languages/:code/translations/:app` | yes | Replace that text              |
| `GET`  | `/api/v1/admin/mail`          | yes             | How this installation sends email (a super admin's) |
| `PATCH`| `/api/v1/admin/mail`          | yes             | Change the mail server; a password sent is sealed, never read back |
| `POST` | `/api/v1/admin/mail/test`     | yes             | Send one test message with what the form holds |
| `GET`  | `/api/v1/admin/mail/content`  | yes             | The words of every email, per language, beside the shipped English |
| `PUT`  | `/api/v1/admin/mail/content/:code` | yes        | Write one language's words for them |
| `GET`  | `/api/v1/admin/otp`           | yes             | The emailed one-time codes, and the flows that ask for one |
| `PATCH`| `/api/v1/admin/otp`           | yes             | Change how long a code is, lasts and may be guessed at |
| `GET`  | `/api/v1/admin/user-sessions` | yes             | Active sessions, newest first (`?search=&user=&after=&limit=`) |
| `DELETE`| `/api/v1/admin/user-sessions/:id` | yes         | Sign one session out           |
| `DELETE`| `/api/v1/admin/users/:id/sessions` | yes        | Sign a user out everywhere: every session, every application's tokens |
| `PATCH`| `/api/v1/admin/organization`  | yes             | Change its settings            |
| `GET`  | `/api/v1/admin/sessions`      | yes             | The caller's own sessions      |

The table above is the start of the admin API; the full list, with the
permission each route needs, is in `registerRoutes` in `internal/api/server.go`.

### The provider

| Method | Path                                  | Description                                  |
| ------ | ------------------------------------- | -------------------------------------------- |
| `GET`  | `/.well-known/openid-configuration`   | Discovery document, with the organisation's `op_tos_uri` and `op_policy_uri` |
| `GET`  | `/.well-known/jwks.json`              | Public signing keys                          |
| `GET`  | `/oauth2/authorize`                   | Start a sign-in; redirects to the sign-in page or back with a code |
| `POST` | `/oauth2/token`                       | `authorization_code`, `refresh_token`, `client_credentials` |
| `GET`  | `/oauth2/userinfo`                    | Claims about the user behind an access token |
| `GET`  | `/oauth2/logout`                      | RP-initiated logout                          |
| `POST` | `/oauth2/revoke`                      | Revoke a refresh token                       |
| `POST` | `/oauth2/introspect`                  | Whether a token is active                    |
| `GET`  | `/oauth2/social/:slug/start`          | Send the browser to a provider to sign in there |
| `GET`/`POST` | `/oauth2/social/:slug/callback` | Where the provider sends it back             |
| `GET`  | `/api/v1/account/requests/:handle`    | The sign-in under way, for the sign-in page  |
| `POST` | `/api/v1/account/login`               | Sign a user in; answers where to go next     |
| `POST` | `/api/v1/account/register`            | Create an account for a sign-in under way    |
| `POST` | `/api/v1/account/forgot-password`     | Email a reset link                           |
| `POST` | `/api/v1/account/reset-password`      | Set a new password through a reset link      |
| `POST` | `/api/v1/account/verify-email`        | Confirm an address through a verification link |

What a token carries is decided in one place, `model.EvaluateToken`, which the
panel's token preview runs too. Signing keys are made on first start, one per
algorithm, and stored encrypted with `XERMESS_SECRET_KEY`. Every code, refresh
token, session and reset link is stored as a SHA-256 hash. Refresh tokens
rotate, and a replayed one revokes its whole family.

The sign-in pages are `web/id`, a SvelteKit app of their own. The
authorization endpoint sends browsers to its `/login` with a handle for the
sign-in; the page shows the application's name, logo, terms and privacy links
from that handle, and posts to the account endpoints, which set the user's
session cookie — a different cookie from the panel's — and answer with the
redirect back to the application.

With that session, the same app is the user's account: their name, their
password, the devices they are signed in on, and the applications holding
refresh tokens for them, each of which they can sign out or disconnect. Those
endpoints take nothing but the session, so a user only ever reaches their own
account:

| Method   | Path                                                 | Description                       |
| -------- | ---------------------------------------------------- | --------------------------------- |
| `GET`    | `/api/v1/account/organization`                       | Who this server signs users in for |
| `GET`    | `/api/v1/account/social-providers`                   | The providers to offer as buttons |
| `GET`    | `/api/v1/account/sso`                                | The SSO connections with a button, and whether any is on |
| `POST`   | `/api/v1/account/sso/discover`                       | The connection an email's domain signs in through |
| `GET`    | `/api/v1/account/me`                                 | The signed-in user                |
| `PATCH`  | `/api/v1/account/me`                                 | Change their name                 |
| `POST`   | `/api/v1/account/password`                           | Change the password; signs out everywhere else |
| `POST`   | `/api/v1/account/email`                              | Start moving to another sign-in address, where the flow allows it |
| `GET`    | `/api/v1/account/sessions`                           | Where they are signed in          |
| `DELETE` | `/api/v1/account/sessions/:id`                       | Sign another device out           |
| `GET`    | `/api/v1/account/connected-applications`             | Apps holding refresh tokens       |
| `DELETE` | `/api/v1/account/connected-applications/:client_id`  | Revoke an app's refresh tokens    |

Sessions are a random token in an HttpOnly cookie; the database keeps only a
SHA-256 hash of it, so a leaked database cannot be signed in with. They last
12 hours. Signing in, failing to sign in, and signing out are all written to
`audit_logs`.

To add an endpoint: write the handler method in the package it belongs to
under `internal/api/`, then mount it in `registerRoutes` in
`internal/api/server.go` — the one place that says which paths exist. Put it
behind `session.Require` unless it is meant to be public.

## Admin panel

The SvelteKit app in `web/console/` is the admin panel.

```sh
make dev     # the API and the apps; the panel is on :5174
```

`make prod` runs the production build instead, behind `web/serve.js`, which
routes `/api/v1/admin` to the admin listener as Caddy does in a deployment.
Every app builds with `adapter-node`; `deploy/docker/web.Dockerfile` packages it.

Then open http://localhost:5174/admin/login. A panel with no administrator
sends you to `/admin/new-super-admin` to make the first one; after that,
signing in leads to `/admin/dashboard`.

The header holds the logo and, right beside it, the control that folds the
sidebar — at the top of the column it folds — then the command palette, links
to the documentation and the source, the theme toggle and the account menu,
where the account panel, the organisation's settings and Sign out live.

`⌘K` (`Ctrl+K`) opens the palette: every page the signed-in administrator may
open, filtered as you type by letters in order rather than a prefix, so `adro`
finds Admin roles. It is built from the same `sections` catalog the sidebar
is, so a page added there needs nothing else to be reachable from it.

Where to browse rather than jump is the dashboard's sidebar, whose column the
header's logo block tops: the two are one width and fold together.

It is a tree. The two pages that answer "what is happening" stand on their own
at the top; everything else is under the subject it belongs to, and a section
opens and closes with a click. Which are closed is a cookie, so the column
arrives as it was left, and the section holding the page being read is named
in full whether or not it is open.

```
Activity · Logs
▾ Applications     Applications · APIs
▾ Authentication   Login flows · Social · SSO integrations · One-time codes
▾ Users            Users · Sessions · Roles
▾ Administration   Administrators · Admin roles          (super admins only)
▾ Settings         Organization · Mail · Languages
```

A section is a subject rather than a bucket, which is why One-time codes sits
under Authentication — where it is read — rather than under Settings, where it
merely lives. `sections.ts` is the one list, and the sidebar and the command
palette both read it.

The page being read is filled and marked down its left edge — on the row
itself, or on the tree's rule where the page is under a section — so the eye
finds it before it reads a word. Folded to icons, a section opens its pages
beside it as a flyout, and the one holding the page being read carries a dot.

Only the list scrolls; the column around it keeps the border and the
background, so the bar is never against the column's edge and opening a
section cannot make the whole column jump.

**Below 55rem the column becomes a panel.** The control beside the logo stops
folding and starts opening: the same rows slide in over a dimmed page and see
themselves out when a page is chosen, on Escape, on a click outside, or when
the window grows enough to hold a column again. It is a media query as well as
a class, so a narrow screen is served a panel that is already off the edge
rather than a column that is taken away once the script runs.

Each link is shown only to an administrator whose roles allow the page. The
logs live at `/admin/dashboard/logs`; the old `/admin/logs` redirects there.
Every page talks to the API.

### Users and their fields

A user record is made of two kinds of field, and the difference is where the
value lives.

**Built-in fields are columns of `users`**: `email`, `email_verified`,
`first_name`, `last_name`, `is_active`. Every installation has them, the
database enforces them, and they can be indexed and searched properly. There
is no table of them — `BuiltinFields` in `internal/model/user_field.go` only
*describes* those columns so the panel can draw them, and a test pins that
description to the columns so the two cannot drift.

**Additional fields are rows of `user_fields`**, with their values in
`users.data`. A super admin adds them in the panel, so keeping a phone number,
a nickname or a joined-on date needs no migration. Their names may not collide
with a built-in one: two fields with one name would be two places to look for
the same thing, so the API refuses it.

`internal/model/user_field.go` owns what a field means for both kinds: the
types on offer (`text`, `number`, `bool`, `email`, `date`), the rules a value
keeps (required, unique, a smallest and largest value, a prefix), and
`Normalise`, which checks a submitted value and returns what should be stored.

The API returns both kinds in one list, built-ins first, each marked
`builtin`. The panel draws its table and its form from that list, so a field
someone adds appears as a column and an input without any code changing.

Searching matches the email or any stored value, because `data` is searched as
text; the search and the verified filter live in the URL, so the server renders
the result and a filtered list can be linked to.

**Your account** opens from the header's account icon as a drawer, the way
every other record does, and is built out of the same `Panel`, `List` and
`Tag` the pages are: the account itself, two-factor sign-in, and the sessions
this account has. The menu's rows open it on the part they name. The theme is
not among them — it is the toggle in the header, and one place to change it is
enough.

Two-factor is managed here, for the reader's own account: setting up an
authenticator, replacing it, new recovery codes, and turning it off where the
policy allows — each of the last three proving the reader holds the phone by
asking the authenticator for a code first. Whether *every* administrator needs
one is not here; that is the policy, and it lives on the Administrators page
along with resetting somebody else's factor.

Changing the email or the password is not built yet; those rows say so and
their buttons are disabled rather than pretending.

`web/console/.env` names the admin API in `API_URL`. The browser never uses it:
it calls `/api/v1/admin/...` on the panel's own origin, and the proxy routes
that to the API. So the session cookie the API sets belongs to the panel's
host, and server-side rendering can read it — on any domain, not only on
localhost.

**Rendering.** Pages are rendered on the server, which is what makes a reload
show the finished page rather than assembling one: the theme, the title and
the content are all in the first response. That means the data has to be
fetched on the server too, so the loads are `+page.server.ts` and
`+layout.server.ts`. They call the API by path with the load's `fetch`, and
`handleFetch` in `hooks.server.ts` (`lib/server/proxy.ts`) sends those calls to
`API_URL` directly, with the reader's cookie — a fetch made by the server
carries none of the browser's cookies on its own.

Signing in and out still happen in the browser, followed by `invalidateAll()`
so the server loads run again with the new session.

**Components.** `lib/components/ui` holds the building blocks — controls,
fields, and the page pieces `PageContainer`, `PageHeader`, `Panel`,
`StatCard`, `List`, `ListItem`, `Thumb`, `Tag` and `Alert`, described in its
README — and each feature folder the pieces only that feature uses. Pages
compose those and never reach for an Ark UI primitive directly, so a change to
how a field looks happens in one file.

**Layout.** Pages with tables fill the width, the way PocketBase does: a table
with room for its columns reads better than one centred in a narrow column.
Pages read top to bottom, such as the profile, sit in `PageContainer`'s centred
column instead. Below 55rem the sidebar becomes a panel over the page rather
than a column beside it.

**Styling.** [Ark UI](https://ark-ui.com) ships no CSS: every part it renders
carries `data-scope` and `data-part`, and `lib/styles/ark.css` styles those
attributes. Everything else refers to the tokens in `tokens.css`. There is no
CSS framework.

The layout is taken from the [PocketBase](https://pocketbase.io) admin UI,
and the colour from Telegram's: a white page in the light theme and a black
one (`#000`) in the dark, cool greys a step apart, and one brand blue, `#0088cc`, for
whatever is chosen, pressed or current — the primary button, the page open in
the sidebar, a ticked box, a switch that is on. Every colour is solid; none is
mixed with transparency.

Fields — inputs, text areas, passwords and selects — are an outlined box
with a floating label inside it, and the hint or error on its own line under
the box. In an empty field the label rests exactly where the value is typed;
with focus or a value it rises to the top of the box at three quarters of its
size, on the same left edge, and the value takes its line. The box has no
fill, so it takes the colour of whatever it sits on, and a hairline border
that darkens under the pointer and turns brand blue while typing. Number
fields have no spinner buttons, and a text area grows with its text up to a
screenful. All of it is in `lib/styles/fields.css`.

What floats over the page is drawn from its own tokens: popovers — menus
and a select's list — from `--popover-*`, dialogs — drawers and the command
palette — from `--dialog-*`. In the light theme both are white, set apart by
a hairline and a shadow. In dark mode they are flat, with no shadow or glow:
a dialog sits at `#090909`, while a popover takes the next step at `#0c0c0c`
so a select or menu opened over a dialog remains distinct. Hairline borders
define both surfaces. A row in any popover is one design: 36px, an icon in the
hint colour, a quiet fill under the pointer, and in a select the chosen option
on a light wash of the brand colour. Everything inside picks its level up
without knowing where it is.

**Theme.** Light or dark, switched by the toggle in the header — one click,
no menu — and remembered in a cookie.

The cookie rather than `localStorage` is the point: `hooks.server.ts` reads it
and writes `data-theme` straight onto the `<html>` tag, so the page arrives
already dark or light. Nothing has to be corrected after the fact, which is
what a flash on refresh actually is. A reader who has not chosen yet gets no
attribute, leaving the `prefers-color-scheme` rules in `tokens.css` to decide
— also before the first paint, because that happens in CSS rather than after
it.

The toggle's icon is chosen in CSS from that same attribute rather than in
JavaScript, so the server renders the right one and the browser has nothing
to correct when it hydrates.

Switching is animated as one cross-fade of the whole page through the View
Transitions API, rather than per-element transitions that each start at a
slightly different moment. Browsers without it simply change. Both that and
the toggle's own icon animation stop at `prefers-reduced-motion`.

**Fonts.** The sign-in app uses Roboto for text and Roboto Mono for code,
bundled from `src/assets/fonts/Roboto` and `src/assets/fonts/Roboto Mono`
under the SIL Open Font License (the `LICENSE` beside them), so every machine
shows the same thing. Each is a variable font — every weight in one file —
split by alphabet: Latin and Cyrillic and their extended sets, with
`unicode-range` in `lib/styles/fonts.css` so a page downloads only what it
shows. They come from [Fontsource](https://fontsource.org)
(`@fontsource-variable/roboto` and `roboto-mono`, 5.3.0); a language in
another alphabet needs that subset's files added the same way.

The console uses Chirp as its primary typeface. Its Latin and Latin Extended
files at 400, 500 and 700 are bundled; the system sans and monospace stacks
cover scripts and code Chirp does not provide. Chirp is X's typeface rather
than an open font; its licensing note is in
`web/console/src/assets/fonts/README.md`.

**Icons** are [Remix Icon](https://remixicon.com), through `svelte-remixicon`.
They are components, so only the ones actually used are bundled — there is no
icon font to download. Pass one to the `Icon` wrapper rather than using it
directly, which keeps sizing and alignment in one place:

```svelte
<Icon icon={RiLogoutBoxRLine} />
```

Give `Icon` a `label` only when the icon carries meaning on its own; beside
text it stays `aria-hidden` so a screen reader does not read the same thing
twice.

### The organisation

Settings · Organization is the installation itself, and a super admin's alone:
what the organisation running it is called, its identifier, its primary
domain, its logo, where users write or ring for help, and the terms and
privacy policy they accept by making an account.

There is one of it, so it is a record of settings rather than a list. The
`organizations` table holds a single row, seeded by the migration from
`model.DefaultOrganization` — which is also what the store writes if it ever
finds none, so a fresh installation and a repaired one start the same. The two
endpoints read it and write it back; nothing creates or deletes one.

It is not a page of notes: everything on it is used.

| Setting                | Where it shows                                                        |
| ---------------------- | --------------------------------------------------------------------- |
| Terms, privacy policy  | `op_tos_uri` and `op_policy_uri` in the discovery document; linked under every sign-in card, and agreed to when registering |
| Name, logo             | The sign-in card, for an application that carries neither             |
| Support email, number  | The foot of every sign-in page: whom to ask when signing in fails      |
| Identifier, domain     | The panel, as what the organisation is called where a name will not do |

An application's own links, name and logo come first where it has them
(`internal/oidc/logout.go`), and each falls back on its own: an application
with terms but no privacy policy shows its terms and the organisation's
policy. The sign-in pages read all of it from `GET
/api/v1/account/organization`, which takes no session — the pages are shown to
people who have not signed in yet — and publishes only what a stranger may
see.

`PATCH /api/v1/admin/organization` is a true PATCH: a setting the request
leaves out keeps the value it has, and an empty string clears one that may be
empty. What a name, a domain, an address, a number or a link may be is
`model.Organization.Validate`, in one place because it is one question; a
refused change writes nothing. Each change is recorded as
`organization.updated`, and the log line says which settings moved.

Reading the page takes `organization.read` and changing it
`organization.write`. Neither is scopable — these are the whole
installation's settings, not one application's — and no seeded role grants
them, so a new installation shows the page to super admins until a role hands
it out.

### Signing in with another account


Authentication · Social is where an administrator registers this server with
Google, Apple, Facebook, Yandex ID, VK, or any other OAuth 2.0 or OpenID
Connect provider. Each one becomes a button beside the password form, on the
sign-in and registration pages.

This server is the client in that exchange, not the provider. `/oauth2/social/
<slug>/start` sends the browser to the provider with `state` and a PKCE
challenge, and leaves the same `state` with the browser in a short-lived
cookie of its own; `/oauth2/social/<slug>/callback` — which answers POST as
well, because Apple posts its answer — refuses an answer that does not come
back in the browser that set off, then trades the code for a token, reads the
identity, and starts the same session a password would have started. From
there everything is the same: the sign-in under way continues back to the
application that asked, and a person with no application waiting lands on
their account page.

```
internal/model/social.go       the records, and SocialSpecs: what each kind's
                               endpoints, scopes and claims are
internal/oidc/social.go        the exchange, and who it signs in
internal/api/social/           the panel's endpoints
```

**What is stored.** `social_providers` is one row per provider, with the
client secret — or, for Apple, the .p8 signing key — encrypted with the
server's secret key, as the token signing keys are. Neither ever leaves the
server: the panel is told only whether one is stored. `user_identities` is one
row per account at a provider, keyed by the provider's own subject, which is
the only thing that identifies somebody reliably. `social_logins` is a sign-in
that has gone to a provider and not come back: its `state` hashed, its PKCE
verifier, and where the person was going, single use and short-lived. The
other half of the `state` is in the browser rather than the database — the
`xermess_sign_in_state` cookie — because a state anyone can replay in any
browser would let whoever finished a sign-in of their own leave somebody else
signed in as them.

**What is not stored** is what every installation would have to keep the same:
where Google's endpoints are, what Yandex calls an address, that VK sends the
address with the token and puts the profile under `response.0`. That is
`model.SocialSpecs`, one entry per kind, so a provider that moves an endpoint
is a change to that file rather than to everybody's database. The two custom
kinds — `oidc` and `oauth2` — have no entry to make: the record carries their
endpoints, and the panel asks for them.

**Which account a sign-in reaches**, in order: the one that already holds this
identity; then the one with the same address, if the provider says it has
verified it *and* the provider is one this installation trusts to say so
(`link_verified_emails`) *and* the account verified its address too — else
whoever registered it before its owner arrived would keep a password to it;
then a new account, if the provider and the
application both take registrations. An address that is taken and unproved is
refused with a sentence saying to sign in with a password and connect the
provider from the account page — otherwise anybody who could make an account
at a provider with somebody else's address could take over theirs here.

**An id_token that came straight from the token endpoint is not checked for a
signature.** It arrived over TLS from the provider's own endpoint, in answer
to a request carrying this client's secret, which OpenID Connect Core section
3.1.3.7 accepts in place of checking it. What is checked is that it names this
client and has not expired.

Reading the page takes `social.read` and changing it `social.write`. The
second is worth guarding: whoever holds it decides which accounts elsewhere
reach this server.

`PATCH /api/v1/admin/social-providers/:id` is a true PATCH, as the
organisation's is: a request that does not mention a setting leaves it as it
is. That is what lets the panel's list turn a provider off with
`{"enabled": false}` and nothing else, without emptying the endpoints and keys
around it. The kind and the identifier are read when a provider is registered
and never again — both are in the address registered with the provider.

**The secrets can be read back.** They are encrypted rather than hashed —
this server has to send them to the provider — so the panel offers to reveal
one, for checking against the provider's console. It takes `social.write`, the
permission that could replace the secret anyway, and every reading is written
to the log as `social_provider.secret_read` with who asked.

**A user's record says which providers they sign in with.** The users table
marks each row with what gets that person in — a password, an account
somewhere else, or both — and their record lists the providers with the
address each gave and when it was last used. A provider can be disconnected
there (`users.write`), which leaves the account itself alone: it is one way in
that goes, not the person.

### SSO integrations

An organisation signs its people in through its own identity provider —
Okta, Microsoft Entra ID, Google Workspace, ADFS, Keycloak, OneLogin — over
OpenID Connect or SAML 2.0. It is what authentik calls a *source*, Keycloak an
*identity provider* and Auth0 an *enterprise connection*; where
[Social](#signing-in-with-another-account) offers "Continue with Google" to
anybody, a connection is set up for one organisation, and **owns email
domains**.

**What a connection decides**, in the order a sign-in meets it:

| Setting | What it does |
| --- | --- |
| Domains | The addresses it signs people in for. It signs in nobody else: an address the provider vouches for outside them is refused (`sso_domain_mismatch`), so a misconfigured provider cannot sign in as anyone elsewhere. A domain belongs to one connection. Domains are optional: without any, the provider is trusted with every address, as authentik and Keycloak trust a source, and since no address leads to it, it has to show its button and cannot be required (`sso_unreachable`). |
| Require SSO | Makes it the only way in for its domains: their password sign-in, registration and reset are refused with `sso_required`, which names the connection, and the sign-in page sends them there with the address as `login_hint`. |
| Button on the sign-in page | "Continue with *name*". Off, people reach it through **Sign in with SSO**, which finds the connection from their work address. |
| Existing accounts | *Link* signs an address that already has an account in to it — the provider owns the domain, so it is the same person — or *Refuse* (`sso_link_refused`). An account that never verified its address is not linked, since whoever registered it would keep its password. Pointing a connection at another provider (a new issuer, or metadata with a new entity ID) forgets its identities, because a subject is only unique within its provider; the accounts link again on their next sign-in. authentik's email_link and email_deny. Without domains, *Link* trusts the provider with every account there is, so choose it only for a provider you would trust with them. |
| Create accounts | Just-in-time provisioning: somebody new gets an account on first sign-in, verified, with the default roles. Off, only existing accounts sign in (`sso_no_account`). |
| Update names | The provider is where names are kept, so each sign-in brings a change there here. |
| Attributes | Where to read the address, the names and the groups. Empty is the usual ones: OIDC's standard claims, and for SAML the names Entra ID, ADFS, Okta, Google Workspace and the LDAP OIDs use. |
| Group to role mapping | Everybody the provider puts in a group gets the role (global or an application's). *Take mapped roles away too* removes a mapped role from someone no longer in its group; off, mapped roles are only added. |

**What is checked**, because the provider's word is what signs somebody in:

- **OpenID Connect**: the issuer's discovery document has to be its own; the
  code is bound to the sign-in by PKCE and the state; the id_token's
  signature is verified against the provider's published keys (refetched once
  when a token names a key not seen yet), with its issuer, audience, expiry
  and nonce. An address the provider says outright it has not verified is not
  used.
- **SAML 2.0** (crewjam/saml): the response has to be signed by the
  certificate in the provider's metadata, be meant for this service provider,
  be in time, and answer the request this server sent. An unsolicited,
  IdP-initiated response is refused, since nothing ties it to the browser
  presenting it, and a response is only accepted once. Each connection has its
  own RSA key and self-signed certificate, made when it is, which requests
  are signed with when the provider requires it.
- **Both**: the address has to be at one of the connection's own domains, and
  the answer has to come back in the browser that started the sign-in — the
  `state` is left there in a cookie as well as stored, so a finished sign-in
  cannot be walked through somebody else's browser.

**Setting one up** is the SSO integrations page. A connection starts off; the
drawer's **Test connection** reads the issuer's discovery or the metadata
before anything is saved — and warns about a scope the provider does not
list, which some providers, Keycloak among them, refuse a whole sign-in over
— and once created its **Service provider** tab has
what to give the provider — the redirect URI for OIDC, and for SAML the ACS
URL, the entity ID and a metadata URL (`/oauth2/sso/<slug>/metadata`) most
providers can be set up from in one paste. A SAML provider's metadata can be
read from its address and read again with **Refresh metadata** when it rolls
its certificate over; the drawer warns a month before that certificate
expires. It takes `sso.read` to see the page and `sso.write` to change it, and
every change is in the activity log, as is a role a sign-in added or took
away (`user.roles_synced`). So is a sign-in that did not complete
(`sso_connection.sign_in_failed`), with the step — authorize, token,
id_token, response, or claims — and exactly why: what the provider said,
such as `invalid_scope` or `invalid_client`, or what it sent instead of a
verified address at the connection's domains, naming the claims there were.
The person is only told the reason in general (`sso_upstream`,
`sso_no_email`, `sso_domain_mismatch`), and that entry is how a connection
being set up gets fixed. A refusal is recorded
only for a sign-in this server started, so a made-up callback link writes
nothing.

The public paths, under the issuer, are `/oauth2/sso/<slug>/start`,
`/callback` (OIDC), `/acs` (SAML, a POST from the provider, outside the CSRF
check for the reason Apple's callback is) and `/metadata`.

### Login flows

Authentication · Login flows is where an administrator writes what a sign-in
asks for: an ordered list of steps, and the things a person is allowed to do
along the way. There are as many flows as somebody has written, one of them is
the default, and an application either names a flow of its own or falls back
to that one.

```
Identify → Password → Other accounts         the default: everything else
Identify → Other accounts                    a shop with no passwords
Identify → Password  (verified, 8 h)         the staff tools
```

A flow carries the steps, whether an account can be made, whether a password
can be reset, whether an address has to be confirmed first, and how long a
session it makes lasts. The steps come from a catalog in
`internal/model/login_flow.go` — the model owns what a step is, what it is
called and what it does, so adding one is an entry there and the panel's
picker follows.

**The editor is a canvas.** A flow opens full-page at
`/dashboard/flows/<id>`, drawn with [Svelte Flow](https://svelteflow.dev): the
sign-in starting at the top, each step as a card, and the session it ends in
at the bottom. The steps that can be added sit in a palette beside it — drag
one onto the flow, click it, or press **+** between two steps and pick one.
Dragging a step reorders it; the first step, which asks who is signing in,
stays first. Selecting a node shows its settings on the other side, and each
setting sits on the node it governs: making an account and requiring a
verified address on Identify, the reset link on Password, the session's
length on "Signed in", the flow's name and whether it is on or the default on
its start. What would stop it saving — no name, no step that lets anybody in
— is said while it is drawn, not after Save, and leaving with changes unsaved
asks first. A new flow starts from a template (Password, Other accounts only,
Staff, Blank); any flow can be duplicated, exported as JSON, and a file
exported here or from another installation imported as a new flow.

**Every way in follows the flow**, whichever application it is for
(`oidc.flowFor`) — a password, a provider, an organisation's identity
provider:

- **Sign-ins are open**, off, refuses every way in (`sign_in_closed`) — a
  password, a provider, an organisation's identity provider, an emailed code,
  and registering, which ends in a session like the rest. It is checked in
  `oidc.startSession`, where they all end, so nothing gets in by a road
  somebody forgot to close. Sessions already made are left alone: closing the
  door does not turn anybody out. The sign-in page draws a closed card rather
  than a form nobody could use. It is not **On**, which says whether an
  application may be pointed at this flow at all;
- a flow without **Password** refuses a password before looking at it
  (`password_not_offered`), makes no accounts with one, and sends no reset
  links; the sign-in page shows the provider buttons alone;
- a flow without **Other accounts** refuses a provider (`social_not_offered`);
- **Confirm the address of a new account** emails a link when somebody signs
  up, and lets them in anyway: the link is waiting for them. **Require a
  verified address** is the other half of that subject and a different
  decision — it sends an account whose address is unconfirmed a link instead
  of a session (`email_not_verified`). A flow can do either, both or neither.
  The link opens `/verify-email`, where a button — not the link itself, which
  a mail scanner would use up — confirms it
  (`POST /api/v1/account/verify-email`). Completing a password reset confirms
  the address too;
- **Offer "Stay signed in"** puts a box beside the password. Ticked, the
  browser keeps the cookie for as long as the session lasts; left alone — or
  not offered at all — the cookie carries no `Max-Age` and the browser drops
  it when its window closes, so a machine somebody was passing through
  forgets them. The session itself lasts as long as the flow says either way.
  A sign-in through a provider has no box to tick and is remembered;
- **People can change their email** offers "Change" beside the address on a
  user's own account page (`POST /api/v1/account/email`). The link goes to
  the address they typed, and the account only moves when it is opened — so
  nobody takes an account by typing an address they cannot read, and nobody
  loses one by misspelling it. It answers the same whether or not somebody
  else already has that address, and a clash is caught when the link is used,
  by which point the person holding it has proved they read that inbox;
- **Emailed code** holds the sign-in once the password has been accepted,
  emails a one-time code to the address on the account, and makes the session
  only when that code is typed back (`POST /api/v1/account/login/code`). What
  a code is — its length, how long it lasts, how many guesses it takes, how
  soon another may be asked for — is the One-time codes page;
- the session lasts as long as the flow says, and so does its cookie.

The emailed code is the address being proved, not the password being doubted,
so the ways in a provider has already proved it for — a social sign-in, an
organisation's identity provider — go straight through. The handle the page
holds is the secret: a six-digit code read over somebody's shoulder is no use
without the browser that asked for it.

The email is written in the language the pages were shown in, read from their
cookie, so a provider's callback — which has no body to say it in — gets it
right too.

**Not every step in the catalog runs yet, and the editor says so.**
`totp`, `terms` and `consent` can be placed as a plan, drawn
dashed and marked "not run yet"; `LoginStepSpec.Implemented` is the one place
that says which is which, and implementing a step is flipping it there and
writing the step. A flow has to include Password or Other accounts, which do
run, so nothing the panel saves can lock everybody out.

The default flow is what everything falls back to, so it cannot be turned off
or removed; making another flow the default takes the mark from it. A flow
that is turned off keeps its applications — they fall back to the default —
and removing one clears the column rather than taking its applications with
it. Reading the page takes `login_flows.read` and writing takes
`login_flows.write`.

### Sessions

User management · Sessions is everyone signed in right now — Keycloak's
Sessions page: each browser session with its user, device, address, when it
started and when it runs out, newest first. Search by the start of an
address, or open one user's sessions from their name or from the Sessions
link in their drawer.

**Sign out** ends one session: that browser has to sign in again, and the
user's applications keep their tokens, as when the user ends a session
themselves on their Security page. **Sign out everywhere** is for a lost
device or a compromised account: every session the user has ends and every
refresh token their applications hold is revoked, in one transaction, so no
application stays signed in until its tokens run out. Both are in the
activity log (`user.session_ended`, `user.signed_out_everywhere`).

A session is part of a user's account, so reading the page takes
`users.read` and signing anybody out takes `users.write`.

It is built for millions of sessions. The list is never counted and never
paged by offset: it reads newest first by an index on `(created_at, id)` and
continues after the last session shown, so the thousandth page costs what
the first does, and the search is a prefix match on the address so it can
use an index (`text_pattern_ops`) rather than read every user.

This page replaced a raw browser over the server's own tables. Keycloak and
authentik have nothing like one, and for good reason: it showed rows rather
than what they mean, and it was the one page from which every account and
every application could be read.

### Languages

Every word the sign-in pages show is looked up by key — `t('login.title')` —
through the [`svelte-i18n`](https://github.com/kaisermann/svelte-i18n)
dictionary and formatter. The text for each language lives in the database: a
`languages` row with its names and settings, and a `translations` row per app
holding one JSON object of text by key. **Settings · Languages** is where languages are added,
translated, offered, made the default and removed, and a change is on the
next page anybody opens: the pages ask the API for their text while
rendering, so nothing is rebuilt.

The admin panel is not one of these apps. It is English, in its own markup —
see [below](#the-admin-panel-is-not-translated).

**The files under `i18n/` are what the server ships with:**

```
i18n/
  id/
    en/       common.json  auth.json  account.json  server.json
              validation.json  email.json
    ru/       common.json  auth.json  account.json  server.json
              validation.json  email.json
  console/
    en/       server.json  validation.json
```

The sign-in catalogs are grouped by subject: `common` holds shared labels and
controls, `auth` the sign-in and recovery screens, `account` the user's own
account, `server` the server's problems, `validation` form and request
validation, and `email` the messages the server sends. The console files are
not a translation: they hold the English sentence for each problem the admin
API answers the panel with.

**Each group is nested by namespace**, so a translator reads related text
together. The Go importer merges every group in a language directory into one
catalog, and everything that looks text up — the apps' `t()`, the database,
the editor — speaks of the dotted key:

```json
{
  "$name": "Russian",
  "$native": "Русский",
  "login": { "title": "Вход", "subtitle_app": "чтобы продолжить в {app}" },
  "error": { "invalid_credentials": "Неверный адрес почты или пароль." },
  "email": { "reset": { "subject": "Сброс пароля" } }
}
```

`login.title` is `t('login.title')`. Keys are lower case with underscores;
`TestShippedGroupsAreNested` holds the files to that. The same few namespaces
recur: one per screen or feature (`login`, `register`, `security`…), `field`
for form labels, `action` for buttons, `error` for everything the server can
refuse (see [Errors](#errors)), and `email` for what it sends. The editor's
**Export JSON** writes this shape and **Import JSON** reads it, or a flat
file of dotted keys.

They are embedded in the binary (`i18n/i18n.go`) and have three jobs:

- **The first start imports them.** `store.EnsureLanguages` runs in `main`: on
  a database with no text at all it writes every shipped language, English on
  and the rest off. From then on the database is the panel's.
- **A release reaches a shipped language.** Each start copies a key a shipped
  group has into the database's copy of that language when the copy lacks it,
  and never touches a message that is there. The one consequence: a message
  cleared in the panel, in a language that ships, comes back on the next
  start — a shipped language can be reworded, and the way to empty one is to
  remove it. A removed language is never brought back by a start; the New
  language drawer offers it back instead.
- **`i18n/id/en/` is the contract.** The keys in its groups are the keys there
  are: coverage is counted against them, a key they do not have is dropped on
  save, and a key a language has no text for is sent in English — the
  database's English, and under that the shipped English groups, so a page
  never shows a bare key.

**Adding a language** is **New language** on the page. Typing a tag — `uz`,
`pt-BR` — fills in both names from the browser's own list, and the language
starts from nothing, from a copy of another one here, or from the shipped
translation if the server has one. It starts off, so it can be translated
before anybody is offered it, and opens on its first untranslated app.

**Translating** is the language's drawer: a tab per app, every key with the
English beside a field for the translation, a search over keys and both
texts, and a switch for only what is left. A translation that drops a
`{parameter}` the English has is flagged under the field. **Export JSON**
writes a file of every key — empty values for what is left to do — with
`$name` and `$native` on top, which is the shape of a merged catalog, so it
can go to a translator and come back through **Import JSON**. An import is
merged over what is there: keys it has text for are replaced, the rest are
kept, and keys this version does not use are skipped and counted.

The drawer's **Settings** tab renames a language, offers it on the sign-in
pages, makes it the default — which takes the mark from whichever language
had it — and sets its place in the picker. The default is always offered; the
base language, English, is always offered and cannot be removed; and the
default cannot be removed until another language is. Reading the page takes
`languages.read`; everything that changes it takes `languages.write`. Each
change is in the activity log.

**On the sign-in pages** the picker is in the top corner, beside the theme
toggle, and names each language in itself with its English name under it.
Choosing one redraws the page where it is — the root layout asks for the new
text and every component re-renders, so a half-typed address survives — with
a cross-fade where the browser has view transitions. Without JavaScript the
entries are links to `?lang=ru`, remembered in a cookie and redirected away,
so a shared address carries nobody's choice. Which language somebody gets is
their saved choice, then their browser's `Accept-Language` (by exact tag,
then by its language part, so `ru-RU` is served `ru`), then the
installation's default — and only ever one that is offered, so a choice of a
language since turned off falls through. `GET /api/v1/account/languages` is
the list and `GET /api/v1/account/languages/:code` one language's text, every
key filled in; a language that is off is a 404 there.

<a id="the-admin-panel-is-not-translated"></a>

**The admin panel is not translated.** It is written in English, in its own
markup, and there is no picker, no cookie and no catalog behind it: a
language an installation adds is a sign-in language, and that is the only
kind there is. The Languages page counts and edits one app, and
`i18n.Apps` is the list — a second translatable app would be added there
rather than assumed here.

The one thing the panel still reads from `i18n/` is not a translation.
`i18n/console/en/` holds the English sentence for every problem the
admin API can answer the panel with, keyed by the code beside it
(`respond.Define(..., respond.Admin)`); the panel shows the sentence the API
sent. Nothing imports those files into the database, offers them for
translation or counts them, and a start removes any panel text an older
version imported.

The sign-in pages resolve the language and fetch their text while rendering
on the server, so the first response is already translated and `<html lang>`
is right in the first byte, which is what a screen reader reads the page's
words with. They bundle English alone, for when the API cannot be reached.

Two sign-in languages ship, English and Russian; any other is added on the
Languages page. `TestTranslationsAreComplete` fails the build if a shipped
language falls behind the base, so a key added without a translation is
caught rather than shipped. The Russian was written alongside the machinery
and would be worth a native speaker's eye before an installation offers it.

### Mail

How this installation sends email, and what each message says. Both are a
super admin's: the settings carry the mail server's password, and the words
are what lands in a user's inbox.

**The mail server** is one record, like the organisation's. It starts as
`XERMESS_SMTP_*` says on the first start and belongs to the panel afterwards,
and it is read when a message is sent rather than held in a field from
startup — so a corrected password takes effect on the next email instead of
the next restart. Host, port and one of three encryptions: `starttls` (what
port 587 expects), `tls` (from the first byte, what 465 expects) and `none`
(a relay on the same machine). The password is sealed with
`XERMESS_SECRET_KEY`, like a social provider's client secret, and nothing
reads it back: the panel is told whether one is stored and no more, so a form
that leaves the field empty keeps it and an empty string clears it.

With sending switched off, every message is written to the log, body
included. That is the default, and it is why a fresh installation can be
tried without a mail server rather than failing every sign-up on a host that
does not answer.

**Send a test message** posts one with what is on the form, whether or not it
has been saved, so a server can be tried before anybody's sign-in depends on
it. What the mail server said when it refused comes back as the error's
`reason` — finding that out is the whole point — and both the attempt and its
outcome are in the activity log.

**The words** are not a record of their own. An email goes out in the
reader's language, so its subject and body are two keys of the sign-in pages'
text, in the `languages` table the Languages page holds; the Mail page shows
those keys as the messages they make, with the shipped English underneath as
the placeholder. A field left empty stores nothing, so clearing one puts the
shipped text back. `model.MailMessageSpecs` is the catalog — an email added to
the server is added there, or the page will not know it exists — and a write
that names any other key is refused rather than quietly dropped.

### One-time codes

The codes the server emails as people sign in: the **Emailed code** step of a
login flow, and nothing else. How long a code is (4 to 10 digits), how long
it lasts (up to an hour), how many wrong guesses one sign-in takes, and how
long before another message may be asked for. The page says how many codes
there are against how many guesses there are, because that is the number
these settings actually decide, and it lists the flows that ask for a code —
or says that none does, since settings nothing reads are a note rather than a
setting.

The codes an authenticator app shows are not these. Those are RFC 6238
(`internal/totp`), whose parameters every app assumes, and they are fixed in
code rather than configurable: an app that read a different value from the
URI would show the wrong codes.

A sign-in waiting for a code is a row in `login_codes`, swept with everything
else that expires. Asking for a code forgets the one before it, so one
message is live at a time; a code that is right is spent by a conditional
update, so two requests typing it at the same moment make one session; and
the guesses already spent stay spent when another message is asked for, so
"send it again" is not a way to start the count over.

### The icon

There is one icon file per app, `static/favicon.svg`, and every page names it
in `app.html` — in the first byte, rather than importing it in a layout, so it
is there before the app boots and on the error pages too.

A path that answers with something other than a page has no head to name it
in, and the browser asks for `/favicon.ico` instead: the provider's JSON at
`/.well-known/openid-configuration`, `/oauth2/*` and the account API all answer
on the app's own origin, so before there was anything there they showed no
icon at all. That path is a route now
(`src/routes/favicon.ico/+server.ts`), and it redirects to the same SVG.

It is a redirect rather than a second file because of the type. Both static
handlers — Vite's in development, the Node adapter's in production — look one
up through mrmime, which has no entry for `.ico`, and the edge sets
`X-Content-Type-Options: nosniff` (`deploy/Caddyfile`), so an icon served
without a type would be refused rather than guessed at. Pointing at the file
in `static/`, which is served as `image/svg+xml`, keeps one icon to change and
one type to be right. A browser too old for SVG icons — Safari before 17 —
shows none, as it did before.

## Database

Migrations are Go files in `migrations/`, run by
[goose](https://github.com/pressly/goose). The server applies pending ones on
start unless `XERMESS_DB_MIGRATE=false`.

`20260917150000_schema.go` is the first: it builds the whole schema from
`model.All()`, adds the indexes a struct tag cannot describe, and seeds the
admin roles, the organisation, the default login flow and the base language.
A fresh database gets its tables from the models, so one file cannot disagree
with them; a database that already exists is stepped forward instead, which
is what every migration after it does — the newest,
`20260923180000_login_management.go`, adds the switches a login flow carries
and the address a verification link may be a change to.

```sh
make migrate-new name=add_admin_phone   # create an empty migration
make migrate-up                         # apply pending
make migrate-status                     # what is applied
make migrate-down                       # roll the newest one back
```

A migration registers itself with goose and gets the transaction goose opened.
`gormTx` wraps that transaction in GORM, so a migration can work with the model
structs instead of writing DDL by hand:

```go
func init() {
	goose.AddMigrationContext(upAddAdminPhone, downAddAdminPhone)
}

func upAddAdminPhone(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}
	return db.Migrator().AddColumn(&model.AdminUser{}, "Phone")
}

func downAddAdminPhone(_ context.Context, tx *sql.Tx) error {
	db, err := gormTx(tx)
	if err != nil {
		return err
	}
	return db.Migrator().DropColumn(&model.AdminUser{}, "Phone")
}
```

Everything a migration does runs inside goose's transaction, so one that fails
half way leaves nothing behind. Write the down function even when you think you
will not need it: it is what makes a bad deploy recoverable.

Because the migrations are Go, the `goose` command-line tool cannot run them —
only a binary that imports them can. That is what `cmd/migrate` is for, and
what the make targets above use.

Never edit a migration that has already run anywhere. Add a new one. (The
schema migration is the exception only while no database is running it: a
rebuilt development database is not "has already run".)

### The first administrator

The migration creates no administrator, and no password exists in any file.

A fresh database has roles but nobody to sign in as, and the API says so:
`GET /api/v1/admin/setup` answers `{"required": true}`. The sign-in page asks
before it draws itself and sends whoever is there to `/admin/new-super-admin`,
which takes a first name, a last name, an email and a password of at least ten
characters, makes a `super_admin` with them, and signs that person in.

`POST /api/v1/admin/setup` is the one open route that writes anything, so it
is worth knowing why that is safe: the check for existing administrators and
the insert are one transaction holding a lock, so two requests at once cannot
both get in, and it refuses with a 409 the moment there is one. It is a door
that closes behind the first person through it.

Signing in is refused for 15 minutes after five wrong passwords in a row; a
super admin setting a new password for the account unlocks it. An unknown
username takes as long to refuse as a wrong password, so the answer does not
say which usernames exist.

The address is the account — it is both the email and the username someone
signs in with, so setup asks for one thing rather than two.

## Redis

Redis does two jobs, and neither is ever the only copy of anything.

**It caches the reads every page makes.** Each sign-in page renders with the
organisation, the sign-in buttons, the login flow, the offered languages and
the whole text of one language; each panel page with the panel's languages
and text. None of that changes more than a few times a week, so the store
reads it through Redis (`internal/cache`) and the database is asked once per
change rather than once per page. The largest value is a language's resolved
text — every key of an app, gaps filled from English — kept as one JSON value
per language and app.

Only these are cached — nothing about users, sessions, tokens or
administrators, where a stale answer would be a security question:

| Group            | What                                                      |
| ---------------- | --------------------------------------------------------- |
| `languages`      | every language, the offered ones, the default, each text  |
| `organization`   | the organisation's settings                               |
| `login_flows`    | the default flow, and each flow by id                     |
| `social_buttons` | the enabled providers as buttons — never the providers themselves, which carry sealed secrets |

**Every key says what it is**, under the prefix and then what it is for, so a
Redis browser shows a tree and one group is one `--scan --pattern`:

```
xermess:cache:<group>:generation              the group's current generation
xermess:cache:<group>:v<generation>:<entry>   one cached value, as JSON
xermess:ratelimit:<scope>:<address>           one rate-limit bucket (scope: public, admin)
```

The entries are named for what they hold — `languages` has `all`,
`offered`, `default`, `code:<tag>` and `text:<tag>:<app>`; `organization`
has `settings`; `login_flows` has `default` and `id:<uuid>`;
`social_buttons` has `enabled`.

A write forgets its whole group once it has committed, by moving the group on
to its next *generation*: forgetting is one `INCR` of the group's
`generation` key. That is why a
change is on the next page every server process renders, and why a reader
that loaded a row just before a write cannot put it back after — it stores it
under the generation that was forgotten, which nobody reads again. Old
generations expire within the hour. A forget that could not reach Redis is
remembered, and that process reads nothing from the cache until it has made
it. `TestCachedTypesSurviveJSON` fails if a cached model gains a field JSON
would leave out.

**It holds the rate limit's counts** — a token bucket per address in a Lua
script, so two processes cannot both spend the last attempt, timed by Redis's
own clock.

Redis going away is never an outage. At startup a configured Redis that does
not answer stops the server, like a database that does not, so a wrong
address is caught at once; with `XERMESS_REDIS_HOST` empty the server runs
without one. After startup, a Redis that stops answering is logged once and
then left alone for five seconds at a time: every read goes straight to the
database and the rate limit counts in each process's memory, so pages stay
as fast as they were before Redis rather than each waiting on a dial that
will fail. Every five seconds one request tries Redis again, and the first
that reaches it makes any forget a save left pending.

Run it without persistence, as the compose file does. It holds nothing that
is not in Postgres, and a Redis restored from an old snapshot would bring
back old generations — and the text in them — until they expire.

The settings are `XERMESS_REDIS_HOST`, `_PORT`, `_USERNAME`, `_PASSWORD`,
`_DB` and `_PREFIX` (default `xermess:`, so one Redis can serve several
installations). `make setup` adds them to a `.env` that predates them. In
production, `deploy/compose.yaml` runs one alongside Postgres with no
persistence, a 256 MB cap and `volatile-lru`, which evicts cached values and
never the generation counters that say which are current.

`make test-integration` runs the whole API suite with the Redis in `.env`
under a throwaway prefix per test, so every end-to-end test also proves a
write is seen through the cache; the `TestLive…` tests in `internal/cache`
and `internal/api/ratelimit` need it too.

## Configuration

Every setting is an environment variable, read from `.env` first; real
environment variables win. `.env.example` lists all of them. `XERMESS_DB_DSN`
has no default, so a missing one stops the server.

`XERMESS_SECRET_KEY` has no default either: it encrypts the signing keys, so
`make setup` writes a random one into a new `.env`. Keep it: a different key
cannot read the stored signing keys, and the server will not start.
`XERMESS_ISSUER` is the public URL tokens name the server by — the id app's
origin, which routes the provider paths to the API — and `XERMESS_ACCOUNT_URL`
defaults to it. `XERMESS_ADMIN_URL` is the console's; only that origin may change
anything through the admin API.

`XERMESS_SMTP_*` seeds the mail server on the first start and nothing after
it: from then on it is the Mail page's, so a password typed wrong is
corrected in the panel rather than in a file, and takes effect on the next
email rather than the next restart. With no `XERMESS_SMTP_HOST`, sending
starts switched off and every message is written to the log — enough to
follow a reset link while developing.

Session cookies are Secure when `XERMESS_ACCOUNT_URL` and `XERMESS_ADMIN_URL`
are https, without a setting to forget. `XERMESS_TRUSTED_PROXIES` lists the
reverse proxies whose `X-Forwarded-For` is believed. With none listed, the
address recorded for every request is the connection's own, so a caller cannot
write a made-up one into the activity log — but behind a proxy it has to be
set, or every request, and the rate limit, count as the proxy's.
`XERMESS_ADMIN_ADDR` must differ from `XERMESS_ADDR`; never publish it.

**At scale.** Nothing the server keeps grows without end. Every hour each
process sweeps what has expired — sign-in requests and codes, refresh
tokens, sessions, reset links, abandoned social and SSO sign-ins — in batches
of 5,000, by an index on `expires_at`, so the sweep never holds up the
sign-ins writing the same tables (`store.Sweep`). The activity log keeps
`XERMESS_AUDIT_RETENTION_DAYS` (365; 0 keeps everything). Addresses are
stored lower case, so signing in finds a user by the unique index whatever
capitals were typed, and `Ada@x` cannot be a second account beside `ada@x`.
`XERMESS_DB_MAX_CONNS` (25) caps each process's connections: all the API
processes together have to stay under Postgres's `max_connections`.

`make test-integration` runs the tests that need Postgres and Redis. They
connect to the Postgres in `.env` only to create a database of their own for
each test, and drop it afterwards, and use the Redis in `.env` under a key
prefix of their own; without `XERMESS_TEST_DB_DSN` set, `go test` skips them,
and without `XERMESS_TEST_REDIS` they run with no cache.

The migrations are compiled into the binary. Goose still needs the directory in
`XERMESS_DB_MIGRATE_DIR` to exist; when it holds no `.go` files, as in the
container image, every compiled-in migration is applied.

## License

[MIT](LICENSE)
