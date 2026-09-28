---
title: "Single sign-on"
description: "Signing in through a customer's own identity provider, over OpenID Connect or SAML."
order: 12
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

## GET /api/v1/admin/sso-connections {#get-api-v1-admin-sso-connections}

Returns every connection, with how many people sign in through each.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`sso.read`](/reference/permissions#sso-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/sso-connections" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/sso-connections"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/sso-connections", nil)
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
    f"{ADMIN_API}/api/v1/admin/sso-connections",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Response

`200 OK`

```json
{
  "connections": [
    {
      "id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
      "created_at": "2026-09-28T09:30:00Z",
      "updated_at": "2026-09-28T09:30:00Z",
      "slug": "example",
      "name": "Example",
      "protocol": "oidc",
      "is_enabled": true,
      "domains": [
        "string"
      ],
      "enforce_domains": true,
      "show_on_login": true,
      "issuer": "https://example.com",
      "client_id": "shop-web",
      "scopes": [
        "openid profile email"
      ],
      "metadata_url": "https://example.com",
      "metadata": "string",
      "name_id_format": "email",
      "sign_requests": true,
      "sp_certificate": "string",
      "matching": "deny",
      "create_users": true,
      "sync_profile": true,
      "email_attribute": "ada@example.com",
      "first_name_attribute": "string",
      "last_name_attribute": "string",
      "groups_attribute": "string",
      "role_mappings": [
        {
          "group": "string",
          "role_id": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d"
        }
      ],
      "sync_roles": true,
      "has_client_secret": true,
      "users": 1,
      "service_provider": {
        "callback_url": "https://app.example.com/callback",
        "acs_url": "https://example.com",
        "entity_id": "string",
        "metadata_url": "https://example.com",
        "certificate": "string"
      },
      "identity_provider": {
        "entity_id": "string",
        "sso_url": "https://example.com",
        "certificates": 1,
        "certificate_expires": "2026-09-28T09:30:00Z"
      }
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

## GET /api/v1/admin/sso-connections/:id {#get-api-v1-admin-sso-connections-id}

Returns one connection.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`sso.read`](/reference/permissions#sso-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/sso-connections/$ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/sso-connections/" + id))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/sso-connections/"+id, nil)
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
    f"{ADMIN_API}/api/v1/admin/sso-connections/{id}",
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
| 404 | [`sso_connection_not_found`](/reference/errors#sso_connection_not_found) | There is no such connection. |

## POST /api/v1/admin/sso-connections {#post-api-v1-admin-sso-connections}

Adds a connection.

A SAML one is given its own key and certificate to sign requests with, which its provider is shown.

A new connection starts off: nobody signs in through it until an administrator has tried it and turned it on.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`sso.write`](/reference/permissions#sso-write) |

```bash title="cURL"
curl -X POST "$ADMIN_API/api/v1/admin/sso-connections" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/sso-connections"))
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
req, err := http.NewRequest(http.MethodPost, adminAPI+"/api/v1/admin/sso-connections", body)
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
    f"{ADMIN_API}/api/v1/admin/sso-connections",
    json={},
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `name` | string or null | Omitnil, at least 1 characters, at most 100 characters. |
| `slug` | string or null | Omitnil, at most 64 characters. |
| `protocol` | one of `oidc`, `saml` or null | Omitnil. |
| `is_enabled` | boolean or null |  |
| `domains` | array of string or null | Omitnil, at most 50 items. |
| `enforce_domains` | boolean or null |  |
| `show_on_login` | boolean or null |  |
| `issuer` | string or null | Omitnil, at most 512 characters. |
| `client_id` | string or null | Omitnil, at most 255 characters. |
| `client_secret` | string or null | Omitnil, at most 1024 characters. |
| `scopes` | array of string or null | Omitnil, at most 20 items. |
| `metadata_url` | string or null | Omitnil, at most 1024 characters. |
| `metadata` | string or null | Omitnil, at most 200000 characters. |
| `name_id_format` | one of `email`, `persistent`, `unspecified` or null | Omitnil. |
| `sign_requests` | boolean or null |  |
| `matching` | one of `link`, `deny` or null | Omitnil. |
| `create_users` | boolean or null |  |
| `sync_profile` | boolean or null |  |
| `sync_roles` | boolean or null |  |
| `email_attribute` | string or null | Omitnil, at most 255 characters. |
| `first_name_attribute` | string or null | Omitnil, at most 255 characters. |
| `last_name_attribute` | string or null | Omitnil, at most 255 characters. |
| `groups_attribute` | string or null | Omitnil, at most 255 characters. |
| `role_mappings` | array of RoleMapping or null | Omitnil, at most 100 items. |
| `role_mappings[].group` | string |  |
| `role_mappings[].role_id` | string |  |

#### Response

Errors come back as an [OAuth error](/reference/errors#oauth-errors): `{"error", "error_description"}`.

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 400 | [`sso_domain_invalid`](/reference/errors#sso_domain_invalid) | {domain} is not a domain. |
| 400 | [`sso_issuer_invalid`](/reference/errors#sso_issuer_invalid) | The issuer must be an https address. |
| 400 | [`sso_metadata_invalid`](/reference/errors#sso_metadata_invalid) | The SAML metadata could not be used: {reason} |
| 400 | [`sso_role_mapping_invalid`](/reference/errors#sso_role_mapping_invalid) | Every mapping needs a group and a role. |
| 400 | [`sso_role_unknown`](/reference/errors#sso_role_unknown) | A mapping names a role that does not exist. |
| 400 | [`sso_unreachable`](/reference/errors#sso_unreachable) | A connection without domains has to show its button on the sign-in page, and cannot be required: no address leads to it. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 409 | [`sso_domain_taken`](/reference/errors#sso_domain_taken) | {domain} already belongs to another connection. |
| 409 | [`sso_slug_taken`](/reference/errors#sso_slug_taken) | Another connection already uses the identifier {slug}. |

## POST /api/v1/admin/sso-connections/test {#post-api-v1-admin-sso-connections-test}

Tries a provider before it is relied on: an OpenID Connect issuer's discovery, or a SAML provider's metadata, from its address or as pasted.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`sso.write`](/reference/permissions#sso-write) |

```bash title="cURL"
curl -X POST "$ADMIN_API/api/v1/admin/sso-connections/test" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"protocol":"oidc"}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/sso-connections/test"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .header("Content-Type", "application/json")
        .POST(BodyPublishers.ofString("""
                {
                  "protocol": "oidc"
                }
                """))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
body := strings.NewReader(`{
	"protocol": "oidc"
}`)
req, err := http.NewRequest(http.MethodPost, adminAPI+"/api/v1/admin/sso-connections/test", body)
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
    f"{ADMIN_API}/api/v1/admin/sso-connections/test",
    json={
        "protocol": "oidc",
    },
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `protocol` <small>required</small> | one of `oidc`, `saml` |  |
| `issuer` | string | At most 512 characters. |
| `scopes` | array of string | At most 32 items, at most 128 items. |
| `metadata_url` | string | At most 1024 characters. |
| `metadata` | string | At most 200000 characters. |

#### Response

`200 OK`

```json
{
  "issuer": "https://example.com",
  "authorization_endpoint": "string",
  "token_endpoint": "string",
  "jwks_uri": "https://example.com",
  "unsupported_scopes": [
    "string"
  ],
  "identity_provider": {
    "entity_id": "string",
    "sso_url": "https://example.com",
    "certificates": 1,
    "certificate_expires": "2026-09-28T09:30:00Z"
  },
  "metadata": "string"
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 400 | [`sso_discovery_failed`](/reference/errors#sso_discovery_failed) | The issuer's discovery document could not be read: {reason} |
| 400 | [`sso_metadata_invalid`](/reference/errors#sso_metadata_invalid) | The SAML metadata could not be used: {reason} |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## PATCH /api/v1/admin/sso-connections/:id {#patch-api-v1-admin-sso-connections-id}

Changes a connection.

Its slug and protocol are not among them: both are in the addresses its provider was given.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`sso.write`](/reference/permissions#sso-write) |

```bash title="cURL"
curl -X PATCH "$ADMIN_API/api/v1/admin/sso-connections/$ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/sso-connections/" + id))
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
req, err := http.NewRequest(http.MethodPatch, adminAPI+"/api/v1/admin/sso-connections/"+id, body)
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
    f"{ADMIN_API}/api/v1/admin/sso-connections/{id}",
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
| `name` | string or null | Omitnil, at least 1 characters, at most 100 characters. |
| `slug` | string or null | Omitnil, at most 64 characters. |
| `protocol` | one of `oidc`, `saml` or null | Omitnil. |
| `is_enabled` | boolean or null |  |
| `domains` | array of string or null | Omitnil, at most 50 items. |
| `enforce_domains` | boolean or null |  |
| `show_on_login` | boolean or null |  |
| `issuer` | string or null | Omitnil, at most 512 characters. |
| `client_id` | string or null | Omitnil, at most 255 characters. |
| `client_secret` | string or null | Omitnil, at most 1024 characters. |
| `scopes` | array of string or null | Omitnil, at most 20 items. |
| `metadata_url` | string or null | Omitnil, at most 1024 characters. |
| `metadata` | string or null | Omitnil, at most 200000 characters. |
| `name_id_format` | one of `email`, `persistent`, `unspecified` or null | Omitnil. |
| `sign_requests` | boolean or null |  |
| `matching` | one of `link`, `deny` or null | Omitnil. |
| `create_users` | boolean or null |  |
| `sync_profile` | boolean or null |  |
| `sync_roles` | boolean or null |  |
| `email_attribute` | string or null | Omitnil, at most 255 characters. |
| `first_name_attribute` | string or null | Omitnil, at most 255 characters. |
| `last_name_attribute` | string or null | Omitnil, at most 255 characters. |
| `groups_attribute` | string or null | Omitnil, at most 255 characters. |
| `role_mappings` | array of RoleMapping or null | Omitnil, at most 100 items. |
| `role_mappings[].group` | string |  |
| `role_mappings[].role_id` | string |  |

#### Response

Errors come back as an [OAuth error](/reference/errors#oauth-errors): `{"error", "error_description"}`.

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 400 | [`sso_domain_invalid`](/reference/errors#sso_domain_invalid) | {domain} is not a domain. |
| 400 | [`sso_issuer_invalid`](/reference/errors#sso_issuer_invalid) | The issuer must be an https address. |
| 400 | [`sso_metadata_invalid`](/reference/errors#sso_metadata_invalid) | The SAML metadata could not be used: {reason} |
| 400 | [`sso_role_mapping_invalid`](/reference/errors#sso_role_mapping_invalid) | Every mapping needs a group and a role. |
| 400 | [`sso_role_unknown`](/reference/errors#sso_role_unknown) | A mapping names a role that does not exist. |
| 400 | [`sso_unreachable`](/reference/errors#sso_unreachable) | A connection without domains has to show its button on the sign-in page, and cannot be required: no address leads to it. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 404 | [`sso_connection_not_found`](/reference/errors#sso_connection_not_found) | There is no such connection. |
| 409 | [`sso_domain_taken`](/reference/errors#sso_domain_taken) | {domain} already belongs to another connection. |

## DELETE /api/v1/admin/sso-connections/:id {#delete-api-v1-admin-sso-connections-id}

Removes a connection and the identities held at it.

The accounts stay, and sign in however else they can.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`sso.write`](/reference/permissions#sso-write) |

```bash title="cURL"
curl -X DELETE "$ADMIN_API/api/v1/admin/sso-connections/$ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/sso-connections/" + id))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .DELETE()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodDelete, adminAPI+"/api/v1/admin/sso-connections/"+id, nil)
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
    f"{ADMIN_API}/api/v1/admin/sso-connections/{id}",
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
| 404 | [`sso_connection_not_found`](/reference/errors#sso_connection_not_found) | There is no such connection. |

## POST /api/v1/admin/sso-connections/:id/refresh-metadata {#post-api-v1-admin-sso-connections-id-refresh-metadata}

Reads a SAML provider's metadata again from its address — for when it has rolled its certificate over.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`sso.write`](/reference/permissions#sso-write) |

```bash title="cURL"
curl -X POST "$ADMIN_API/api/v1/admin/sso-connections/$ID/refresh-metadata" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/sso-connections/" + id + "/refresh-metadata"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .POST(BodyPublishers.noBody())
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodPost, adminAPI+"/api/v1/admin/sso-connections/"+id+"/refresh-metadata", nil)
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
response = requests.post(
    f"{ADMIN_API}/api/v1/admin/sso-connections/{id}/refresh-metadata",
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
| 400 | [`sso_metadata_invalid`](/reference/errors#sso_metadata_invalid) | The SAML metadata could not be used: {reason} |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 404 | [`sso_connection_not_found`](/reference/errors#sso_connection_not_found) | There is no such connection. |
