---
title: "Mail"
description: "How email is sent, and what each message says."
order: 19
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

## GET /api/v1/admin/mail {#get-api-v1-admin-mail}

Returns the mail settings.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/mail" \
  -b "loginer_session=$SESSION"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/mail"))
        .header("Cookie", "loginer_session=" + SESSION)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/mail", nil)
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
    f"{ADMIN_API}/api/v1/admin/mail",
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Response

`200 OK`

```json
{
  "mail": {
    "is_enabled": true,
    "host": "string",
    "port": 1,
    "encryption": "string",
    "username": "string",
    "from_address": "string",
    "from_name": "string",
    "has_password": true
  },
  "encryptions": [
    "none"
  ]
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## PATCH /api/v1/admin/mail {#patch-api-v1-admin-mail}

Changes the settings.

What a request leaves out is left as it is, so a client that knows about one field does not clear the rest — and the password is left as it is unless one was typed, since nothing ever reads the stored one back to send it here again.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl -X PATCH "$ADMIN_API/api/v1/admin/mail" \
  -b "loginer_session=$SESSION" \
  -H "Content-Type: application/json" \
  -d '{}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/mail"))
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
req, err := http.NewRequest(http.MethodPatch, adminAPI+"/api/v1/admin/mail", body)
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
    f"{ADMIN_API}/api/v1/admin/mail",
    json={},
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `is_enabled` | boolean or null |  |
| `host` | string or null |  |
| `port` | integer or null |  |
| `encryption` | string or null |  |
| `username` | string or null |  |
| `password` | string or null |  |
| `from_address` | string or null |  |
| `from_name` | string or null |  |

#### Response

`200 OK`

```json
{
  "mail": {
    "is_enabled": true,
    "host": "string",
    "port": 1,
    "encryption": "string",
    "username": "string",
    "from_address": "string",
    "from_name": "string",
    "has_password": true
  },
  "encryptions": [
    "none"
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

## POST /api/v1/admin/mail/test {#post-api-v1-admin-mail-test}

Sends one message, to find out whether the settings work before anybody's sign-in depends on them.

It sends with what the form holds rather than with what is stored, since the point is to try a server before saving it — except for the password, which the form only holds when one has just been typed. Left out, the stored one is used, so a working server can be tested again without typing its password each time.

| | |
| --- | --- |
| Auth | Admin session · super admins only |
| Rate limit | Per IP address |

```bash title="cURL"
curl -X POST "$ADMIN_API/api/v1/admin/mail/test" \
  -b "loginer_session=$SESSION" \
  -H "Content-Type: application/json" \
  -d '{}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/mail/test"))
        .header("Cookie", "loginer_session=" + SESSION)
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
req, err := http.NewRequest(http.MethodPost, adminAPI+"/api/v1/admin/mail/test", body)
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
    f"{ADMIN_API}/api/v1/admin/mail/test",
    json={},
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `to` | string |  |
| `is_enabled` | boolean or null |  |
| `host` | string or null |  |
| `port` | integer or null |  |
| `encryption` | string or null |  |
| `username` | string or null |  |
| `password` | string or null |  |
| `from_address` | string or null |  |
| `from_name` | string or null |  |

#### Response

`200 OK`

```json
{
  "sent": true,
  "to": "string"
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 400 | [`mail_test_needs_recipient`](/reference/errors#mail_test_needs_recipient) | Say where the test message should go. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 429 | [`rate_limited`](/reference/errors#rate_limited) | Too many attempts. Try again in {seconds} seconds. |

## GET /api/v1/admin/mail/content {#get-api-v1-admin-mail-content}

Returns the words of every email the server sends, for each language the sign-in pages are offered in: what the language says, and what English says underneath it.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/mail/content" \
  -b "loginer_session=$SESSION"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/mail/content"))
        .header("Cookie", "loginer_session=" + SESSION)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/mail/content", nil)
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
    f"{ADMIN_API}/api/v1/admin/mail/content",
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Response

`200 OK`

```json
{
  "messages": [
    {
      "kind": "string",
      "label": "string",
      "description": "string",
      "subject_key": "string",
      "body_key": "string",
      "params": [
        {
          "name": "Example",
          "description": "string"
        }
      ]
    }
  ],
  "languages": [
    {
      "code": "string",
      "name": "Example",
      "native_name": "string",
      "offered": true,
      "messages": {},
      "base": {}
    }
  ]
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## PUT /api/v1/admin/mail/content/:code {#put-api-v1-admin-mail-content-code}

Writes one language's words for the emails.

Only the keys the messages are made of: this page holds a handful of a language's text and has the rest nowhere, so it merges rather than replacing (store.SaveTranslationKeys), and a key that is not an email's is refused rather than quietly dropped — a page sending one is a page with a bug, not an old file being imported.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl -X PUT "$ADMIN_API/api/v1/admin/mail/content/$CODE" \
  -b "loginer_session=$SESSION" \
  -H "Content-Type: application/json" \
  -d '{"messages":{}}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/mail/content/" + code))
        .header("Cookie", "loginer_session=" + SESSION)
        .header("Content-Type", "application/json")
        .PUT(BodyPublishers.ofString("""
                {
                  "messages": {}
                }
                """))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
body := strings.NewReader(`{
	"messages": {}
}`)
req, err := http.NewRequest(http.MethodPut, adminAPI+"/api/v1/admin/mail/content/"+code, body)
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
response = requests.put(
    f"{ADMIN_API}/api/v1/admin/mail/content/{code}",
    json={
        "messages": {},
    },
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `code` <small>required</small> | path |  |

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `messages` | map of string |  |

#### Response

`200 OK`

```json
{
  "language": {
    "code": "string",
    "name": "Example",
    "native_name": "string",
    "offered": true,
    "messages": {},
    "base": {}
  }
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 400 | [`mail_content_key_unknown`](/reference/errors#mail_content_key_unknown) | {key} is not part of any email this server sends. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 404 | [`language_not_found`](/reference/errors#language_not_found) | There is no such language. |
