---
title: "Signing keys"
description: "The keys tokens are signed with, and rotating them."
order: 21
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

## GET /api/v1/admin/signing-keys {#get-api-v1-admin-signing-keys}

Describes every published key: the one signing, the next one waiting, and the retired ones still published.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/signing-keys" \
  -b "loginer_session=$SESSION"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/signing-keys"))
        .header("Cookie", "loginer_session=" + SESSION)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/signing-keys", nil)
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
    f"{ADMIN_API}/api/v1/admin/signing-keys",
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Response

`200 OK`

```json
{
  "keys": [
    {
      "kid": "string",
      "algorithm": "string",
      "state": "string",
      "created_at": "2026-09-28T09:30:00Z",
      "signs_from": "2026-09-28T09:30:00Z",
      "retired_at": "2026-09-28T09:30:00Z",
      "deleted_at": "2026-09-28T09:30:00Z"
    }
  ]
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## POST /api/v1/admin/signing-keys/rotate {#post-api-v1-admin-signing-keys-rotate}

Makes new keys now.

| | |
| --- | --- |
| Auth | Admin session · super admins only |
| Rate limit | Per IP address |

```bash title="cURL"
curl -X POST "$ADMIN_API/api/v1/admin/signing-keys/rotate" \
  -b "loginer_session=$SESSION" \
  -H "Content-Type: application/json" \
  -d '{"immediate":true,"revoke_old":true}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/signing-keys/rotate"))
        .header("Cookie", "loginer_session=" + SESSION)
        .header("Content-Type", "application/json")
        .POST(BodyPublishers.ofString("""
                {
                  "immediate": true,
                  "revoke_old": true
                }
                """))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
body := strings.NewReader(`{
	"immediate": true,
	"revoke_old": true
}`)
req, err := http.NewRequest(http.MethodPost, adminAPI+"/api/v1/admin/signing-keys/rotate", body)
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
response = requests.post(
    f"{ADMIN_API}/api/v1/admin/signing-keys/rotate",
    json={
        "immediate": True,
        "revoke_old": True,
    },
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `immediate` | boolean | `immediate` makes the new keys sign now, and retires the old ones. |
| `revoke_old` | boolean | `revoke_old` deletes the old keys as well, so APIs stop trusting every token they signed. It implies `immediate`. |

#### Response

`200 OK`

```json
{
  "keys": [
    {
      "kid": "string",
      "algorithm": "string",
      "state": "string",
      "created_at": "2026-09-28T09:30:00Z",
      "signs_from": "2026-09-28T09:30:00Z",
      "retired_at": "2026-09-28T09:30:00Z",
      "deleted_at": "2026-09-28T09:30:00Z"
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
| 429 | [`rate_limited`](/reference/errors#rate_limited) | Too many attempts. Try again in {seconds} seconds. |
