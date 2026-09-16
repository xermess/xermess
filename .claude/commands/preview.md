---
description: Run the app on spare ports and look at a page in the browser
argument-hint: "[path, e.g. /admin/dashboard/organization]"
---

Look at $ARGUMENTS (default: `/admin/dashboard/organization`) in a running
panel, without disturbing whatever is already running.

First check whether the normal stack is up — anything answering on 8080, 8081,
5173 or 5174. If nothing is, `make dev` is the simple way and you can stop
reading. If something is, leave it alone and raise a throwaway instance beside
it:

1. A database of its own:
   `psql "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" -c "CREATE DATABASE xermess_preview;"`
2. The API, built from the working tree into `bin/` (which is ignored), on
   spare ports, applying its migrations on start:

   ```sh
   go build -o bin/xermess-preview ./cmd/xermess
   XERMESS_ADDR=:8090 XERMESS_ADMIN_ADDR=:8091 \
     XERMESS_DB_DSN="postgres://postgres:postgres@localhost:5432/xermess_preview?sslmode=disable" \
     XERMESS_DB_DRIVER=postgres XERMESS_DB_MIGRATE=true XERMESS_DB_MIGRATE_DIR=./migrations \
     XERMESS_ISSUER=http://localhost:5183 XERMESS_ACCOUNT_URL=http://localhost:5183 \
     XERMESS_ADMIN_URL=http://localhost:5184 XERMESS_ADMIN_MFA=optional \
     XERMESS_SECRET_KEY=preview-secret-key-0123456789abcdef \
     bin/xermess-preview &
   ```

   `XERMESS_ADMIN_MFA=optional` is what makes signing in one step instead of
   needing an authenticator.

3. The panel against it: `API_URL=http://localhost:8091 bunx vite dev --port 5184 --strictPort`
   from `web/console`.
4. The first administrator, since the database is empty:

   ```sh
   curl -s -X POST http://localhost:8091/api/v1/admin/setup \
     -H "Content-Type: application/json" -H "Origin: http://localhost:5184" \
     -d '{"email":"preview@example.com","password":"preview-password-1","first_name":"Preview"}'
   ```

5. Sign in at `http://localhost:5184/admin/login` with the browser tools and go
   to the path. Look at it in both themes if the change is visual — the header's
   toggle switches them.

Then take it all down: stop both processes, `DROP DATABASE xermess_preview WITH (FORCE)`,
and close the tab. Say what you saw, and what you changed if you fixed
something.
