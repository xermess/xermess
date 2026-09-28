---
title: "Users"
description: "The people your organisation manages."
order: 7
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

## GET /api/v1/admin/users {#get-api-v1-admin-users}

Returns a page of users, newest first.

`search` matches the email or any text the user-defined fields hold, so one box covers the whole record rather than one column.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`users.read`](/reference/permissions#users-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/users?search=value&role=value" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/users" + "?search=value&role=value"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/users"+"?search=value&role=value", nil)
if err != nil {
	return err
}
req.Header.Set("Authorization", "Bearer "+adminToken)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.get(
    f"{ADMIN_API}/api/v1/admin/users",
    params={
        "search": "value",
        "role": "value",
    },
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `search` | query |  |
| `role` | query |  |

#### Response

`200 OK`

```json
{
  "users": [
    {
      "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
      "created_at": "2026-09-28T09:30:00Z",
      "updated_at": "2026-09-28T09:30:00Z",
      "email": "ada@example.com",
      "is_email_verified": true,
      "first_name": "Ada",
      "last_name": "Lovelace",
      "is_active": true,
      "is_password_temporary": true,
      "has_password": true,
      "social_accounts": [
        {
          "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
          "provider": "string",
          "slug": "example",
          "kind": "apple",
          "email": "ada@example.com",
          "connected_at": "2026-09-28T09:30:00Z",
          "last_login_at": "2026-09-28T09:30:00Z"
        }
      ],
      "last_login_at": "2026-09-28T09:30:00Z",
      "locked_until": "2026-09-28T09:30:00Z",
      "data": {},
      "roles": [
        {
          "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
          "created_at": "2026-09-28T09:30:00Z",
          "updated_at": "2026-09-28T09:30:00Z",
          "application_id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
          "name": "Example",
          "description": "string",
          "is_default": true,
          "inherits": [
            {
              "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
              "created_at": "2026-09-28T09:30:00Z",
              "updated_at": "2026-09-28T09:30:00Z",
              "application_id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
              "name": "Example",
              "description": "string",
              "is_default": true,
              "inherits": [],
              "api_scopes": []
            }
          ],
          "api_scopes": [
            {
              "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
              "created_at": "2026-09-28T09:30:00Z",
              "updated_at": "2026-09-28T09:30:00Z",
              "api_id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
              "name": "Example",
              "description": "string",
              "is_default": true
            }
          ]
        }
      ]
    }
  ],
  "total": 1,
  "limit": 1,
  "offset": 1
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## GET /api/v1/admin/users/:id {#get-api-v1-admin-users-id}

Returns one user.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`users.read`](/reference/permissions#users-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/users/$ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/users/" + id))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/users/"+id, nil)
if err != nil {
	return err
}
req.Header.Set("Authorization", "Bearer "+adminToken)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.get(
    f"{ADMIN_API}/api/v1/admin/users/{id}",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `id` <small>required</small> | path |  |

#### Response

`200 OK`

```json
{
  "user": {
    "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
    "created_at": "2026-09-28T09:30:00Z",
    "updated_at": "2026-09-28T09:30:00Z",
    "email": "ada@example.com",
    "is_email_verified": true,
    "first_name": "Ada",
    "last_name": "Lovelace",
    "is_active": true,
    "is_password_temporary": true,
    "has_password": true,
    "social_accounts": [
      {
        "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
        "provider": "string",
        "slug": "example",
        "kind": "apple",
        "email": "ada@example.com",
        "connected_at": "2026-09-28T09:30:00Z",
        "last_login_at": "2026-09-28T09:30:00Z"
      }
    ],
    "last_login_at": "2026-09-28T09:30:00Z",
    "locked_until": "2026-09-28T09:30:00Z",
    "data": {},
    "roles": [
      {
        "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
        "created_at": "2026-09-28T09:30:00Z",
        "updated_at": "2026-09-28T09:30:00Z",
        "application_id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
        "name": "Example",
        "description": "string",
        "is_default": true,
        "inherits": [
          {
            "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
            "created_at": "2026-09-28T09:30:00Z",
            "updated_at": "2026-09-28T09:30:00Z",
            "application_id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
            "name": "Example",
            "description": "string",
            "is_default": true,
            "inherits": [],
            "api_scopes": []
          }
        ],
        "api_scopes": [
          {
            "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
            "created_at": "2026-09-28T09:30:00Z",
            "updated_at": "2026-09-28T09:30:00Z",
            "api_id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
            "name": "Example",
            "description": "string",
            "is_default": true
          }
        ]
      }
    ]
  }
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## GET /api/v1/admin/users/:id/roles {#get-api-v1-admin-users-id-roles}

Returns the roles a user holds as a token would carry them: the global roles, and for each application its roles — each as the ones given directly and every role those include.

The client_id query parameter narrows the applications to one. Applications the administrator cannot see are left out.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`users.read`](/reference/permissions#users-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/users/$ID/roles?client_id=shop-web" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/users/" + id + "/roles" + "?client_id=shop-web"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/users/"+id+"/roles"+"?client_id=shop-web", nil)
if err != nil {
	return err
}
req.Header.Set("Authorization", "Bearer "+adminToken)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.get(
    f"{ADMIN_API}/api/v1/admin/users/{id}/roles",
    params={
        "client_id": "shop-web",
    },
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `id` <small>required</small> | path |  |
| `client_id` | query | The application's client ID. |

#### Response

`200 OK`

```json
{
  "global": {
    "roles": [
      "string"
    ],
    "effective_roles": [
      "string"
    ]
  },
  "applications": [
    {
      "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
      "client_id": "shop-web",
      "name": "Example",
      "roles": [
        "string"
      ],
      "effective_roles": [
        "string"
      ]
    }
  ]
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## GET /api/v1/admin/users/:id/role-mappings {#get-api-v1-admin-users-id-role-mappings}

Returns every role a user holds, Keycloak's role mapping: the roles given directly, and the roles that come to the user through them, each with what it comes through.

Roles of applications the administrator cannot see are left out.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`users.read`](/reference/permissions#users-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/users/$ID/role-mappings" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/users/" + id + "/role-mappings"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/users/"+id+"/role-mappings", nil)
if err != nil {
	return err
}
req.Header.Set("Authorization", "Bearer "+adminToken)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.get(
    f"{ADMIN_API}/api/v1/admin/users/{id}/role-mappings",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `id` <small>required</small> | path |  |

#### Response

`200 OK`

```json
{
  "roles": [
    {
      "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
      "name": "Example",
      "description": "string",
      "application": {
        "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
        "name": "Example",
        "client_id": "shop-web"
      },
      "composite": true,
      "assigned": true,
      "via": [
        {
          "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
          "name": "Example"
        }
      ]
    }
  ]
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## POST /api/v1/admin/users {#post-api-v1-admin-users}

Adds a user.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`users.write`](/reference/permissions#users-write) |

```bash title="cURL"
curl -X POST "$ADMIN_API/api/v1/admin/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"ada@example.com"}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/users"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .header("Content-Type", "application/json")
        .POST(BodyPublishers.ofString("""
                {
                  "email": "ada@example.com"
                }
                """))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
body := strings.NewReader(`{
	"email": "ada@example.com"
}`)
req, err := http.NewRequest(http.MethodPost, adminAPI+"/api/v1/admin/users", body)
if err != nil {
	return err
}
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+adminToken)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.post(
    f"{ADMIN_API}/api/v1/admin/users",
    json={
        "email": "ada@example.com",
    },
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `email` <small>required</small> | string | An email address, at most 255 characters. |
| `is_email_verified` | boolean or null | `is_email_verified` and `is_active` may be left out: a new user is then unverified and active, and an existing one keeps what they had. |
| `first_name` | string | At most 100 characters. |
| `last_name` | string | At most 100 characters. |
| `is_active` | boolean or null |  |
| `password` | string | `password` is required when creating a user. When updating one it is optional: left empty, the password the user has is kept. At most 72 characters. |
| `confirm_password` | string | The same as `password`. |
| `is_password_temporary` | boolean or null | `is_password_temporary` marks the password as one the user has to replace. Left out, a new user's password is not temporary and an existing user's keeps what it was. |
| `data` | map of any |  |

#### Response

`201 Created`

```json
{
  "user": {
    "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
    "created_at": "2026-09-28T09:30:00Z",
    "updated_at": "2026-09-28T09:30:00Z",
    "email": "ada@example.com",
    "is_email_verified": true,
    "first_name": "Ada",
    "last_name": "Lovelace",
    "is_active": true,
    "is_password_temporary": true,
    "has_password": true,
    "social_accounts": [
      {
        "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
        "provider": "string",
        "slug": "example",
        "kind": "apple",
        "email": "ada@example.com",
        "connected_at": "2026-09-28T09:30:00Z",
        "last_login_at": "2026-09-28T09:30:00Z"
      }
    ],
    "last_login_at": "2026-09-28T09:30:00Z",
    "locked_until": "2026-09-28T09:30:00Z",
    "data": {},
    "roles": [
      {
        "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
        "created_at": "2026-09-28T09:30:00Z",
        "updated_at": "2026-09-28T09:30:00Z",
        "application_id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
        "name": "Example",
        "description": "string",
        "is_default": true,
        "inherits": [
          {
            "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
            "created_at": "2026-09-28T09:30:00Z",
            "updated_at": "2026-09-28T09:30:00Z",
            "application_id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
            "name": "Example",
            "description": "string",
            "is_default": true,
            "inherits": [],
            "api_scopes": []
          }
        ],
        "api_scopes": [
          {
            "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
            "created_at": "2026-09-28T09:30:00Z",
            "updated_at": "2026-09-28T09:30:00Z",
            "api_id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
            "name": "Example",
            "description": "string",
            "is_default": true
          }
        ]
      }
    ]
  }
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## PATCH /api/v1/admin/users/:id {#patch-api-v1-admin-users-id}

Replaces a user's email, verified flag and fields, and the password too when a new one is given.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`users.write`](/reference/permissions#users-write) |

```bash title="cURL"
curl -X PATCH "$ADMIN_API/api/v1/admin/users/$ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"ada@example.com"}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/users/" + id))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .header("Content-Type", "application/json")
        .method("PATCH", BodyPublishers.ofString("""
                {
                  "email": "ada@example.com"
                }
                """))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
body := strings.NewReader(`{
	"email": "ada@example.com"
}`)
req, err := http.NewRequest(http.MethodPatch, adminAPI+"/api/v1/admin/users/"+id, body)
if err != nil {
	return err
}
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+adminToken)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.patch(
    f"{ADMIN_API}/api/v1/admin/users/{id}",
    json={
        "email": "ada@example.com",
    },
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `id` <small>required</small> | path |  |

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `email` <small>required</small> | string | An email address, at most 255 characters. |
| `is_email_verified` | boolean or null | `is_email_verified` and `is_active` may be left out: a new user is then unverified and active, and an existing one keeps what they had. |
| `first_name` | string | At most 100 characters. |
| `last_name` | string | At most 100 characters. |
| `is_active` | boolean or null |  |
| `password` | string | `password` is required when creating a user. When updating one it is optional: left empty, the password the user has is kept. At most 72 characters. |
| `confirm_password` | string | The same as `password`. |
| `is_password_temporary` | boolean or null | `is_password_temporary` marks the password as one the user has to replace. Left out, a new user's password is not temporary and an existing user's keeps what it was. |
| `data` | map of any |  |

#### Response

`200 OK`

```json
{
  "user": {
    "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
    "created_at": "2026-09-28T09:30:00Z",
    "updated_at": "2026-09-28T09:30:00Z",
    "email": "ada@example.com",
    "is_email_verified": true,
    "first_name": "Ada",
    "last_name": "Lovelace",
    "is_active": true,
    "is_password_temporary": true,
    "has_password": true,
    "social_accounts": [
      {
        "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
        "provider": "string",
        "slug": "example",
        "kind": "apple",
        "email": "ada@example.com",
        "connected_at": "2026-09-28T09:30:00Z",
        "last_login_at": "2026-09-28T09:30:00Z"
      }
    ],
    "last_login_at": "2026-09-28T09:30:00Z",
    "locked_until": "2026-09-28T09:30:00Z",
    "data": {},
    "roles": [
      {
        "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
        "created_at": "2026-09-28T09:30:00Z",
        "updated_at": "2026-09-28T09:30:00Z",
        "application_id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
        "name": "Example",
        "description": "string",
        "is_default": true,
        "inherits": [
          {
            "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
            "created_at": "2026-09-28T09:30:00Z",
            "updated_at": "2026-09-28T09:30:00Z",
            "application_id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
            "name": "Example",
            "description": "string",
            "is_default": true,
            "inherits": [],
            "api_scopes": []
          }
        ],
        "api_scopes": [
          {
            "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
            "created_at": "2026-09-28T09:30:00Z",
            "updated_at": "2026-09-28T09:30:00Z",
            "api_id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
            "name": "Example",
            "description": "string",
            "is_default": true
          }
        ]
      }
    ]
  }
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## DELETE /api/v1/admin/users/:id {#delete-api-v1-admin-users-id}

Removes a user for good.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`users.write`](/reference/permissions#users-write) |

```bash title="cURL"
curl -X DELETE "$ADMIN_API/api/v1/admin/users/$ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/users/" + id))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .DELETE()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodDelete, adminAPI+"/api/v1/admin/users/"+id, nil)
if err != nil {
	return err
}
req.Header.Set("Authorization", "Bearer "+adminToken)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.delete(
    f"{ADMIN_API}/api/v1/admin/users/{id}",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `id` <small>required</small> | path |  |

#### Response

`204 No Content`

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## DELETE /api/v1/admin/users/:id/social-accounts/:identity {#delete-api-v1-admin-users-id-social-accounts-identity}

Takes away a provider a user signs in with: an account of theirs somewhere else that should no longer reach this one.

Their account stays, and so does every other way into it.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`users.write`](/reference/permissions#users-write) |

```bash title="cURL"
curl -X DELETE "$ADMIN_API/api/v1/admin/users/$ID/social-accounts/$IDENTITY" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/users/" + id + "/social-accounts/" + identity))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .DELETE()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodDelete, adminAPI+"/api/v1/admin/users/"+id+"/social-accounts/"+identity, nil)
if err != nil {
	return err
}
req.Header.Set("Authorization", "Bearer "+adminToken)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.delete(
    f"{ADMIN_API}/api/v1/admin/users/{id}/social-accounts/{identity}",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `id` <small>required</small> | path |  |
| `identity` <small>required</small> | path |  |

#### Response

`204 No Content`

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## POST /api/v1/admin/users/:id/role-mappings {#post-api-v1-admin-users-id-role-mappings}

Gives a user roles directly, global and application roles alike, leaving what they already hold as it is.

Each role is allowed by its scope: role_assignments.write for its application, or for the whole panel when it is global. One refused role refuses the whole request.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`role_assignments.write`](/reference/permissions#role_assignments-write) for at least one application |

```bash title="cURL"
curl -X POST "$ADMIN_API/api/v1/admin/users/$ID/role-mappings" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"roles":["0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d"]}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/users/" + id + "/role-mappings"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .header("Content-Type", "application/json")
        .POST(BodyPublishers.ofString("""
                {
                  "roles": [
                    "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d"
                  ]
                }
                """))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
body := strings.NewReader(`{
	"roles": [
		"0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d"
	]
}`)
req, err := http.NewRequest(http.MethodPost, adminAPI+"/api/v1/admin/users/"+id+"/role-mappings", body)
if err != nil {
	return err
}
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Authorization", "Bearer "+adminToken)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.post(
    f"{ADMIN_API}/api/v1/admin/users/{id}/role-mappings",
    json={
        "roles": [
            "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
        ],
    },
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `id` <small>required</small> | path |  |

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `roles` | array of string (uuid) |  |

#### Response

`200 OK`

```json
{
  "roles": [
    {
      "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
      "name": "Example",
      "description": "string",
      "application": {
        "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
        "name": "Example",
        "client_id": "shop-web"
      },
      "composite": true,
      "assigned": true,
      "via": [
        {
          "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
          "name": "Example"
        }
      ]
    }
  ]
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## DELETE /api/v1/admin/users/:id/role-mappings/:role {#delete-api-v1-admin-users-id-role-mappings-role}

Takes away a role the user was given directly.

A role that only comes to the user through another cannot be taken away on its own: the role it comes through has to go.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`role_assignments.write`](/reference/permissions#role_assignments-write) for at least one application |

```bash title="cURL"
curl -X DELETE "$ADMIN_API/api/v1/admin/users/$ID/role-mappings/$ROLE" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/users/" + id + "/role-mappings/" + role))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .DELETE()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodDelete, adminAPI+"/api/v1/admin/users/"+id+"/role-mappings/"+role, nil)
if err != nil {
	return err
}
req.Header.Set("Authorization", "Bearer "+adminToken)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.delete(
    f"{ADMIN_API}/api/v1/admin/users/{id}/role-mappings/{role}",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `id` <small>required</small> | path |  |
| `role` <small>required</small> | path |  |

#### Response

`200 OK`

```json
{
  "roles": [
    {
      "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
      "name": "Example",
      "description": "string",
      "application": {
        "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
        "name": "Example",
        "client_id": "shop-web"
      },
      "composite": true,
      "assigned": true,
      "via": [
        {
          "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
          "name": "Example"
        }
      ]
    }
  ]
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
