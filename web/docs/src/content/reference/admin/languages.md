---
title: "Languages"
description: "The sign-in pages' languages and their text."
order: 14
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

## GET /api/v1/admin/languages {#get-api-v1-admin-languages}

Returns every language, with how much of each app it translates, and the shipped languages this installation does not have.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`languages.read`](/reference/permissions#languages-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/languages" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/languages"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/languages", nil)
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
    f"{ADMIN_API}/api/v1/admin/languages",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Response

`200 OK`

```json
{
  "languages": [
    {
      "code": "string",
      "name": "Example",
      "native_name": "string",
      "apps": [
        "string"
      ],
      "coverage": {},
      "missing": {},
      "is_enabled": true,
      "is_default": true,
      "position": 1,
      "base": true,
      "shipped": true,
      "updated_at": "2026-09-28T09:30:00Z"
    }
  ],
  "apps": [
    "string"
  ],
  "shipped": [
    {
      "code": "string",
      "name": "Example",
      "native_name": "string"
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

## GET /api/v1/admin/languages/:code/translations/:app {#get-api-v1-admin-languages-code-translations-app}

Returns one language's text for one app, for the editor.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`languages.read`](/reference/permissions#languages-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/languages/$CODE/translations/$APP" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/languages/" + code + "/translations/" + app))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/languages/"+code+"/translations/"+app, nil)
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
    f"{ADMIN_API}/api/v1/admin/languages/{code}/translations/{app}",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `code` <small>required</small> | path |  |
| `app` <small>required</small> | path |  |

#### Response

`200 OK`

```json
{
  "app": "string",
  "keys": [
    "string"
  ],
  "base": {},
  "messages": {}
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 404 | [`language_not_found`](/reference/errors#language_not_found) | There is no such language. |
| 404 | [`translation_app_not_found`](/reference/errors#translation_app_not_found) | There is no such part of the product to translate. |

## POST /api/v1/admin/languages {#post-api-v1-admin-languages}

Adds a language, with the text it starts from (createRequest).

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`languages.write`](/reference/permissions#languages-write) |

```bash title="cURL"
curl -X POST "$ADMIN_API/api/v1/admin/languages" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code":"string","name":"Example","native_name":"string"}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/languages"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .header("Content-Type", "application/json")
        .POST(BodyPublishers.ofString("""
                {
                  "code": "string",
                  "name": "Example",
                  "native_name": "string"
                }
                """))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
body := strings.NewReader(`{
	"code": "string",
	"name": "Example",
	"native_name": "string"
}`)
req, err := http.NewRequest(http.MethodPost, adminAPI+"/api/v1/admin/languages", body)
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
    f"{ADMIN_API}/api/v1/admin/languages",
    json={
        "code": "string",
        "name": "Example",
        "native_name": "string",
    },
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `code` <small>required</small> | string | At most 16 characters, languagetag. |
| `name` <small>required</small> | string | At most 64 characters. |
| `native_name` <small>required</small> | string | At most 64 characters. |
| `is_enabled` | boolean or null |  |
| `is_default` | boolean or null |  |
| `copy_from` | string |  |

#### Response

Errors come back as an [OAuth error](/reference/errors#oauth-errors): `{"error", "error_description"}`.

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 400 | [`language_copy_missing`](/reference/errors#language_copy_missing) | There is no language {code} to copy. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 409 | [`language_code_taken`](/reference/errors#language_code_taken) | There is already a language with this code. |

## PATCH /api/v1/admin/languages/:code {#patch-api-v1-admin-languages-code}

Changes a language's names and whether and where it is offered.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`languages.write`](/reference/permissions#languages-write) |

```bash title="cURL"
curl -X PATCH "$ADMIN_API/api/v1/admin/languages/$CODE" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/languages/" + code))
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
req, err := http.NewRequest(http.MethodPatch, adminAPI+"/api/v1/admin/languages/"+code, body)
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
    f"{ADMIN_API}/api/v1/admin/languages/{code}",
    json={},
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
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
| `name` | string or null | Omitnil, at least 1 characters, at most 64 characters. |
| `native_name` | string or null | Omitnil, at least 1 characters, at most 64 characters. |
| `is_enabled` | boolean or null |  |
| `is_default` | boolean or null |  |
| `position` | integer or null | Omitnil, at least 0. |

#### Response

Errors come back as an [OAuth error](/reference/errors#oauth-errors): `{"error", "error_description"}`.

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 400 | [`language_default_required`](/reference/errors#language_default_required) | Make another language the default rather than unmarking this one. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 404 | [`language_not_found`](/reference/errors#language_not_found) | There is no such language. |

## DELETE /api/v1/admin/languages/:code {#delete-api-v1-admin-languages-code}

Removes a language and its text.

Anybody who had chosen it is shown the default from their next page on.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`languages.write`](/reference/permissions#languages-write) |

```bash title="cURL"
curl -X DELETE "$ADMIN_API/api/v1/admin/languages/$CODE" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/languages/" + code))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .DELETE()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodDelete, adminAPI+"/api/v1/admin/languages/"+code, nil)
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
    f"{ADMIN_API}/api/v1/admin/languages/{code}",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `code` <small>required</small> | path |  |

#### Response

`204 No Content`

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`language_base_protected`](/reference/errors#language_base_protected) | English is what every other language falls back to, so it stays. |
| 400 | [`language_default_protected`](/reference/errors#language_default_protected) | Make another language the default before removing this one. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 404 | [`language_not_found`](/reference/errors#language_not_found) | There is no such language. |

## PUT /api/v1/admin/languages/:code/translations/:app {#put-api-v1-admin-languages-code-translations-app}

Replaces one language's text for one app.

Keys the app does not look up are dropped and counted rather than refused, so a file from an older release still imports.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`languages.write`](/reference/permissions#languages-write) |

```bash title="cURL"
curl -X PUT "$ADMIN_API/api/v1/admin/languages/$CODE/translations/$APP" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"messages":{}}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/languages/" + code + "/translations/" + app))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
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
req, err := http.NewRequest(http.MethodPut, adminAPI+"/api/v1/admin/languages/"+code+"/translations/"+app, body)
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
response = requests.put(
    f"{ADMIN_API}/api/v1/admin/languages/{code}/translations/{app}",
    json={
        "messages": {},
    },
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `code` <small>required</small> | path |  |
| `app` <small>required</small> | path |  |

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `messages` <small>required</small> | map of string |  |

#### Response

`200 OK`

```json
{
  "language": {
    "code": "string",
    "name": "Example",
    "native_name": "string",
    "apps": [
      "string"
    ],
    "coverage": {},
    "missing": {},
    "is_enabled": true,
    "is_default": true,
    "position": 1,
    "base": true,
    "shipped": true,
    "updated_at": "2026-09-28T09:30:00Z"
  },
  "ignored": 1
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 400 | [`translation_too_long`](/reference/errors#translation_too_long) | The text of {key} must be at most {max} characters. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 404 | [`language_not_found`](/reference/errors#language_not_found) | There is no such language. |
| 404 | [`translation_app_not_found`](/reference/errors#translation_app_not_found) | There is no such part of the product to translate. |
