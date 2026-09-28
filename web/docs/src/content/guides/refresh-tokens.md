---
title: Refresh tokens
description: Keep a user signed in past the access token's hour, and what rotation means for you.
order: 3
section: Sign in
nav: 'Refresh tokens'
icon: refresh
---

Ask for the `offline_access` scope, on an application with the `refresh_token` grant, and the token response carries a `refresh_token`. Trade it for new tokens when the access token expires:

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/token" \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -d "grant_type=refresh_token" \
  -d "refresh_token=$REFRESH_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder(URI.create(discovery.tokenEndpoint()))
        .header("Authorization", "Basic " + Base64.getEncoder()
                .encodeToString((CLIENT_ID + ":" + CLIENT_SECRET).getBytes(UTF_8)))
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(BodyPublishers.ofString("grant_type=refresh_token"
                + "&refresh_token=" + URLEncoder.encode(refreshToken, UTF_8)))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient().send(request, BodyHandlers.ofString());
// Store the new refresh_token from the answer before using anything else in it.
```

```go title="Go"
// golang.org/x/oauth2 refreshes on its own: a TokenSource over the token
// you have hands out a new one once it expires.
source := config.TokenSource(ctx, token)
fresh, err := source.Token()
if err != nil {
	return err
}
if fresh.RefreshToken != token.RefreshToken {
	store(fresh) // rotated: keep the new one
}
```

```python title="Python"
response = requests.post(
    discovery["token_endpoint"],
    data={"grant_type": "refresh_token", "refresh_token": refresh_token},
    auth=(CLIENT_ID, CLIENT_SECRET),
)
response.raise_for_status()
tokens = response.json()
store(tokens["refresh_token"])  # rotated: keep the new one first
```

A public client sends `client_id` in the form instead of Basic credentials. `scope` may narrow what the new tokens carry; asking for a scope the original did not grant is refused with `invalid_scope`.

## Rotation

Every refresh **spends** the refresh token it was given and answers with the next one. Store the new one, every time, before you use the new access token.

If a refresh token that was already spent is presented again, one of the two presenters is not your application — so {{name}} revokes the whole family: the token presented and every one issued from it since. The next refresh from your application fails with `invalid_grant`, and the user signs in again. That is on purpose: it makes a stolen refresh token worth one use at most.

So:

- Never refresh the same token twice in parallel. If several requests notice an expired token at once, let one refresh and the others wait for it.
- Write the new refresh token before you do anything else with the answer.

## What a refresh re-checks

Roles, scopes and the user's state are evaluated again at each refresh. A role taken away, an application's access to an API withdrawn, or a user deactivated takes effect at the next refresh — the tokens carry the change, or the refresh is refused with `invalid_grant`.

## Ending it

Revoke the refresh token when the user signs out of your application:

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/revoke" \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -d "token=$REFRESH_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder(URI.create(discovery.revocationEndpoint()))
        .header("Authorization", "Basic " + Base64.getEncoder()
                .encodeToString((CLIENT_ID + ":" + CLIENT_SECRET).getBytes(UTF_8)))
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(BodyPublishers.ofString("token=" + URLEncoder.encode(refreshToken, UTF_8)))
        .build();
HttpClient.newHttpClient().send(request, BodyHandlers.discarding());
```

```go title="Go"
form := url.Values{"token": {refreshToken}}
req, _ := http.NewRequest(http.MethodPost, discovery.RevocationEndpoint, strings.NewReader(form.Encode()))
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
req.SetBasicAuth(clientID, clientSecret)
res, err := http.DefaultClient.Do(req)
```

```python title="Python"
requests.post(
    discovery["revocation_endpoint"],
    data={"token": refresh_token},
    auth=(CLIENT_ID, CLIENT_SECRET),
).raise_for_status()
```

That revokes its whole family. The answer is `200` whatever the token was — even one that did not exist, as RFC 7009 requires. Access tokens cannot be revoked: they are signed JWTs APIs check on their own, and simply expire.
