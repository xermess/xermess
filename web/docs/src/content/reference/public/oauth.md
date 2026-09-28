---
title: "OAuth 2.0 and OpenID Connect"
description: "The endpoints your application signs users in with, and gets and checks tokens from."
order: 2
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

## GET /.well-known/openid-configuration {#get-well-known-openid-configuration}

Answers the OpenID Connect discovery document.

| | |
| --- | --- |
| Auth | None |

```bash title="cURL"
curl "$ISSUER/.well-known/openid-configuration"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/.well-known/openid-configuration"))
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, issuer+"/.well-known/openid-configuration", nil)
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
response = requests.get(f"{ISSUER}/.well-known/openid-configuration")
response.raise_for_status()
```

#### Response

`200 OK`

```json
{
  "issuer": "https://example.com",
  "authorization_endpoint": "string",
  "token_endpoint": "string",
  "userinfo_endpoint": "string",
  "end_session_endpoint": "string",
  "revocation_endpoint": "string",
  "introspection_endpoint": "string",
  "jwks_uri": "https://example.com",
  "response_types_supported": [
    "string"
  ],
  "response_modes_supported": [
    "string"
  ],
  "grant_types_supported": [
    "string"
  ],
  "subject_types_supported": [
    "string"
  ],
  "id_token_signing_alg_values_supported": [
    "string"
  ],
  "scopes_supported": [
    "string"
  ],
  "claims_supported": [
    "string"
  ],
  "token_endpoint_auth_methods_supported": [
    "string"
  ],
  "revocation_endpoint_auth_methods_supported": [
    "string"
  ],
  "introspection_endpoint_auth_methods_supported": [
    "string"
  ],
  "code_challenge_methods_supported": [
    "string"
  ],
  "prompt_values_supported": [
    "string"
  ],
  "authorization_response_iss_parameter_supported": true,
  "op_tos_uri": "https://example.com",
  "op_policy_uri": "https://example.com",
  "service_documentation": "string",
  "claims_parameter_supported": true,
  "request_parameter_supported": true,
  "request_uri_parameter_supported": true
}
```

## GET /.well-known/jwks.json {#get-well-known-jwks-json}

Answers the public keys tokens are signed with.

| | |
| --- | --- |
| Auth | None |

```bash title="cURL"
curl "$ISSUER/.well-known/jwks.json"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/.well-known/jwks.json"))
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, issuer+"/.well-known/jwks.json", nil)
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
response = requests.get(f"{ISSUER}/.well-known/jwks.json")
response.raise_for_status()
```

#### Response

`200 OK`

```json
{
  "keys": [
    {
      "kty": "string",
      "use": "string",
      "alg": "string",
      "kid": "string",
      "n": "string",
      "e": "string",
      "crv": "string",
      "x": "string",
      "y": "string"
    }
  ]
}
```

## GET /oauth2/authorize {#get-oauth2-authorize}

Starts a sign-in for an application, and redirects: to the sign-in page, or straight back with a code for a user already signed in.

| | |
| --- | --- |
| Auth | None |

```bash title="cURL"
curl "$ISSUER/oauth2/authorize?client_id=shop-web&redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback&response_type=code"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/authorize" + "?client_id=shop-web&redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback&response_type=code"))
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, issuer+"/oauth2/authorize"+"?client_id=shop-web&redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback&response_type=code", nil)
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
response = requests.get(
    f"{ISSUER}/oauth2/authorize",
    params={
        "client_id": "shop-web",
        "redirect_uri": "https://app.example.com/callback",
        "response_type": "code",
    },
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `client_id` | query | The application's client ID. |
| `redirect_uri` | query | Where the browser is sent back to. It has to match one of the application's registered redirect URIs exactly; at the token endpoint, it is the one the authorization request named. |
| `response_type` | query | Always `code`. |
| `response_mode` | query | `query`, or left out. |
| `scope` | query | Space-separated scopes: `openid`, `profile`, `email`, `offline_access`, `roles`, and the scopes of the API named in `audience`. |
| `state` | query | An opaque value sent back unchanged with the answer, which ties the answer to the request that asked for it. |
| `nonce` | query | A value copied into the ID token, which ties the token to this sign-in. |
| `code_challenge` | query | The base64url SHA-256 of the code verifier, 43 characters. Required when the application requires PKCE, as every public client does. |
| `code_challenge_method` | query | `S256`, the only method accepted. |
| `audience` | query | The identifier of the API the access token is for, as it is registered on the panel's APIs page. |
| `prompt` | query | `none` to be sent back with `login_required` rather than shown a sign-in page, or `login` to ask for the password even when the user is signed in. |
| `max_age` | query | Seconds since the user last entered their password, past which they are asked again. |
| `login_hint` | query | An email address to fill in on the sign-in page. |

#### Response

`302` — redirects the browser; see the `Location` header.

`400 Bad Request`

## POST /oauth2/authorize {#post-oauth2-authorize}

Starts a sign-in for an application, and redirects: to the sign-in page, or straight back with a code for a user already signed in.

| | |
| --- | --- |
| Auth | None |

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/authorize" \
  -d "client_id=shop-web" \
  -d "redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback" \
  -d "response_type=code" \
  -d "response_mode=value" \
  -d "scope=openid+profile+email" \
  -d "state=af0ifjsldkj" \
  -d "nonce=n-0S6_WzA2Mj" \
  -d "code_challenge=value" \
  -d "code_challenge_method=S256" \
  -d "audience=https%3A%2F%2Fapi.example.com" \
  -d "prompt=login" \
  -d "max_age=value" \
  -d "login_hint=value"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/authorize"))
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(BodyPublishers.ofString(
                "client_id=shop-web"
                + "&redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback"
                + "&response_type=code"
                + "&response_mode=value"
                + "&scope=openid+profile+email"
                + "&state=af0ifjsldkj"
                + "&nonce=n-0S6_WzA2Mj"
                + "&code_challenge=value"
                + "&code_challenge_method=S256"
                + "&audience=https%3A%2F%2Fapi.example.com"
                + "&prompt=login"
                + "&max_age=value"
                + "&login_hint=value"))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
form := url.Values{
	"client_id":             {"shop-web"},
	"redirect_uri":          {"https://app.example.com/callback"},
	"response_type":         {"code"},
	"response_mode":         {"value"},
	"scope":                 {"openid profile email"},
	"state":                 {"af0ifjsldkj"},
	"nonce":                 {"n-0S6_WzA2Mj"},
	"code_challenge":        {"value"},
	"code_challenge_method": {"S256"},
	"audience":              {"https://api.example.com"},
	"prompt":                {"login"},
	"max_age":               {"value"},
	"login_hint":            {"value"},
}
req, err := http.NewRequest(http.MethodPost, issuer+"/oauth2/authorize", strings.NewReader(form.Encode()))
if err != nil {
	return err
}
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.post(
    f"{ISSUER}/oauth2/authorize",
    data={
        "client_id": "shop-web",
        "redirect_uri": "https://app.example.com/callback",
        "response_type": "code",
        "response_mode": "value",
        "scope": "openid profile email",
        "state": "af0ifjsldkj",
        "nonce": "n-0S6_WzA2Mj",
        "code_challenge": "value",
        "code_challenge_method": "S256",
        "audience": "https://api.example.com",
        "prompt": "login",
        "max_age": "value",
        "login_hint": "value",
    },
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `client_id` | form | The application's client ID. |
| `redirect_uri` | form | Where the browser is sent back to. It has to match one of the application's registered redirect URIs exactly; at the token endpoint, it is the one the authorization request named. |
| `response_type` | form | Always `code`. |
| `response_mode` | form | `query`, or left out. |
| `scope` | form | Space-separated scopes: `openid`, `profile`, `email`, `offline_access`, `roles`, and the scopes of the API named in `audience`. |
| `state` | form | An opaque value sent back unchanged with the answer, which ties the answer to the request that asked for it. |
| `nonce` | form | A value copied into the ID token, which ties the token to this sign-in. |
| `code_challenge` | form | The base64url SHA-256 of the code verifier, 43 characters. Required when the application requires PKCE, as every public client does. |
| `code_challenge_method` | form | `S256`, the only method accepted. |
| `audience` | form | The identifier of the API the access token is for, as it is registered on the panel's APIs page. |
| `prompt` | form | `none` to be sent back with `login_required` rather than shown a sign-in page, or `login` to ask for the password even when the user is signed in. |
| `max_age` | form | Seconds since the user last entered their password, past which they are asked again. |
| `login_hint` | form | An email address to fill in on the sign-in page. |

#### Response

`302` — redirects the browser; see the `Location` header.

`400 Bad Request`

## POST /oauth2/token {#post-oauth2-token}

Issues tokens.

| | |
| --- | --- |
| Auth | Client credentials — HTTP Basic; a public client sends `client_id` alone |
| Rate limit | Per IP address, ten times looser than the sign-in pages |

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/token" \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -d "grant_type=authorization_code" \
  -d "code=SplxlOBeZQQYbYS6WxSbIA" \
  -d "redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback" \
  -d "code_verifier=value"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/token"))
        .header("Authorization", "Basic " + Base64.getEncoder()
                .encodeToString((CLIENT_ID + ":" + CLIENT_SECRET).getBytes(UTF_8)))
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(BodyPublishers.ofString(
                "grant_type=authorization_code"
                + "&code=SplxlOBeZQQYbYS6WxSbIA"
                + "&redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback"
                + "&code_verifier=value"))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
form := url.Values{
	"grant_type":    {"authorization_code"},
	"code":          {"SplxlOBeZQQYbYS6WxSbIA"},
	"redirect_uri":  {"https://app.example.com/callback"},
	"code_verifier": {"value"},
}
req, err := http.NewRequest(http.MethodPost, issuer+"/oauth2/token", strings.NewReader(form.Encode()))
if err != nil {
	return err
}
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
req.SetBasicAuth(clientID, clientSecret)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.post(
    f"{ISSUER}/oauth2/token",
    data={
        "grant_type": "authorization_code",
        "code": "SplxlOBeZQQYbYS6WxSbIA",
        "redirect_uri": "https://app.example.com/callback",
        "code_verifier": "value",
    },
    auth=(CLIENT_ID, CLIENT_SECRET),
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `client_secret` | form | The application's secret, when it sends it in the form rather than with HTTP Basic. Never sent by a public client. |
| `client_id` | form | The application's client ID. |
| `grant_type` | form | One of `authorization_code`, `refresh_token`, `client_credentials`. |
| `code` | form | The authorization code the browser was sent back with. It can be spent once, within 2 minutes. |
| `redirect_uri` | form | Where the browser is sent back to. It has to match one of the application's registered redirect URIs exactly; at the token endpoint, it is the one the authorization request named. |
| `code_verifier` | form | The PKCE verifier whose S256 hash was sent as `code_challenge`. |
| `refresh_token` | form | The refresh token to exchange. Refresh tokens rotate: the one sent is spent, and the answer carries the next. |
| `scope` | form | Space-separated scopes: `openid`, `profile`, `email`, `offline_access`, `roles`, and the scopes of the API named in `audience`. |
| `audience` | form | The identifier of the API the access token is for, as it is registered on the panel's APIs page. |

#### Response

`200 OK`

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs…",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "eyJhbGciOiJSUzI1NiIs…",
  "id_token": "eyJhbGciOiJSUzI1NiIs…",
  "scope": "openid profile email"
}
```

Errors come back as an [OAuth error](/reference/errors#oauth-errors): `{"error", "error_description"}`.

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 429 | [`rate_limited`](/reference/errors#rate_limited) | Too many attempts. Try again in {seconds} seconds. |

## GET /oauth2/userinfo {#get-oauth2-userinfo}

Answers the claims about the user an access token is for.

| | |
| --- | --- |
| Auth | Bearer access token |

```bash title="cURL"
curl "$ISSUER/oauth2/userinfo" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/userinfo"))
        .header("Authorization", "Bearer " + ACCESS_TOKEN)
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, issuer+"/oauth2/userinfo", nil)
if err != nil {
	return err
}
req.Header.Set("Authorization", "Bearer "+accessToken)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.get(
    f"{ISSUER}/oauth2/userinfo",
    headers={"Authorization": f"Bearer {ACCESS_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `Authorization` | header | Reads the access token from the Authorization header, or from the form of a POST (RFC 6750 section 2). |

#### Response

`200 OK`

Errors come back as an [OAuth error](/reference/errors#oauth-errors): `{"error", "error_description"}`.

## POST /oauth2/userinfo {#post-oauth2-userinfo}

Answers the claims about the user an access token is for.

| | |
| --- | --- |
| Auth | Bearer access token |

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/userinfo" \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/userinfo"))
        .header("Authorization", "Bearer " + ACCESS_TOKEN)
        .POST(BodyPublishers.noBody())
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodPost, issuer+"/oauth2/userinfo", nil)
if err != nil {
	return err
}
req.Header.Set("Authorization", "Bearer "+accessToken)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.post(
    f"{ISSUER}/oauth2/userinfo",
    headers={"Authorization": f"Bearer {ACCESS_TOKEN}"},
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `access_token` | form | The access token, for a client that cannot send an Authorization header. |
| `Authorization` | header | Reads the access token from the Authorization header, or from the form of a POST (RFC 6750 section 2). |

#### Response

`200 OK`

Errors come back as an [OAuth error](/reference/errors#oauth-errors): `{"error", "error_description"}`.

## GET /oauth2/logout {#get-oauth2-logout}

Signs the user out of this server, and redirects.

| | |
| --- | --- |
| Auth | None |

```bash title="cURL"
curl "$ISSUER/oauth2/logout?id_token_hint=value&client_id=shop-web&post_logout_redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/logout" + "?id_token_hint=value&client_id=shop-web&post_logout_redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback"))
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, issuer+"/oauth2/logout"+"?id_token_hint=value&client_id=shop-web&post_logout_redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback", nil)
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
response = requests.get(
    f"{ISSUER}/oauth2/logout",
    params={
        "id_token_hint": "value",
        "client_id": "shop-web",
        "post_logout_redirect_uri": "https://app.example.com/callback",
    },
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `id_token_hint` | query | An ID token this server issued, which says whose session to end. It may have expired. |
| `client_id` | query | The application's client ID. |
| `post_logout_redirect_uri` | query | Where to send the browser once signed out. It has to be one the application registered; without one, the browser lands on the sign-in app's signed-out page. |
| `state` | query | An opaque value sent back unchanged with the answer, which ties the answer to the request that asked for it. |

#### Response

`302` — redirects the browser; see the `Location` header.

`400 Bad Request`

## POST /oauth2/logout {#post-oauth2-logout}

Signs the user out of this server, and redirects.

| | |
| --- | --- |
| Auth | None |

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/logout" \
  -d "id_token_hint=value" \
  -d "client_id=shop-web" \
  -d "post_logout_redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback" \
  -d "state=af0ifjsldkj"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/logout"))
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(BodyPublishers.ofString(
                "id_token_hint=value"
                + "&client_id=shop-web"
                + "&post_logout_redirect_uri=https%3A%2F%2Fapp.example.com%2Fcallback"
                + "&state=af0ifjsldkj"))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
form := url.Values{
	"id_token_hint":            {"value"},
	"client_id":                {"shop-web"},
	"post_logout_redirect_uri": {"https://app.example.com/callback"},
	"state":                    {"af0ifjsldkj"},
}
req, err := http.NewRequest(http.MethodPost, issuer+"/oauth2/logout", strings.NewReader(form.Encode()))
if err != nil {
	return err
}
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.post(
    f"{ISSUER}/oauth2/logout",
    data={
        "id_token_hint": "value",
        "client_id": "shop-web",
        "post_logout_redirect_uri": "https://app.example.com/callback",
        "state": "af0ifjsldkj",
    },
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `id_token_hint` | form | An ID token this server issued, which says whose session to end. It may have expired. |
| `client_id` | form | The application's client ID. |
| `post_logout_redirect_uri` | form | Where to send the browser once signed out. It has to be one the application registered; without one, the browser lands on the sign-in app's signed-out page. |
| `state` | form | An opaque value sent back unchanged with the answer, which ties the answer to the request that asked for it. |

#### Response

`302` — redirects the browser; see the `Location` header.

`400 Bad Request`

## POST /oauth2/revoke {#post-oauth2-revoke}

Revokes a refresh token.

| | |
| --- | --- |
| Auth | Client credentials — HTTP Basic; a public client sends `client_id` alone |
| Rate limit | Per IP address, ten times looser than the sign-in pages |

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/revoke" \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -d "token=eyJhbGciOiJSUzI1NiIs%E2%80%A6"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/revoke"))
        .header("Authorization", "Basic " + Base64.getEncoder()
                .encodeToString((CLIENT_ID + ":" + CLIENT_SECRET).getBytes(UTF_8)))
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(BodyPublishers.ofString(
                "token=eyJhbGciOiJSUzI1NiIs%E2%80%A6"))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
form := url.Values{
	"token": {"eyJhbGciOiJSUzI1NiIs…"},
}
req, err := http.NewRequest(http.MethodPost, issuer+"/oauth2/revoke", strings.NewReader(form.Encode()))
if err != nil {
	return err
}
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
req.SetBasicAuth(clientID, clientSecret)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.post(
    f"{ISSUER}/oauth2/revoke",
    data={
        "token": "eyJhbGciOiJSUzI1NiIs…",
    },
    auth=(CLIENT_ID, CLIENT_SECRET),
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `client_secret` | form | The application's secret, when it sends it in the form rather than with HTTP Basic. Never sent by a public client. |
| `client_id` | form | The application's client ID. |
| `token` | form | The token the request is about. |

#### Response

`200 OK`

Errors come back as an [OAuth error](/reference/errors#oauth-errors): `{"error", "error_description"}`.

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 429 | [`rate_limited`](/reference/errors#rate_limited) | Too many attempts. Try again in {seconds} seconds. |

## POST /oauth2/introspect {#post-oauth2-introspect}

Says whether a token is active.

| | |
| --- | --- |
| Auth | Client credentials — HTTP Basic; a public client sends `client_id` alone |
| Rate limit | Per IP address, ten times looser than the sign-in pages |

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/introspect" \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -d "token=eyJhbGciOiJSUzI1NiIs%E2%80%A6"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/introspect"))
        .header("Authorization", "Basic " + Base64.getEncoder()
                .encodeToString((CLIENT_ID + ":" + CLIENT_SECRET).getBytes(UTF_8)))
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(BodyPublishers.ofString(
                "token=eyJhbGciOiJSUzI1NiIs%E2%80%A6"))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
form := url.Values{
	"token": {"eyJhbGciOiJSUzI1NiIs…"},
}
req, err := http.NewRequest(http.MethodPost, issuer+"/oauth2/introspect", strings.NewReader(form.Encode()))
if err != nil {
	return err
}
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
req.SetBasicAuth(clientID, clientSecret)

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.post(
    f"{ISSUER}/oauth2/introspect",
    data={
        "token": "eyJhbGciOiJSUzI1NiIs…",
    },
    auth=(CLIENT_ID, CLIENT_SECRET),
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `client_secret` | form | The application's secret, when it sends it in the form rather than with HTTP Basic. Never sent by a public client. |
| `client_id` | form | The application's client ID. |
| `token` | form | The token the request is about. |

#### Response

`200 OK`

Errors come back as an [OAuth error](/reference/errors#oauth-errors): `{"error", "error_description"}`.

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 429 | [`rate_limited`](/reference/errors#rate_limited) | Too many attempts. Try again in {seconds} seconds. |

## GET /oauth2/social/:slug/start {#get-oauth2-social-slug-start}

Sends the browser to a provider to sign in there, leaving the sign-in's state with the browser so the callback can tell this browser's answer from one somebody else's sign-in produced.

| | |
| --- | --- |
| Auth | None |

```bash title="cURL"
curl "$ISSUER/oauth2/social/$SLUG/start?request=value&next=value"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/social/" + slug + "/start" + "?request=value&next=value"))
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, issuer+"/oauth2/social/"+slug+"/start"+"?request=value&next=value", nil)
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
response = requests.get(
    f"{ISSUER}/oauth2/social/{slug}/start",
    params={
        "request": "value",
        "next": "value",
    },
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `slug` <small>required</small> | path |  |
| `request` | query |  |
| `next` | query |  |

#### Response

`302` — redirects the browser; see the `Location` header.

## GET /oauth2/social/:slug/callback {#get-oauth2-social-slug-callback}

Is where the provider sends the browser back to.

It answers GET and POST: nearly every provider redirects with the code in the query, and Apple posts it as a form when a name or an address was asked for.

| | |
| --- | --- |
| Auth | None |

```bash title="cURL"
curl "$ISSUER/oauth2/social/$SLUG/callback?code=SplxlOBeZQQYbYS6WxSbIA&state=af0ifjsldkj&error=value"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/social/" + slug + "/callback" + "?code=SplxlOBeZQQYbYS6WxSbIA&state=af0ifjsldkj&error=value"))
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, issuer+"/oauth2/social/"+slug+"/callback"+"?code=SplxlOBeZQQYbYS6WxSbIA&state=af0ifjsldkj&error=value", nil)
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
response = requests.get(
    f"{ISSUER}/oauth2/social/{slug}/callback",
    params={
        "code": "SplxlOBeZQQYbYS6WxSbIA",
        "state": "af0ifjsldkj",
        "error": "value",
    },
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `slug` <small>required</small> | path |  |
| `code` | query | The authorization code the browser was sent back with. It can be spent once, within 2 minutes. |
| `state` | query | An opaque value sent back unchanged with the answer, which ties the answer to the request that asked for it. |
| `error` | query |  |
| `error_description` | query |  |

#### Response

`302` — redirects the browser; see the `Location` header.

## POST /oauth2/social/:slug/callback {#post-oauth2-social-slug-callback}

Is where the provider sends the browser back to.

It answers GET and POST: nearly every provider redirects with the code in the query, and Apple posts it as a form when a name or an address was asked for.

| | |
| --- | --- |
| Auth | None |

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/social/$SLUG/callback" \
  -d "code=SplxlOBeZQQYbYS6WxSbIA" \
  -d "state=af0ifjsldkj"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/social/" + slug + "/callback"))
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(BodyPublishers.ofString(
                "code=SplxlOBeZQQYbYS6WxSbIA"
                + "&state=af0ifjsldkj"))
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
form := url.Values{
	"code":  {"SplxlOBeZQQYbYS6WxSbIA"},
	"state": {"af0ifjsldkj"},
}
req, err := http.NewRequest(http.MethodPost, issuer+"/oauth2/social/"+slug+"/callback", strings.NewReader(form.Encode()))
if err != nil {
	return err
}
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

res, err := http.DefaultClient.Do(req)
if err != nil {
	return err
}
defer res.Body.Close()
```

```python title="Python"
response = requests.post(
    f"{ISSUER}/oauth2/social/{slug}/callback",
    data={
        "code": "SplxlOBeZQQYbYS6WxSbIA",
        "state": "af0ifjsldkj",
    },
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `slug` <small>required</small> | path |  |
| `code` | query | The authorization code the browser was sent back with. It can be spent once, within 2 minutes. |
| `state` | query | An opaque value sent back unchanged with the answer, which ties the answer to the request that asked for it. |
| `error` | query |  |
| `error_description` | query |  |
| `code` | form | The authorization code the browser was sent back with. It can be spent once, within 2 minutes. |
| `state` | form | An opaque value sent back unchanged with the answer, which ties the answer to the request that asked for it. |

#### Response

`302` — redirects the browser; see the `Location` header.

## GET /oauth2/sso/:slug/start {#get-oauth2-sso-slug-start}

Sends the browser to a connection's identity provider.

The address typed on the sign-in page comes along as `login_hint`, so it is not asked for twice.

| | |
| --- | --- |
| Auth | None |
| Rate limit | Per IP address |

```bash title="cURL"
curl "$ISSUER/oauth2/sso/$SLUG/start?request=value&next=value&login_hint=value"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/sso/" + slug + "/start" + "?request=value&next=value&login_hint=value"))
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, issuer+"/oauth2/sso/"+slug+"/start"+"?request=value&next=value&login_hint=value", nil)
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
response = requests.get(
    f"{ISSUER}/oauth2/sso/{slug}/start",
    params={
        "request": "value",
        "next": "value",
        "login_hint": "value",
    },
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `slug` <small>required</small> | path |  |
| `request` | query |  |
| `next` | query |  |
| `login_hint` | query | An email address to fill in on the sign-in page. |

#### Response

`302` — redirects the browser; see the `Location` header.

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 429 | [`rate_limited`](/reference/errors#rate_limited) | Too many attempts. Try again in {seconds} seconds. |

## GET /oauth2/sso/:slug/callback {#get-oauth2-sso-slug-callback}

Is where an OpenID Connect provider sends the browser back to.

| | |
| --- | --- |
| Auth | None |

```bash title="cURL"
curl "$ISSUER/oauth2/sso/$SLUG/callback?error=value&state=af0ifjsldkj&error_description=value"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/sso/" + slug + "/callback" + "?error=value&state=af0ifjsldkj&error_description=value"))
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, issuer+"/oauth2/sso/"+slug+"/callback"+"?error=value&state=af0ifjsldkj&error_description=value", nil)
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
response = requests.get(
    f"{ISSUER}/oauth2/sso/{slug}/callback",
    params={
        "error": "value",
        "state": "af0ifjsldkj",
        "error_description": "value",
    },
)
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `slug` <small>required</small> | path |  |
| `error` | query |  |
| `state` | query | An opaque value sent back unchanged with the answer, which ties the answer to the request that asked for it. |
| `error_description` | query |  |
| `code` | query | The authorization code the browser was sent back with. It can be spent once, within 2 minutes. |

#### Response

`302` — redirects the browser; see the `Location` header.

## POST /oauth2/sso/:slug/acs {#post-oauth2-sso-slug-acs}

Is the assertion consumer service a SAML provider posts its response to.

| | |
| --- | --- |
| Auth | None |
| Rate limit | Per IP address |

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/sso/$SLUG/acs"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/sso/" + slug + "/acs"))
        .POST(BodyPublishers.noBody())
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodPost, issuer+"/oauth2/sso/"+slug+"/acs", nil)
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
response = requests.post(f"{ISSUER}/oauth2/sso/{slug}/acs")
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `slug` <small>required</small> | path |  |

#### Response

`302` — redirects the browser; see the `Location` header.

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 429 | [`rate_limited`](/reference/errors#rate_limited) | Too many attempts. Try again in {seconds} seconds. |

## GET /oauth2/sso/:slug/metadata {#get-oauth2-sso-slug-metadata}

Is this server's SAML metadata as one connection's service provider: the file an identity provider is set up from.

| | |
| --- | --- |
| Auth | None |

```bash title="cURL"
curl "$ISSUER/oauth2/sso/$SLUG/metadata"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder()
        .uri(URI.create(ISSUER + "/oauth2/sso/" + slug + "/metadata"))
        .GET()
        .build();

HttpResponse<String> response = HttpClient.newHttpClient()
        .send(request, BodyHandlers.ofString());
```

```go title="Go"
req, err := http.NewRequest(http.MethodGet, issuer+"/oauth2/sso/"+slug+"/metadata", nil)
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
response = requests.get(f"{ISSUER}/oauth2/sso/{slug}/metadata")
response.raise_for_status()
```

#### Parameters

| Name | In | Description |
| --- | --- | --- |
| `slug` <small>required</small> | path |  |

#### Response

`200 OK`

#### Errors

| Status | Code | Means |
| --- | --- | --- |
| 404 | [`not_found`](/reference/errors#not_found) | There is nothing here. |
