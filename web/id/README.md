# id

The app xermess **users** see. Nothing in it is for administrators — that is
`web/console`.

- **Signing in** for an application: `/login`, `/register`, `/forgot-password`,
  `/reset-password`, and `/error` and `/logged-out`, which the provider sends
  browsers to. Each page shows the application's name, logo, terms and privacy
  links.
- **Managing the account**, once signed in: `/` (profile), `/security`
  (password, and the devices signed in on), `/applications` (apps that can act
  for you, and disconnecting them).

```sh
cp .env.example .env   # API_URL: the API's public listener
bun install
bun run dev            # http://localhost:5173
```

The API must know this app's origin: `XERMESS_ISSUER=http://localhost:5173`.
The browser calls the API on this same origin, so no CORS setting is needed.

## Layout

```
src/routes/(auth)/        the sign-in pages: one centred card, no session needed
src/routes/(account)/     the account pages: behind the session, with a header and tabs
src/lib/api/              the typed calls to /api/v1/account
src/lib/server/session.ts reading the session in a server load, and sending the cookie on
src/lib/server/proxy.ts   server-side fetches of API paths, sent straight to API_URL
src/lib/components/       Button, TextField, PasswordField, Panel, AuthCard… — Svelte only
src/lib/i18n/             svelte-i18n adapter, grouped English fallback, and the t() context
src/lib/styles/           tokens.css, base.css, fonts.css
src/hooks.server.ts       the theme before first paint; no framing, no referrer
```

The components depend on nothing but Svelte, so the pages users load stay
small. The palette matches `web/console`, but no code is shared with it.
