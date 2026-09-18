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

locales/                       the translations the server ships with, imported into the
                               database on its first start; en.json is the list of keys

internal/config/config.go      reads .env
internal/database/database.go  opens the connection
internal/database/migrate.go   applies migrations
internal/store/                every query in the project, one file per subject
internal/auth/auth.go          signs administrators in and out, records what they do
internal/oidc/                 the OAuth 2.0 / OpenID Connect provider: authorize, tokens,
                               userinfo, logout, sessions, registration, password resets,
                               and signing in with an account somewhere else
internal/jose/                 signing and checking JWTs, publishing keys, sealing them
internal/mail/                 sending email over SMTP, or into the log
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
internal/api/database/         the server's own tables, read only
internal/api/languages/        the languages, their text, and the panel's own
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
web/console/src/lib/components/profile/  the account: its sessions and signing out
web/console/src/lib/state/             what the panel remembers: the theme, the sidebar's width
web/console/src/lib/utils/             how values are shown
web/console/src/lib/data/demo.ts       placeholder rows for the sections with no backend
web/console/src/lib/server/api.ts      calling the API from a server load, with the session
web/console/src/lib/constants.ts       the names both sides agree on: the cookies
web/console/src/lib/styles/            fonts.css, tokens.css, base.css, ark.css
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
internal/store/database.go     the read-only browser over the server's own tables
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

Every cookie-authenticated API takes changes only from its app's origin
(`internal/api/csrf`): another origin gets 403, a body that is not JSON 415.
Sign-in, registration, setup and password resets are rate limited per address
(`internal/api/ratelimit`).

## API

| Method | Path                          | Needs a session | Description                    |
| ------ | ----------------------------- | --------------- | ------------------------------ |
| `GET`  | `/healthz`                    | no              | The server is up               |
| `POST` | `/api/v1/admin/auth/login`    | no              | Sign in, sets the session cookie |
| `POST` | `/api/v1/admin/auth/logout`   | yes             | Sign out, revokes the session  |
| `GET`  | `/api/v1/admin/me`            | yes             | The signed-in administrator    |
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
| `GET`  | `/api/v1/admin/login-flows`   | yes             | The login flows, and the steps one can be made of |
| `POST` | `/api/v1/admin/login-flows`   | yes             | Write a flow                   |
| `PATCH`| `/api/v1/admin/login-flows/:id` | yes           | Change a flow                  |
| `DELETE`| `/api/v1/admin/login-flows/:id` | yes          | Remove a flow                  |
| `GET`  | `/api/v1/admin/languages`     | yes             | The languages, and how much of each is translated |
| `POST` | `/api/v1/admin/languages`     | yes             | Add one, empty or copied from another |
| `PATCH`| `/api/v1/admin/languages/:code` | yes           | Rename it, offer it, or make it the default |
| `DELETE`| `/api/v1/admin/languages/:code` | yes          | Remove it and its text         |
| `GET`  | `/api/v1/admin/languages/:code/translations/:app` | yes | Its text for `id` or `console`, beside the English |
| `PUT`  | `/api/v1/admin/languages/:code/translations/:app` | yes | Replace that text              |
| `GET`  | `/api/v1/admin/panel/languages` | no            | The languages the panel can be shown in |
| `GET`  | `/api/v1/admin/panel/languages/:code` | no      | The panel's text in one of them |
| `GET`  | `/api/v1/admin/database/tables` | yes           | The server's own tables        |
| `GET`  | `/api/v1/admin/database/tables/:table` | yes    | One table's columns and rows (`?limit=&offset=`) |
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
| `GET`    | `/api/v1/account/me`                                 | The signed-in user                |
| `PATCH`  | `/api/v1/account/me`                                 | Change their name                 |
| `POST`   | `/api/v1/account/password`                           | Change the password; signs out everywhere else |
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

The header holds the logo, the theme toggle and the account menu, which is
where Profile and Sign out live. Where to go is the dashboard's sidebar, whose
column the header's logo block tops: the two are one width and fold together.

```
Activity · Logs
Applications     Applications · APIs · SSO integrations (soon)
Authentication   Database · Social · Login flows
User management  Users · Roles
Administration   Administrators · Admin roles   (super admins only)
Settings         Organization · Languages
```

Each link is shown only to an administrator whose roles allow the page. The
logs live at `/admin/dashboard/logs`; the old `/admin/logs` redirects there.
SSO integrations is the one section still marked "Soon", and its page says
what is planned; every other page talks to the API.

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

The profile page collects what belongs to the signed-in account, in a
centred column: profile information, sign-in and security (email, password),
preferences (theme, language) and sessions. Theme, language and sign out work;
the rest are marked as not available yet and their controls are disabled
rather than pretending. Two-factor sign-in is not here — it belongs to the Administrators
page, which is where the policy lives and where one administrator's factor is
read and reset.

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
column instead. Below 55rem the sidebar becomes a scrollable row above the
content.

**Styling.** [Ark UI](https://ark-ui.com) ships no CSS: every part it renders
carries `data-scope` and `data-part`, and `lib/styles/ark.css` styles those
attributes. Everything else refers to the tokens in `tokens.css`. There is no
CSS framework.

The palette is taken from the [PocketBase](https://pocketbase.io) admin UI —
its near-black primary, soft grey secondary, navy header and filled inputs —
with the dark theme built the way PocketBase builds it: one base colour mixed
with increasing amounts of white, so the greys stay in step.

Fields follow their settings page exactly: one filled block with the label
inside it at the top and the value below, no border anywhere, and focus shown
by the block darkening (`#e4e8ec` to `#dce0e5`) while the label goes from
`#687278` to the full text colour. The measurements — a 24px label row over a
38.5px value row, 5px radius, 13px bold label, 13px side padding — are in
`lib/styles/ark.css`.

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

**Fonts.** Product Sans for text and Consolas for code, both loaded with
`local()` only. Neither can be bundled — Product Sans is Google's corporate
typeface and is not licensed for redistribution, and Consolas ships with
Windows and Office — so a machine that has them uses them and one that does
not falls back quietly. To self-host licensed copies, put the files in
`static/fonts` and add a `url(...)` source in `lib/styles/fonts.css`.

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
challenge; `/oauth2/social/<slug>/callback` — which answers POST as well,
because Apple posts its answer — trades the code for a token, reads the
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
verifier, and where the person was going, single use and short-lived.

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
(`link_verified_emails`); then a new account, if the provider and the
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

### Login flows

Authentication · Login flows is where an administrator writes what a sign-in
asks for: an ordered list of steps, and the things a person is allowed to do
along the way. There are as many flows as somebody has written, one of them is
the default, and an application either names a flow of its own or falls back
to that one.

```
Identify → Password → Another account        offered to everything else
Identify → Password → Emailed code           offered to the staff tools
```

A flow carries the steps, whether an account can be made, whether a password
can be reset, whether an address has to be confirmed first, and how long a
session it makes lasts. The steps come from a catalog in
`internal/model/login_flow.go` — the model owns what a step is, what it is
called and what it does, so adding one is an entry there and the panel's
picker follows.

**What the server runs today is not the whole catalog, and it says so.** The
sign-in pages read a flow's options: `GET /api/v1/account/login-options`
answers with the effective flow for the sign-in under way, and the pages use
it to decide whether to offer "Create an account", whether to offer
"Forgotten your password", and whether to show the provider buttons. The
server holds the same two: a registration is refused when the flow does not
allow one (`oidc.Register`), and a reset link is not sent when it does not
offer resets (`oidc.ForgotPassword`) — silently, so the page still cannot be
used to find out which addresses have accounts.

Walking the steps themselves is not built. A flow may name `email_code`,
`totp`, `terms` or `consent`, and those steps are marked "not run yet" in the
catalog, in the list and in the drawer, so a flow reads as the plan it is
rather than a promise. `LoginStepSpec.Implemented` is the one place that says
which is which: implementing a step is flipping it there and writing the step.

The default flow is what everything falls back to, so it cannot be turned off
or removed; making another flow the default takes the mark from it. A flow
that is turned off keeps its applications — they fall back to the default —
and removing one clears the column rather than taking its applications with
it. Reading the page takes `login_flows.read` and writing takes
`login_flows.write`.

### The database browser

Authentication · Database is what the migration built, as the database holds
it: every table with how many rows and columns it has, and a page of any one
of them. It is the place to look when a panel page is not showing what you
expected, and the only place that shows the tables a page has no editor for —
the codes, the tokens and the sessions the provider issues.

It reads and nothing else. There is no endpoint behind it that writes a row,
and there is not meant to be: a user, an application or a role is changed on
the page that knows what one is and what changing it costs.

**A password, a key, a one-time code or the hash standing in for a token is
never read.** Those columns are dropped from the SELECT rather than blanked
afterwards, so a value that is not shown is a value that never left the
database; the panel still lists the column and marks it hidden, which is
truer than leaving it out and looking like the table has no such thing. Which
columns those are is a rule on the name rather than a list of columns —
anything ending in `_hash` or `_secret`, and `secret`, `password`,
`private_key` and `recovery_codes` — so a model added later is covered by the
naming this project already follows. `TestHiddenColumn` holds it in both
directions: `secret_hint` and `allow_password_reset` are not secrets.

Nothing a request sends ever reaches a query's text. A table is only read
after its name has been found in the list of tables the database actually has,
and the columns are named through the driver's quoting — `api_scopes` has a
column called `default`, which is a syntax error unquoted.

It takes `database.read`, which is the strongest of the read permissions: every
account and every application is visible through it, so no seeded role but
`admin` grants it.

### Languages

Every word either app shows is looked up by key — `t('login.title')` — and the
text for each language lives in the database: a `languages` row with its names
and settings, and a `translations` row per app holding one JSON object of text
by key. **Settings · Languages** is where languages are added, translated,
offered, made the default and removed, and a change is on the next page
anybody opens: both apps ask the API for their text while rendering, so
nothing is rebuilt.

**The files under `locales/` are what the server ships with:**

```
locales/
  id/       en.json  ky.json  ru.json     the sign-in pages and a user's own account
  console/  en.json  ky.json  ru.json     the admin panel
```

They are embedded in the binary (`locales/locales.go`) and have three jobs:

- **The first start imports them.** `store.EnsureLanguages` runs in `main`: on
  a database with no text at all it writes every shipped language, English on
  and the rest off. From then on the database is the panel's.
- **A release reaches a shipped language.** Each start copies a key a shipped
  file has into the database's copy of that language when the copy lacks it,
  and never touches a message that is there. The one consequence: a message
  cleared in the panel, in a language that ships, comes back on the next
  start — a shipped language can be reworded, and the way to empty one is to
  remove it. A removed language is never brought back by a start; the New
  language drawer offers it back instead.
- **`en.json` is the contract.** Its keys are the keys there are: coverage is
  counted against them, a key they do not have is dropped on save, and a key
  a language has no text for is sent in English — the database's English, and
  under that the shipped file's, so a page never shows a bare key.

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
`$name` and `$native` on top, which is the shape of the shipped files, so it
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
entries are links to `?lang=ky`, remembered in a cookie and redirected away,
so a shared address carries nobody's choice. Which language somebody gets is
their saved choice, then their browser's `Accept-Language` (by exact tag,
then by its language part, so `ru-RU` is served `ru`), then the
installation's default — and only ever one that is offered, so a choice of a
language since turned off falls through. `GET /api/v1/account/languages` is
the list and `GET /api/v1/account/languages/:code` one language's text, every
key filled in; a language that is off is a 404 there.

**In the admin panel** the language is one administrator's own preference on
one machine — a cookie, set on the Profile page beside the theme — rather
than a setting of the installation. Every language with some of the panel
translated is offered there, whatever the Languages page says: that page
decides what *users* see. Saving the text of the language the panel is shown
in redraws the panel.

Both apps resolve the language and fetch its text while rendering on the
server, so the first response is already translated and `<html lang>` is
right in the first byte, which is what a screen reader reads the page's words
with. Each app bundles English alone, for when the API cannot be reached.

What is translated today: all of the sign-in pages and a user's own account,
and the admin panel's shell — its sidebar, its account menu, the Profile page
and the Languages page. The rest of the panel's pages are still English in the
markup; moving one over is replacing its strings with `t('key')` and adding
the keys to `locales/console/*.json`. `TestTranslationsAreComplete` fails the
build if a shipped language falls behind the base, so a key added without a
translation is caught rather than shipped.

Kyrgyz and Russian were written alongside the machinery and would be worth a
native speaker's eye before an installation offers them.

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

Right now there is one: `20260917150000_schema.go` builds the whole schema
from `model.All()`, adds the indexes a struct tag cannot describe, and seeds
the admin roles, the organisation, the default login flow and the base
language. While nothing is deployed, that is the
better shape — the models are the schema, and one file cannot disagree with
them. The first database that has to be brought forward without being rebuilt
is when the second migration gets written; from then on the rule below holds.

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
anything through the admin API. With no
`XERMESS_SMTP_HOST`, password reset emails are written to the log.

Session cookies are Secure when `XERMESS_ACCOUNT_URL` and `XERMESS_ADMIN_URL`
are https, without a setting to forget. `XERMESS_TRUSTED_PROXIES` lists the
reverse proxies whose `X-Forwarded-For` is believed. With none listed, the
address recorded for every request is the connection's own, so a caller cannot
write a made-up one into the activity log — but behind a proxy it has to be
set, or every request, and the rate limit, count as the proxy's.
`XERMESS_ADMIN_ADDR` must differ from `XERMESS_ADDR`; never publish it.

`make test-integration` runs the tests that need Postgres. They connect to
the server in `.env` only to create a database of their own for each test,
and drop it afterwards; without `XERMESS_TEST_DB_DSN` set, `go test` skips
them.

The migrations are compiled into the binary. Goose still needs the directory in
`XERMESS_DB_MIGRATE_DIR` to exist; when it holds no `.go` files, as in the
container image, every compiled-in migration is applied.

## License

[MIT](LICENSE)
