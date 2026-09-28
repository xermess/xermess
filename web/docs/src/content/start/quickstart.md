---
title: Quickstart
description: Register an application, sign a user in, and get their tokens — in five steps.
order: 1
icon: flashlight
---

## 1. Register your application

In the panel, open **Applications → New**. Pick its type:

| Type               | Use for                         | Secret |
| ------------------ | ------------------------------- | ------ |
| Web                | An app with a server            | Yes    |
| Single-page        | An app that runs in the browser | No     |
| Native             | A mobile or desktop app         | No     |
| Machine-to-machine | A service with no user          | Yes    |

Add `http://localhost:3000/callback` as a **redirect URI**, then copy the **client ID** and **client secret** — the secret is shown once.

## 2. Make a PKCE pair

A random **verifier** you keep, and its SHA-256 **challenge** you send.

```bash title="cURL"
VERIFIER=$(openssl rand -base64 48 | tr -d '=+/' | cut -c1-64)
CHALLENGE=$(printf '%s' "$VERIFIER" | openssl dgst -sha256 -binary | openssl base64 | tr '+/' '-_' | tr -d '=')
```

```java title="Java"
byte[] random = new byte[48];
new SecureRandom().nextBytes(random);
String verifier = Base64.getUrlEncoder().withoutPadding().encodeToString(random);
String challenge = Base64.getUrlEncoder().withoutPadding().encodeToString(
        MessageDigest.getInstance("SHA-256").digest(verifier.getBytes(US_ASCII)));
```

```go title="Go"
verifier := oauth2.GenerateVerifier() // golang.org/x/oauth2
challenge := oauth2.S256ChallengeFromVerifier(verifier)
```

```python title="Python"
verifier = secrets.token_urlsafe(48)
challenge = base64.urlsafe_b64encode(
    hashlib.sha256(verifier.encode()).digest()
).rstrip(b"=").decode()
```

## 3. Send the user to sign in

Open this in a browser:

```http
$ISSUER/oauth2/authorize?response_type=code&client_id=CLIENT_ID&redirect_uri=http://localhost:3000/callback&scope=openid profile email&state=xyz&code_challenge=CHALLENGE&code_challenge_method=S256
```

After signing in, the browser lands on your redirect URI with a `code`:

```http
http://localhost:3000/callback?code=SplxlOBeZQQYbYS6WxSbIA&state=xyz&iss=http://localhost:5173
```

## 4. Exchange the code for tokens

Within two minutes, and only once:

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/token" \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -d "grant_type=authorization_code" \
  -d "code=$CODE" \
  -d "redirect_uri=http://localhost:3000/callback" \
  -d "code_verifier=$VERIFIER"
```

```java title="Java"
String form = "grant_type=authorization_code"
        + "&code=" + URLEncoder.encode(code, UTF_8)
        + "&redirect_uri=" + URLEncoder.encode("http://localhost:3000/callback", UTF_8)
        + "&code_verifier=" + verifier;

HttpRequest request = HttpRequest.newBuilder(URI.create(ISSUER + "/oauth2/token"))
        .header("Authorization", "Basic " + Base64.getEncoder()
                .encodeToString((CLIENT_ID + ":" + CLIENT_SECRET).getBytes(UTF_8)))
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(BodyPublishers.ofString(form))
        .build();
String tokens = HttpClient.newHttpClient().send(request, BodyHandlers.ofString()).body();
```

```go title="Go"
form := url.Values{
	"grant_type":    {"authorization_code"},
	"code":          {code},
	"redirect_uri":  {"http://localhost:3000/callback"},
	"code_verifier": {verifier},
}
req, _ := http.NewRequest(http.MethodPost, issuer+"/oauth2/token", strings.NewReader(form.Encode()))
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
req.SetBasicAuth(clientID, clientSecret)
res, err := http.DefaultClient.Do(req)
```

```python title="Python"
tokens = requests.post(
    f"{ISSUER}/oauth2/token",
    data={
        "grant_type": "authorization_code",
        "code": code,
        "redirect_uri": "http://localhost:3000/callback",
        "code_verifier": verifier,
    },
    auth=(CLIENT_ID, CLIENT_SECRET),
).json()
```

```json
{
	"access_token": "eyJhbGciOiJSUzI1NiIs…",
	"token_type": "Bearer",
	"expires_in": 3600,
	"id_token": "eyJhbGciOiJSUzI1NiIs…",
	"scope": "openid profile email"
}
```

A single-page or native app sends `-d "client_id=$CLIENT_ID"` instead of a secret.

## 5. Get the user

```bash title="cURL"
curl "$ISSUER/oauth2/userinfo" -H "Authorization: Bearer $ACCESS_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder(URI.create(ISSUER + "/oauth2/userinfo"))
        .header("Authorization", "Bearer " + accessToken)
        .build();
String user = HttpClient.newHttpClient().send(request, BodyHandlers.ofString()).body();
```

```go title="Go"
req, _ := http.NewRequest(http.MethodGet, issuer+"/oauth2/userinfo", nil)
req.Header.Set("Authorization", "Bearer "+accessToken)
res, err := http.DefaultClient.Do(req)
```

```python title="Python"
user = requests.get(
    f"{ISSUER}/oauth2/userinfo",
    headers={"Authorization": f"Bearer {access_token}"},
).json()
```

```json
{
	"sub": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
	"name": "Ada Lovelace",
	"email": "ada@example.com",
	"email_verified": true
}
```

That's a full sign-in. Next: [do it with a library](/guides/authorization-code), [keep users signed in](/guides/refresh-tokens), or [protect your API](/guides/protecting-an-api).
