# Deploying xermess

```
                    internet                          staff network / VPN only
                        │                                        │
             https://id.mywebsite.com              https://admin-id.mywebsite.com
                        │                                        │
                ┌───────┴────────────────── Caddy ───────────────┴───────┐
                │  /oauth2/* /.well-known/*       /api/v1/admin/*        │
                │  /api/v1/account/*    else      │            else      │
                └────┬──────────────────┬─────────┼─────────────┬────────┘
                     │                  │         │             │
               api :8080            id          api :8081       console
             (public listener)    :3000     (admin listener)    :3000
                     └────── Postgres · Redis ────┘
```

Two sites, two apps, one API process with two listeners:

| Site | Serves | API listener | Who can reach it |
| --- | --- | --- | --- |
| `id.mywebsite.com` | the id app, the OAuth/OIDC provider, the account API | public, `:8080` | everyone |
| `admin-id.mywebsite.com` | the console, the admin API | admin, `:8081` | staff only |

## Files

| File | What |
| --- | --- |
| `compose.yaml` | Postgres, Redis, `api`, `id`, `console` and Caddy on one internal network |
| `Caddyfile` | the two sites and which paths go to which listener |
| `docker/api.Dockerfile` | the API; it applies migrations when it starts |
| `docker/web.Dockerfile` | any app under `web/`, chosen with `APP`; the context is the repository, because the apps import `locales/` |
| `.env.example` | the settings compose reads from `deploy/.env` |
| `compose.local.yaml` | an overlay for trying the stack on one machine |
| `Caddyfile.local` | the same two sites on `.localhost`, with Caddy's own CA |
| `../scripts/deploy-local.sh` | what `make deploy-local` runs |

`make deploy-logs` follows the API; `make deploy-down` stops the stack.

## Why it is shaped like this

- **The admin API is not on the public listener at all.** `/api/v1/admin/*`
  answers only on `:8081`, which only the staff site routes to. Hiding the
  panel's page is not enough when its API is on the internet; this way there
  is nothing to find.
- **Each app is on the same origin as the API it calls.** The session cookie
  then belongs to that one host: the browser sends it without CORS, the app's
  server-side renderer can read it, and no other subdomain of mywebsite.com
  ever receives it.
- **Staff access is enforced before the panel.** The Caddyfile refuses anyone
  outside the staff networks. Stronger: no public DNS for the admin host, and
  reach it over WireGuard/Tailscale or an identity-aware proxy.

## Run it

```sh
cp deploy/.env.example deploy/.env      # fill in URLs, secret key, database password, SMTP
$EDITOR deploy/Caddyfile                # hostnames, email, staff networks
make deploy-up                          # docker compose -f deploy/compose.yaml up -d --build
```

Then open `https://admin-id.mywebsite.com/admin/login` from the staff network
to create the first administrator.

## Trying it on one machine

The same stack, with the one part that cannot work locally swapped out: the
hostnames become `id.localhost` and `admin-id.localhost`, Caddy signs them
with its own CA instead of going to Let's Encrypt, and the admin site drops
the staff-network check that would otherwise answer the person running it
404.

```sh
make deploy-local        # writes deploy/.env the first time, then builds and starts
```

It prints where to go: `https://admin-id.localhost/admin/login` sends you to
the page that makes the first administrator. Chrome and Firefox resolve
`*.localhost` themselves; the certificate warning is Caddy's own CA, and
`Caddyfile.local` says at its foot how to trust it.

`make deploy-local-logs` follows the API, `make deploy-local-down` stops the
stack and drops its database. Run the compose commands by hand and both files
have to be named every time, because an explicit `-f` means compose picks up
no override on its own:

```sh
docker compose -f deploy/compose.yaml -f deploy/compose.local.yaml ps
```

## Checklist before going live

- [ ] `PUBLIC_URL` and `ADMIN_URL` are `https` — session cookies are Secure because of it.
- [ ] Use a domain you own for the admin host (not `.local`, which is reserved
      for mDNS and cannot get a trusted certificate).
- [ ] The staff networks in the Caddyfile are right, or the admin host is VPN-only.
- [ ] `XERMESS_SECRET_KEY` is stored in your secret manager. Losing it signs everyone out.
- [ ] SMTP is configured, or reset emails only reach the log.
- [ ] Postgres is backed up.
- [ ] Nothing but Caddy publishes a port (`docker compose ps`).
- [ ] Register your applications' redirect URIs as `https`.
- [ ] `XERMESS_ADMIN_MFA` is `required`. `compose.yaml` asks for that unless
      `deploy/.env` says otherwise — the server's own default is `optional`, so
      do not rely on it elsewhere. Every administrator has saved their recovery
      codes.
- [ ] Someone knows how to reset an administrator's two-factor sign-in (a super
      admin, from Administrators) — and that at least two super admins exist.

## Not in this directory

- Two-factor sign-in for administrators is what this stack asks for, not what
  the server defaults to — see the checklist. It does not replace the network
  restriction; each covers what the other misses.
- Signing keys rotate on their own every 90 days. If one may have leaked, a super
  admin rotates at once with revocation: `POST /api/v1/admin/signing-keys/rotate`
  with `{"revoke_old": true}`.
- Rate limiting counts in Redis, so several API replicas share one limit. Drop
  Redis from the stack and each process counts on its own again.
