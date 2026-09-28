---
title: "Organization"
description: "The organisation's name, branding and agreements."
order: 10
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

## GET /api/v1/admin/organization {#get-api-v1-admin-organization}

Returns the organisation.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`organization.read`](/reference/permissions#organization-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/organization" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/organization"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/organization", nil)
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
    f"{ADMIN_API}/api/v1/admin/organization",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Response

`200 OK`

```json
{
  "organization": {
    "name": "Example",
    "slug": "example",
    "logo_url": "https://example.com/logo.png",
    "support_email": "ada@example.com",
    "support_phone": "+1 555 0100",
    "terms_url": "https://example.com",
    "privacy_url": "https://example.com",
    "timezone": "Europe/London",
    "created_at": "2026-09-28T09:30:00Z"
  }
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## PATCH /api/v1/admin/organization {#patch-api-v1-admin-organization}

Changes the settings.

What a request leaves out is left as it is, so a client that knows about one field does not clear the rest.

| | |
| --- | --- |
| Auth | Admin session · needs [`organization.write`](/reference/permissions#organization-write) |

```bash title="cURL"
curl -X PATCH "$ADMIN_API/api/v1/admin/organization" \
  -b "loginer_session=$SESSION" \
  -H "Content-Type: application/json" \
  -d '{}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/organization"))
        .header("Cookie", "loginer_session=" + SESSION)
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
req, err := http.NewRequest(http.MethodPatch, adminAPI+"/api/v1/admin/organization", body)
if err != nil {
	return err
}
req.Header.Set("Content-Type", "application/json")
req.AddCookie(&http.Cookie{Name: "loginer_session", Value: session})

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.patch(
    f"{ADMIN_API}/api/v1/admin/organization",
    json={},
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `name` | string or null |  |
| `slug` | string or null |  |
| `logo_url` | string or null |  |
| `support_email` | string or null |  |
| `support_phone` | string or null |  |
| `terms_url` | string or null |  |
| `privacy_url` | string or null |  |
| `timezone` | string or null |  |

#### Response

`200 OK`

```json
{
  "organization": {
    "name": "Example",
    "slug": "example",
    "logo_url": "https://example.com/logo.png",
    "support_email": "ada@example.com",
    "support_phone": "+1 555 0100",
    "terms_url": "https://example.com",
    "privacy_url": "https://example.com",
    "timezone": "Europe/London",
    "created_at": "2026-09-28T09:30:00Z"
  }
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
