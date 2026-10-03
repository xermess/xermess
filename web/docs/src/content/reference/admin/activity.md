---
title: "Activity"
description: "The dashboard's counts and the activity log."
order: 6
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

## GET /api/v1/admin/overview {#get-api-v1-admin-overview}

Is the panel's front page for a range of days: totals, sign-in outcomes, daily activity, the busiest people and the latest entries.

Its five reads run concurrently.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`activity.read`](/reference/permissions#activity-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/overview" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/overview"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/overview", nil)
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
    f"{ADMIN_API}/api/v1/admin/overview",
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Response

`200 OK`

```json
{
  "days": 1,
  "counts": {
    "users": 1,
    "active_users": 1,
    "new_users": 1,
    "applications": 1,
    "enabled_applications": 1,
    "apis": 1,
    "api_scopes": 1,
    "user_roles": 1,
    "admins": 1,
    "locked_admins": 1,
    "active_sessions": 1
  },
  "sign_ins": {
    "since": "2026-09-28T09:30:00Z",
    "succeeded": 1,
    "failed": 1,
    "blocked": 1
  },
  "daily": [
    {
      "day": "string",
      "events": 1,
      "failures": 1
    }
  ],
  "top_actors": [
    {
      "actor": "string",
      "events": 1
    }
  ],
  "activity": [
    {
      "id": "string",
      "action": "string",
      "actor": "string",
      "ip": "string",
      "target": {
        "type": "string",
        "id": "string",
        "name": "Example"
      },
      "detail": "string",
      "created_at": "2026-09-28T09:30:00Z"
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

## GET /api/v1/admin/logs {#get-api-v1-admin-logs}

Lists one page of the activity log, newest first, narrowed by the filters parseFilter reads, with the cursor the next page starts from.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`activity.read`](/reference/permissions#activity-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/logs?q=value&actor=value&action=value" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/logs" + "?q=value&actor=value&action=value"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/logs"+"?q=value&actor=value&action=value", nil)
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
    f"{ADMIN_API}/api/v1/admin/logs",
    params={
        "q": "value",
        "actor": "value",
        "action": "value",
    },
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `q` | query |  |
| `actor` | query |  |
| `action` | query |  |
| `from` | query |  |
| `to` | query |  |
| `before` | query |  |

#### Response

`200 OK`

```json
{
  "logs": [
    {
      "id": "string",
      "action": "string",
      "actor": "string",
      "ip": "string",
      "target": {
        "type": "string",
        "id": "string",
        "name": "Example"
      },
      "detail": "string",
      "created_at": "2026-09-28T09:30:00Z",
      "user_agent": "string"
    }
  ],
  "next": "string"
}
```

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`logs_filter_invalid`](/reference/errors#logs_filter_invalid) | These filters cannot be applied. Check the dates and try again. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |

## GET /api/v1/admin/logs/export {#get-api-v1-admin-logs-export}

Writes matching entries as CSV, newest first, up to exportLimit, hiding exactly what the logs page hides.

| | |
| --- | --- |
| Auth | Admin session or [admin-cli token](/guides/admin-api) · needs [`activity.read`](/reference/permissions#activity-read) |

```bash title="cURL"
curl "$ADMIN_API/api/v1/admin/logs/export?q=value&actor=value&action=value" \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ADMIN_API + "/api/v1/admin/logs/export" + "?q=value&actor=value&action=value"))
        .header("Authorization", "Bearer " + ADMIN_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, adminAPI+"/api/v1/admin/logs/export"+"?q=value&actor=value&action=value", nil)
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
    f"{ADMIN_API}/api/v1/admin/logs/export",
    params={
        "q": "value",
        "actor": "value",
        "action": "value",
    },
    headers={"Authorization": f"Bearer {ADMIN_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `q` | query |  |
| `actor` | query |  |
| `action` | query |  |
| `from` | query |  |
| `to` | query |  |
| `before` | query |  |

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 400 | [`logs_filter_invalid`](/reference/errors#logs_filter_invalid) | These filters cannot be applied. Check the dates and try again. |
| 401 | [`not_signed_in`](/reference/errors#not_signed_in) | You are not signed in. |
| 401 | [`token_refused`](/reference/errors#token_refused) | This access token is not accepted by the admin API: it has expired, is for another API, is not a service's token, or its application has lost access. |
| 403 | [`forbidden`](/reference/errors#forbidden) | You do not have permission to do that. |
