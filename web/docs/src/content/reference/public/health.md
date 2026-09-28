---
title: "Health"
description: "Whether the server is up, for load balancers and uptime checks."
order: 1
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

## GET /healthz {#get-healthz}

Says the server is up.

| | |
| --- | --- |
| Auth | None |

```bash title="cURL"
curl "$ISSUER/healthz"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/healthz"))
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, issuer+"/healthz", nil)
if err != nil {
	return err
}

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.get(f"{ISSUER}/healthz")
response.raise_for_status()
```

#### Response

`200 OK`

```json
{
  "status": "string"
}
```
