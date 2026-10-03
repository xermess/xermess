---
title: "One-time codes"
description: "How the one-time codes sent by email behave."
order: 20
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

## GET /api/v1/admin/otp {#get-api-v1-admin-otp}

Returns the settings and the login flows that use emailed codes.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/otp" \
  -b "loginer_session=$SESSION"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/otp"))
        .header("Cookie", "loginer_session=" + SESSION)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/otp", nil)
if err != nil {
	return err
}
req.AddCookie(&http.Cookie{Name: "loginer_session", Value: session})

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.get(
    f"{ADMIN_API}/api/v1/admin/otp",
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Response

`200 OK`

```json
{
  "otp": {
    "code_length": 1,
    "lifetime_minutes": 3600,
    "max_attempts": 1,
    "resend_seconds": 1
  },
  "limits": {
    "min_length": 1,
    "max_length": 1,
    "max_lifetime_minutes": 3600,
    "max_attempts": 1,
    "max_resend_seconds": 1
  },
  "flows": [
    {
      "id": "string",
      "name": "Example",
      "slug": "example",
      "is_enabled": true,
      "is_default": true
    }
  ]
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## PATCH /api/v1/admin/otp {#patch-api-v1-admin-otp}

Changes the settings.

What a request leaves out is left as it is.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl -X PATCH "$ADMIN_API/api/v1/admin/otp" \
  -b "loginer_session=$SESSION" \
  -H "Content-Type: application/json" \
  -d '{"code_length":1,"lifetime_minutes":3600,"max_attempts":1,"resend_seconds":1}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/otp"))
        .header("Cookie", "loginer_session=" + SESSION)
        .header("Content-Type", "application/json")
        .method("PATCH", BodyPublishers.ofString("""
                {
                  "code_length": 1,
                  "lifetime_minutes": 3600,
                  "max_attempts": 1,
                  "resend_seconds": 1
                }
                """))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
body := strings.NewReader(`{
	"code_length": 1,
	"lifetime_minutes": 3600,
	"max_attempts": 1,
	"resend_seconds": 1
}`)
req, err := http.NewRequest(http.MethodPatch, adminAPI+"/api/v1/admin/otp", body)
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
    f"{ADMIN_API}/api/v1/admin/otp",
    json={
        "code_length": 1,
        "lifetime_minutes": 3600,
        "max_attempts": 1,
        "resend_seconds": 1,
    },
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `code_length` | integer or null |  |
| `lifetime_minutes` | integer or null |  |
| `max_attempts` | integer or null |  |
| `resend_seconds` | integer or null |  |

#### Response

`200 OK`

```json
{
  "otp": {
    "code_length": 1,
    "lifetime_minutes": 3600,
    "max_attempts": 1,
    "resend_seconds": 1
  },
  "limits": {
    "min_length": 1,
    "max_length": 1,
    "max_lifetime_minutes": 3600,
    "max_attempts": 1,
    "max_resend_seconds": 1
  },
  "flows": [
    {
      "id": "string",
      "name": "Example",
      "slug": "example",
      "is_enabled": true,
      "is_default": true
    }
  ]
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
