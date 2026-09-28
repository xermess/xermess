---
title: Machine to machine
description: A service calling an API as itself, with the client credentials grant.
order: 2
section: Protect APIs
nav: 'Service to service'
icon: robot-2
---

A service with no user — a nightly job, one backend calling another — authenticates as itself with the **client credentials** grant. There is no browser and no sign-in page: it posts its credentials to the token endpoint and gets an access token for an API.

## Set it up

1. In the panel, register an application of type **machine-to-machine**. It gets a secret and the `client_credentials` grant, and nothing else.
2. Under **APIs**, open the API it will call and authorize the application, choosing the scopes it may have.

## Get a token

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/token" \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -d "grant_type=client_credentials" \
  -d "audience=https://api.example.com" \
  -d "scope=orders:read"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder(URI.create(discovery.tokenEndpoint()))
        .header("Authorization", "Basic " + Base64.getEncoder()
                .encodeToString((CLIENT_ID + ":" + CLIENT_SECRET).getBytes(UTF_8)))
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(BodyPublishers.ofString("grant_type=client_credentials"
                + "&audience=" + URLEncoder.encode("https://api.example.com", UTF_8)
                + "&scope=" + URLEncoder.encode("orders:read", UTF_8)))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient().send(request, BodyHandlers.ofString());
```

```go title="Go"
// golang.org/x/oauth2/clientcredentials
config := clientcredentials.Config{
	ClientID:       clientID,
	ClientSecret:   clientSecret,
	TokenURL:       discovery.TokenEndpoint,
	Scopes:         []string{"orders:read"},
	EndpointParams: url.Values{"audience": {"https://api.example.com"}},
}
client := config.Client(ctx) // asks for a new token as each one expires
```

```python title="Python"
response = requests.post(
    discovery["token_endpoint"],
    data={
        "grant_type": "client_credentials",
        "audience": "https://api.example.com",
        "scope": "orders:read",
    },
    auth=(CLIENT_ID, CLIENT_SECRET),
)
response.raise_for_status()
access_token = response.json()["access_token"]
```

```json
{
	"access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6ImF0K2p3dCJ9…",
	"token_type": "Bearer",
	"expires_in": 3600,
	"scope": "orders:read"
}
```

There is no refresh token and no ID token: when the access token expires, ask for another. Cache it until then — every request to the token endpoint counts against a rate limit, ten times the sign-in pages' but still a limit.

The token's `sub` is the application's client ID, since there is no user. An API serving both users and services tells them apart by that: a user's `sub` is a UUID.

## When it is refused

| `error`               | Means                                                                                                     |
| --------------------- | --------------------------------------------------------------------------------------------------------- |
| `invalid_client`      | Wrong client ID or secret, or the credentials sent a way the application was not registered for           |
| `unauthorized_client` | The application lacks the grant, is a public client, or may not have a token for that API or those scopes |
| `invalid_request`     | `audience` names no API                                                                                   |
