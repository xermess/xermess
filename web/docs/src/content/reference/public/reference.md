---
title: "OpenAPI document"
description: "The server's own OpenAPI document, for your tooling."
order: 3
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

## GET /.well-known/openapi.json {#get-well-known-openapi-json}

Answers the OpenAPI document.

| | |
| --- | --- |
| Auth | None |

```bash title="cURL"
curl "$ISSUER/.well-known/openapi.json"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/.well-known/openapi.json"))
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, issuer+"/.well-known/openapi.json", nil)
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
response = requests.get(f"{ISSUER}/.well-known/openapi.json")
response.raise_for_status()
```

#### Response

`200 OK`
