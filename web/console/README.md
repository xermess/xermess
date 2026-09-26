# console

The Loginer **admin panel**: users, roles, applications, APIs, administrators
and the activity log. Only administrators use it; users sign in and manage
their accounts in `web/id`.

```sh
cp .env.example .env   # API_URL: the API's admin listener
bun install
bun run dev            # http://localhost:5174/admin/login
```

How it is built — server-rendered loads, the query cache, the design system on
Ark UI — is described in the repository's README.
