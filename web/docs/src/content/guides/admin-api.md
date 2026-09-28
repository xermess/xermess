---
title: Manage users with admin-cli
description: Call the Admin API from your own code — create, update and delete users — with the built-in admin-cli client.
order: 1
section: Manage users
nav: 'With admin-cli'
icon: terminal-box
---

| Guide                                        | Who acts                           | Manages         | Use for                                                |
| -------------------------------------------- | ---------------------------------- | --------------- | ------------------------------------------------------ |
| **[With admin-cli](/guides/admin-api)**      | Your backend, as a service         | Any user        | Provisioning, admin tools, syncing from another system |
| **[User self-service](/guides/account-api)** | A signed-in user, through your app | Only themselves | Profile and security screens inside your app           |

Everything the panel does, your code can do through the **Admin API**: create users, change them, give them roles, sign them out. It authenticates with **admin-cli**, a machine-to-machine application every installation starts with.

## How it works

1. `admin-cli` gets an access token with the **client credentials** grant, for the Admin API (audience `{{adminAPI}}`).
2. The token's scopes are **admin permissions** — `users.read`, `users.write`, and so on. You choose which in the panel.
3. Your code calls `$ADMIN_API/api/v1/admin/...` with `Authorization: Bearer <token>`. Each route checks the same permission it checks for a person.

## 1. Set up admin-cli

In the panel:

1. Open **Applications → admin-cli → API access**, and under **Admin API** tick the permissions it needs — for managing users, `users.read`, `users.write` and `role_assignments.write`.
2. On its **Settings** tab, press **Rotate secret**. Copy the secret: it is shown once.

> [!TIP]
> Give each integration its own application instead of sharing admin-cli: create a **machine-to-machine** application and authorize it for the **Admin API** the same way. Each then has its own secret, scopes and entry in the activity log.

## 2. Get a token

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/token" \
  -u "admin-cli:$CLIENT_SECRET" \
  -d "grant_type=client_credentials" \
  -d "audience={{adminAPI}}"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder(URI.create(ISSUER + "/oauth2/token"))
        .header("Authorization", "Basic " + Base64.getEncoder()
                .encodeToString(("admin-cli:" + CLIENT_SECRET).getBytes(UTF_8)))
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(BodyPublishers.ofString("grant_type=client_credentials"
                + "&audience=" + URLEncoder.encode("{{adminAPI}}", UTF_8)))
        .build();
String tokens = HttpClient.newHttpClient().send(request, BodyHandlers.ofString()).body();
```

```go title="Go"
// golang.org/x/oauth2/clientcredentials: refreshes the token as it expires.
config := clientcredentials.Config{
	ClientID:       "admin-cli",
	ClientSecret:   clientSecret,
	TokenURL:       issuer + "/oauth2/token",
	EndpointParams: url.Values{"audience": {"{{adminAPI}}"}},
}
client := config.Client(ctx)
```

```python title="Python"
tokens = requests.post(
    f"{ISSUER}/oauth2/token",
    data={"grant_type": "client_credentials", "audience": "{{adminAPI}}"},
    auth=("admin-cli", CLIENT_SECRET),
).json()
ADMIN_TOKEN = tokens["access_token"]
```

```json
{
	"access_token": "eyJhbGciOiJSUzI1NiIs…",
	"token_type": "Bearer",
	"expires_in": 300,
	"scope": "users.read users.write role_assignments.write"
}
```

With no `scope`, the token gets every permission admin-cli is allowed. Send `scope` to ask for fewer. The token lasts five minutes — ask for a new one when it expires.

## 3. Manage users

The Admin API is on the panel's address: `$ADMIN_API` below, such as `https://admin-id.example.com`.

### Create a user

```bash title="cURL"
curl -X POST "$ADMIN_API/api/v1/admin/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"ada@example.com","first_name":"Ada","last_name":"Lovelace","password":"a-long-password","confirm_password":"a-long-password","is_password_temporary":true}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder(URI.create(ADMIN_API + "/api/v1/admin/users"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .header("Content-Type", "application/json")
        .POST(BodyPublishers.ofString("""
                {"email": "ada@example.com", "first_name": "Ada", "last_name": "Lovelace",
                 "password": "a-long-password", "confirm_password": "a-long-password",
                 "is_password_temporary": true}
                """))
        .build();
HttpResponse<String> response = HttpClient.newHttpClient().send(request, BodyHandlers.ofString());
```

```go title="Go"
body := strings.NewReader(`{"email": "ada@example.com", "first_name": "Ada", "last_name": "Lovelace",
	"password": "a-long-password", "confirm_password": "a-long-password", "is_password_temporary": true}`)
req, _ := http.NewRequest(http.MethodPost, adminAPI+"/api/v1/admin/users", body)
req.Header.Set("Content-Type", "application/json")
res, err := client.Do(req) // the clientcredentials client adds the token
```

```python title="Python"
user = requests.post(
    f"{ADMIN_API}/api/v1/admin/users",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
    json={
        "email": "ada@example.com",
        "first_name": "Ada",
        "last_name": "Lovelace",
        "password": "a-long-password",
        "confirm_password": "a-long-password",
        "is_password_temporary": True,
    },
).json()["user"]
```

`is_password_temporary` makes the user choose a new password at their first sign-in.

### Find, change and delete

| Do                  | Call                                                                 | Permission               |
| ------------------- | -------------------------------------------------------------------- | ------------------------ |
| Search users        | `GET /api/v1/admin/users?search=ada`                                 | `users.read`             |
| Get one             | `GET /api/v1/admin/users/{id}`                                       | `users.read`             |
| Update              | `PATCH /api/v1/admin/users/{id}`                                     | `users.write`            |
| Deactivate          | `PATCH /api/v1/admin/users/{id}` with `"is_active": false`           | `users.write`            |
| Sign out everywhere | `DELETE /api/v1/admin/users/{id}/sessions`                           | `users.write`            |
| Give roles          | `POST /api/v1/admin/users/{id}/role-mappings` with `{"roles": [id]}` | `role_assignments.write` |
| Take a role away    | `DELETE /api/v1/admin/users/{id}/role-mappings/{role}`               | `role_assignments.write` |
| Delete              | `DELETE /api/v1/admin/users/{id}`                                    | `users.write`            |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/users?search=ada" -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder(URI.create(ADMIN_API + "/api/v1/admin/users?search=ada"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .build();
String users = HttpClient.newHttpClient().send(request, BodyHandlers.ofString()).body();
```

```go title="Go"
res, err := client.Get(adminAPI + "/api/v1/admin/users?search=ada")
```

```python title="Python"
users = requests.get(
    f"{ADMIN_API}/api/v1/admin/users",
    params={"search": "ada"},
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
).json()["users"]
```

`PATCH` takes the user's fields whole — `email` included — so read the user first and send it back changed. Every route, with its fields and answers, is in the [Admin API reference](/reference/admin). Routes marked **admin-cli token** take one.

## What a token can't do

- **Super-admin routes** — administrators, their roles, mail, signing keys, the cache — and an administrator's own account (`/me`, sessions, second factor) take a person's session only.
- **A user's token** is refused, even one issued for the Admin API: only a service's client credentials token is accepted.

## Good to know

- **Scopes are checked on every call.** Untick a permission in the panel and the next call is refused, even with a token issued before.
- **Turn admin-cli off** under **Applications → admin-cli** to stop it at once. It can't be deleted: the server makes it again at startup.
- **The activity log** records admin-cli's changes under its name.
- Errors are the usual `{"error", "code"}` — `403 forbidden` when a permission is missing, `401 token_refused` when the token is expired or not accepted. See [Errors](/reference/errors).
