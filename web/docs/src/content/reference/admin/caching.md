---
title: "Cache"
description: "What the Redis cache holds, and clearing it."
order: 22
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

## GET /api/v1/admin/cache {#get-api-v1-admin-cache}

Sums up both databases: how many keys of each kind, each group's generation and what it holds, and the Redis server they share.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/cache" \
  -b "loginer_session=$SESSION"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/cache"))
        .header("Cookie", "loginer_session=" + SESSION)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/cache", nil)
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
    f"{ADMIN_API}/api/v1/admin/cache",
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Response

`200 OK`

```json
{
  "configured": true,
  "databases": [
    {
      "name": "Example",
      "number": 1,
      "available": true,
      "keys": 1,
      "kinds": {},
      "groups": [
        {
          "name": "Example",
          "generation": 1,
          "entries": 1,
          "stale": 1,
          "known": true
        }
      ],
      "sessions": {},
      "editable": true,
      "memory_bytes": 1,
      "version": "string"
    }
  ]
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## GET /api/v1/admin/cache/:database/keys {#get-api-v1-admin-cache-database-keys}

Lists a page of one database's keys.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/cache/$DATABASE/keys?kind=value&group=value&search=value" \
  -b "loginer_session=$SESSION"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/cache/" + database + "/keys" + "?kind=value&group=value&search=value"))
        .header("Cookie", "loginer_session=" + SESSION)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/cache/"+database+"/keys"+"?kind=value&group=value&search=value", nil)
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
    f"{ADMIN_API}/api/v1/admin/cache/{database}/keys",
    params={
        "kind": "value",
        "group": "value",
        "search": "value",
    },
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `database` <small>required</small> | path |  |
| `kind` | query |  |
| `group` | query |  |
| `search` | query |  |
| `cursor` | query |  |
| `limit` | query |  |

#### Response

`200 OK`

```json
{
  "keys": [
    {
      "name": "Example",
      "kind": "string",
      "group": "string",
      "entry": "string",
      "ttl_seconds": 1,
      "size_bytes": 1,
      "editable": true
    }
  ],
  "cursor": "string"
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`cache_cursor_invalid`](/reference/errors#cache_cursor_invalid) | The listing could not be continued. Start it again. |
| 400 | [`cache_kind_invalid`](/reference/errors#cache_kind_invalid) | There is no such kind of key. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 404 | [`cache_database_not_found`](/reference/errors#cache_database_not_found) | There is no such Redis database. |

## GET /api/v1/admin/cache/:database/key {#get-api-v1-admin-cache-database-key}

Answers one key and what it holds.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/cache/$DATABASE/key?name=Example" \
  -b "loginer_session=$SESSION"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/cache/" + database + "/key" + "?name=Example"))
        .header("Cookie", "loginer_session=" + SESSION)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/cache/"+database+"/key"+"?name=Example", nil)
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
    f"{ADMIN_API}/api/v1/admin/cache/{database}/key",
    params={
        "name": "Example",
    },
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `database` <small>required</small> | path |  |
| `name` | query |  |

#### Response

`200 OK`

```json
{
  "key": {
    "name": "Example",
    "kind": "string",
    "group": "string",
    "entry": "string",
    "ttl_seconds": 1,
    "size_bytes": 1,
    "editable": true,
    "type": "string",
    "value": "aGVsbG8="
  }
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`cache_key_invalid`](/reference/errors#cache_key_invalid) | That is not a key name. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 404 | [`cache_database_not_found`](/reference/errors#cache_database_not_found) | There is no such Redis database. |
| 404 | [`cache_key_not_found`](/reference/errors#cache_key_not_found) | That key is not in Redis any more. |
| 409 | [`cache_generation_key`](/reference/errors#cache_generation_key) | A generation counter cannot be removed. Clear its group instead. |
| 409 | [`cache_key_not_editable`](/reference/errors#cache_key_not_editable) | Only a cached value in the cache database can be edited. |

## PUT /api/v1/admin/cache/:database/key {#put-api-v1-admin-cache-database-key}

Replaces a cached value by hand.

Only a value of a group's current generation in the cache database can be: see cache.Write.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl -X PUT "$ADMIN_API/api/v1/admin/cache/$DATABASE/key" \
  -b "loginer_session=$SESSION" \
  -H "Content-Type: application/json" \
  -d '{"value":"aGVsbG8=","ttl_seconds":1}'
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/cache/" + database + "/key"))
        .header("Cookie", "loginer_session=" + SESSION)
        .header("Content-Type", "application/json")
        .PUT(BodyPublishers.ofString("""
                {
                  "value": "aGVsbG8=",
                  "ttl_seconds": 1
                }
                """))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
body := strings.NewReader(`{
	"value": "aGVsbG8=",
	"ttl_seconds": 1
}`)
req, err := http.NewRequest(http.MethodPut, adminAPI+"/api/v1/admin/cache/"+database+"/key", body)
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
    f"{ADMIN_API}/api/v1/admin/cache/{database}/key",
    json={
        "value": "aGVsbG8=",
        "ttl_seconds": 1,
    },
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `database` <small>required</small> | path |  |
| `name` | query |  |

#### Body

| Field | Type | Description |
| --- | --- | --- |
| `value` | string (byte) | `value` is the new value, as JSON of any shape. |
| `ttl_seconds` | integer | `ttl_seconds` is how long it should live; left out or zero keeps what it had left. |

#### Response

`200 OK`

```json
{
  "key": {
    "name": "Example",
    "kind": "string",
    "group": "string",
    "entry": "string",
    "ttl_seconds": 1,
    "size_bytes": 1,
    "editable": true,
    "type": "string",
    "value": "aGVsbG8="
  }
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`cache_key_invalid`](/reference/errors#cache_key_invalid) | That is not a key name. |
| 400 | [`cache_ttl_invalid`](/reference/errors#cache_ttl_invalid) | The time to live has to be between 0 and {max} seconds. |
| 400 | [`cache_value_invalid`](/reference/errors#cache_value_invalid) | The value has to be valid JSON. |
| 400 | [`invalid_body`](/reference/errors#invalid_body) | The request could not be read. Reload the page and try again. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 404 | [`cache_database_not_found`](/reference/errors#cache_database_not_found) | There is no such Redis database. |
| 404 | [`cache_key_not_found`](/reference/errors#cache_key_not_found) | That key is not in Redis any more. |
| 409 | [`cache_generation_key`](/reference/errors#cache_generation_key) | A generation counter cannot be removed. Clear its group instead. |
| 409 | [`cache_key_not_editable`](/reference/errors#cache_key_not_editable) | Only a cached value in the cache database can be edited. |

## DELETE /api/v1/admin/cache/:database/key {#delete-api-v1-admin-cache-database-key}

Removes one key.

The next request that wanted it reads the database again.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl -X DELETE "$ADMIN_API/api/v1/admin/cache/$DATABASE/key" \
  -b "loginer_session=$SESSION"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/cache/" + database + "/key"))
        .header("Cookie", "loginer_session=" + SESSION)
        .DELETE()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodDelete, adminAPI+"/api/v1/admin/cache/"+database+"/key", nil)
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
response = requests.delete(
    f"{ADMIN_API}/api/v1/admin/cache/{database}/key",
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `database` <small>required</small> | path |  |
| `name` | query |  |

#### Response

`204 No Content`

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`cache_key_invalid`](/reference/errors#cache_key_invalid) | That is not a key name. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 404 | [`cache_database_not_found`](/reference/errors#cache_database_not_found) | There is no such Redis database. |
| 404 | [`cache_key_not_found`](/reference/errors#cache_key_not_found) | That key is not in Redis any more. |
| 409 | [`cache_generation_key`](/reference/errors#cache_generation_key) | A generation counter cannot be removed. Clear its group instead. |
| 409 | [`cache_key_not_editable`](/reference/errors#cache_key_not_editable) | Only a cached value in the cache database can be edited. |

## POST /api/v1/admin/cache/:database/groups/:group/clear {#post-api-v1-admin-cache-database-groups-group-clear}

Forgets everything cached in one group, for every server process at once.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl -X POST "$ADMIN_API/api/v1/admin/cache/$DATABASE/groups/$GROUP/clear" \
  -b "loginer_session=$SESSION"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/cache/" + database + "/groups/" + group + "/clear"))
        .header("Cookie", "loginer_session=" + SESSION)
        .POST(BodyPublishers.noBody())
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodPost, adminAPI+"/api/v1/admin/cache/"+database+"/groups/"+group+"/clear", nil)
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
response = requests.post(
    f"{ADMIN_API}/api/v1/admin/cache/{database}/groups/{group}/clear",
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `database` <small>required</small> | path |  |
| `group` <small>required</small> | path |  |

#### Response

`204 No Content`

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 404 | [`cache_database_not_found`](/reference/errors#cache_database_not_found) | There is no such Redis database. |
| 404 | [`cache_group_not_found`](/reference/errors#cache_group_not_found) | There is no such cache group. |

## DELETE /api/v1/admin/cache/:database {#delete-api-v1-admin-cache-database}

Removes every key this server keeps in one database.

Flushing sessions signs nobody out, but resets every rate limit.

| | |
| --- | --- |
| Auth | Admin session · super admins only |

```bash title="cURL"
curl -X DELETE "$ADMIN_API/api/v1/admin/cache/$DATABASE" \
  -b "loginer_session=$SESSION"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/cache/" + database))
        .header("Cookie", "loginer_session=" + SESSION)
        .DELETE()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodDelete, adminAPI+"/api/v1/admin/cache/"+database, nil)
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
response = requests.delete(
    f"{ADMIN_API}/api/v1/admin/cache/{database}",
    cookies={"loginer_session": SESSION},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `database` <small>required</small> | path |  |

#### Response

`204 No Content`

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 403 | [`cross_origin`](/reference/errors#cross_origin) | Requests from {origin} are not allowed. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
| 404 | [`cache_database_not_found`](/reference/errors#cache_database_not_found) | There is no such Redis database. |
