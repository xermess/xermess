---
title: "Social providers"
description: "Signing in with an account at another provider."
order: 11
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

## GET /api/v1/admin/social-providers {#get-api-v1-admin-social-providers}

Returns every configured provider, and the kinds a new one may be.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`social.read`](/reference/permissions#social-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/social-providers" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/social-providers"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/social-providers", nil)
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
    f"{ADMIN_API}/api/v1/admin/social-providers",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Response

`200 OK`

```json
{
  "providers": [
    {
      "id": "string",
      "kind": "apple",
      "slug": "example",
      "name": "Example",
      "client_id": "shop-web",
      "has_client_secret": true,
      "team_id": "string",
      "key_id": "string",
      "has_private_key": true,
      "scopes": [
        "openid profile email"
      ],
      "token_auth": "basic",
      "token_auth_used": "basic",
      "authorize_url": "https://example.com",
      "token_url": "https://example.com",
      "userinfo_url": "https://example.com",
      "is_enabled": true,
      "link_verified_emails": true,
      "allow_registration": true,
      "position": 1,
      "callback_url": "https://app.example.com/callback",
      "identities": 1,
      "created_at": "2026-09-28T09:30:00Z"
    }
  ],
  "kinds": [
    {
      "kind": "apple",
      "label": "string",
      "authorize_url": "https://example.com",
      "token_url": "https://example.com",
      "userinfo_url": "https://example.com",
      "scopes": [
        "openid profile email"
      ],
      "custom": true,
      "signed_secret": true,
      "token_auth": "basic",
      "docs": "string"
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

## GET /api/v1/admin/social-providers/:id {#get-api-v1-admin-social-providers-id}

Returns one provider.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`social.read`](/reference/permissions#social-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/social-providers/$ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/social-providers/" + id))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/social-providers/"+id, nil)
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
    f"{ADMIN_API}/api/v1/admin/social-providers/{id}",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `id` <small>required</small> | path |  |

#### Response

Errors come back as an [OAuth error](/reference/errors#oauth-errors): `{"error", "error_description"}`.

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## GET /api/v1/admin/social-providers/:id/secret {#get-api-v1-admin-social-providers-id-secret}

Answers with the client secret itself, for an administrator who has to check what is configured against the provider's console.

It is stored encrypted rather than hashed, so it can be read back — which makes reading it an event worth recording. It takes the permission that could replace it anyway, and every reading is written to the log with who asked.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`social.write`](/reference/permissions#social-write) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/social-providers/$ID/secret" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/social-providers/" + id + "/secret"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/social-providers/"+id+"/secret", nil)
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
    f"{ADMIN_API}/api/v1/admin/social-providers/{id}/secret",
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
  "secret": "string",
  "kind": "string"
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## POST /api/v1/admin/social-providers {#post-api-v1-admin-social-providers}

Registers a provider.

It is not reachable by anyone signing in until somebody has signed in with it once — the credentials are only ever proved right by the provider itself.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`social.write`](/reference/permissions#social-write) |

```bash title="cURL"
curl -X POST "$ADMIN_API/api/v1/admin/social-providers" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/social-providers"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .header("Content-Type", "application/json")
        .POST(BodyPublishers.ofString("""
                {}
                """))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
body := strings.NewReader(`{}`)
req, err := http.NewRequest(http.MethodPost, adminAPI+"/api/v1/admin/social-providers", body)
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
    f"{ADMIN_API}/api/v1/admin/social-providers",
    json={},
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `kind` | string | `kind` and `slug` are read when a provider is registered and never again: the slug is in the address registered with the provider, and the kind decides which addresses those are. |
| `slug` | string |  |
| `name` | string or null |  |
| `client_id` | string or null |  |
| `client_secret` | string or null |  |
| `private_key` | string or null |  |
| `team_id` | string or null |  |
| `key_id` | string or null |  |
| `scopes` | array of string or null |  |
| `token_auth` | string or null | `token_auth` is how the secret is presented at the token endpoint. An empty string leaves it to the kind's own default. |
| `authorize_url` | string or null |  |
| `token_url` | string or null |  |
| `userinfo_url` | string or null |  |
| `is_enabled` | boolean or null |  |
| `link_verified_emails` | boolean or null |  |
| `allow_registration` | boolean or null |  |
| `position` | integer or null | `position` is where the button sits. Left out, a new provider goes last and an existing one stays where it is. |

#### Response

Errors come back as an [OAuth error](/reference/errors#oauth-errors): `{"error", "error_description"}`.

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## PATCH /api/v1/admin/social-providers/:id {#patch-api-v1-admin-social-providers-id}

Changes a provider's settings.

Its kind and its identifier are not among them: both are in the address registered with the provider, and a sign-in that comes back to an address nobody answers is a worse failure than having to register a second provider.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`social.write`](/reference/permissions#social-write) |

```bash title="cURL"
curl -X PATCH "$ADMIN_API/api/v1/admin/social-providers/$ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/social-providers/" + id))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .header("Content-Type", "application/json")
        .method("PATCH", BodyPublishers.ofString("""
                {}
                """))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
body := strings.NewReader(`{}`)
req, err := http.NewRequest(http.MethodPatch, adminAPI+"/api/v1/admin/social-providers/"+id, body)
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
    f"{ADMIN_API}/api/v1/admin/social-providers/{id}",
    json={},
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
| `kind` | string | `kind` and `slug` are read when a provider is registered and never again: the slug is in the address registered with the provider, and the kind decides which addresses those are. |
| `slug` | string |  |
| `name` | string or null |  |
| `client_id` | string or null |  |
| `client_secret` | string or null |  |
| `private_key` | string or null |  |
| `team_id` | string or null |  |
| `key_id` | string or null |  |
| `scopes` | array of string or null |  |
| `token_auth` | string or null | `token_auth` is how the secret is presented at the token endpoint. An empty string leaves it to the kind's own default. |
| `authorize_url` | string or null |  |
| `token_url` | string or null |  |
| `userinfo_url` | string or null |  |
| `is_enabled` | boolean or null |  |
| `link_verified_emails` | boolean or null |  |
| `allow_registration` | boolean or null |  |
| `position` | integer or null | `position` is where the button sits. Left out, a new provider goes last and an existing one stays where it is. |

#### Response

Errors come back as an [OAuth error](/reference/errors#oauth-errors): `{"error", "error_description"}`.

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## DELETE /api/v1/admin/social-providers/:id {#delete-api-v1-admin-social-providers-id}

Removes a provider, and every identity held at it.

The accounts stay — a user who had no password keeps one way in fewer, which is why the panel says how many people that is before it asks.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`social.write`](/reference/permissions#social-write) |

```bash title="cURL"
curl -X DELETE "$ADMIN_API/api/v1/admin/social-providers/$ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/social-providers/" + id))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .DELETE()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodDelete, adminAPI+"/api/v1/admin/social-providers/"+id, nil)
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
    f"{ADMIN_API}/api/v1/admin/social-providers/{id}",
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
