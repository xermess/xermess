---
title: Sign users in
description: The authorization code flow with PKCE — the one flow every application with a user uses.
order: 1
section: Sign in
nav: 'Sign users in'
icon: login-box
---

> [!TIP]
> In production, use a library: Spring Security's `oauth2-login` (Java), `golang.org/x/oauth2` with `go-oidc` (Go), or Authlib (Python). Give it `$ISSUER/.well-known/openid-configuration`. This page shows what it does for you.

## 1. Redirect to sign in

Keep a random `state`, `nonce` and PKCE `verifier` in the user's session, then redirect:

```http
GET $ISSUER/oauth2/authorize
  ?response_type=code
  &client_id=CLIENT_ID
  &redirect_uri=https://app.example.com/callback
  &scope=openid profile email
  &state=STATE
  &nonce=NONCE
  &code_challenge=CHALLENGE
  &code_challenge_method=S256
```

| Parameter               | Value                                                                           |
| ----------------------- | ------------------------------------------------------------------------------- |
| `response_type`         | `code`                                                                          |
| `client_id`             | Your client ID                                                                  |
| `redirect_uri`          | A registered redirect URI — matched exactly                                     |
| `scope`                 | `openid` plus what you need: `profile`, `email`, `offline_access`, `roles`      |
| `state`, `nonce`        | Random values you check later                                                   |
| `code_challenge`        | base64url(SHA-256(verifier))                                                    |
| `code_challenge_method` | `S256`                                                                          |
| `audience`              | _Optional._ The API the access token is for                                     |
| `prompt`                | _Optional._ `login` to always ask for the password; `none` never to show a page |

Building it in code:

```java title="Java"
String location = discovery.authorizationEndpoint() + "?" + String.join("&",
        "response_type=code",
        "client_id=" + URLEncoder.encode(CLIENT_ID, UTF_8),
        "redirect_uri=" + URLEncoder.encode("https://app.example.com/callback", UTF_8),
        "scope=" + URLEncoder.encode("openid profile email", UTF_8),
        "state=" + state,
        "nonce=" + nonce,
        "code_challenge=" + challenge,
        "code_challenge_method=S256");
```

```go title="Go"
config := &oauth2.Config{
	ClientID:     clientID,
	ClientSecret: clientSecret,
	RedirectURL:  "https://app.example.com/callback",
	Scopes:       []string{"openid", "profile", "email"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  discovery.AuthorizationEndpoint,
		TokenURL: discovery.TokenEndpoint,
	},
}
url := config.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier),
	oauth2.SetAuthURLParam("nonce", nonce))
```

```python title="Python"
location = discovery["authorization_endpoint"] + "?" + urlencode({
    "response_type": "code",
    "client_id": CLIENT_ID,
    "redirect_uri": "https://app.example.com/callback",
    "scope": "openid profile email",
    "state": state,
    "nonce": nonce,
    "code_challenge": challenge,
    "code_challenge_method": "S256",
})
```

## 2. Handle the callback

```http
https://app.example.com/callback?code=CODE&state=STATE&iss=https://id.example.com
```

Check that `state` is yours and `iss` is your issuer. If sign-in didn't happen, you get `error` instead:

| `error`               | Why                                                         |
| --------------------- | ----------------------------------------------------------- |
| `access_denied`       | The user may not use this app — inactive, or missing a role |
| `login_required`      | `prompt=none`, and the user isn't signed in                 |
| `invalid_request`     | A parameter is wrong — see `error_description`              |
| `unauthorized_client` | The app is disabled or can't use this flow                  |

## 3. Exchange the code

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/token" \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -d "grant_type=authorization_code" \
  -d "code=$CODE" \
  -d "redirect_uri=https://app.example.com/callback" \
  -d "code_verifier=$VERIFIER"
```

```java title="Java"
String form = "grant_type=authorization_code"
        + "&code=" + URLEncoder.encode(code, UTF_8)
        + "&redirect_uri=" + URLEncoder.encode("https://app.example.com/callback", UTF_8)
        + "&code_verifier=" + verifier;

HttpRequest request = HttpRequest.newBuilder(URI.create(discovery.tokenEndpoint()))
        .header("Authorization", "Basic " + Base64.getEncoder()
                .encodeToString((CLIENT_ID + ":" + CLIENT_SECRET).getBytes(UTF_8)))
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(BodyPublishers.ofString(form))
        .build();
HttpResponse<String> response = HttpClient.newHttpClient().send(request, BodyHandlers.ofString());
```

```go title="Go"
token, err := config.Exchange(ctx, code, oauth2.VerifierOption(verifier))
if err != nil {
	return err
}
rawIDToken, _ := token.Extra("id_token").(string)
```

```python title="Python"
tokens = requests.post(
    discovery["token_endpoint"],
    data={
        "grant_type": "authorization_code",
        "code": code,
        "redirect_uri": "https://app.example.com/callback",
        "code_verifier": verifier,
    },
    auth=(CLIENT_ID, CLIENT_SECRET),
).json()
```

The body is a form, not JSON. Apps without a secret send `client_id` in the form instead of `-u`.

## 4. Check the ID token

Check its signature and claims — [Tokens](/guides/tokens) shows how — then use `sub` as the user's ID. It never changes; their email can.

## Good to know

- **Already signed in?** The user comes straight back with a code, no page shown. That's single sign-on across your apps.
- **Mobile and single-page apps** use the same flow with no secret. Open the sign-in page in the system browser, not a web view.
- **Never** keep tokens in `localStorage` in the browser.
