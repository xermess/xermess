---
title: Protect an API
description: Check the access tokens your API is called with, against the keys the server publishes.
order: 1
section: Protect APIs
nav: 'Check access tokens'
icon: shield-check
---

An API receives `Authorization: Bearer <access token>` and has to decide whether to believe it. It does that on its own: the token is a signed JWT, and everything needed to check it is published.

## Check the JWT yourself

Access tokens are JWTs (RFC 9068), signed with keys the server publishes. Checking one needs no call to {{name}} beyond fetching those keys now and then:

1. The header's `typ` is `at+jwt` — so an ID token cannot be passed off as an access token.
2. The signature checks out with the key at `jwks_uri` that `kid` names, with the algorithm your API is set to sign with (RS256 by default).
3. `iss` is the issuer; `aud` contains **your API's identifier**; `exp` has not passed.
4. `scope` contains what the endpoint needs.

```java title="Java"
// com.nimbusds:nimbus-jose-jwt
var keys = JWKSourceBuilder.create(new URL(ISSUER + "/.well-known/jwks.json")).build();
var processor = new DefaultJWTProcessor<SecurityContext>();
processor.setJWSTypeVerifier(new DefaultJOSEObjectTypeVerifier<>(new JOSEObjectType("at+jwt")));
processor.setJWSKeySelector(new JWSVerificationKeySelector<>(JWSAlgorithm.RS256, keys));
processor.setJWTClaimsSetVerifier(new DefaultJWTClaimsVerifier<>(
        "https://api.example.com",
        new JWTClaimsSet.Builder().issuer(ISSUER).build(),
        Set.of("sub", "exp", "scope")));

// Build the processor once; it caches the keys and refetches on a new kid.
JWTClaimsSet claims = processor.process(bearerToken, null);
List<String> scopes = List.of(claims.getStringClaim("scope").split(" "));
```

```go title="Go"
// github.com/lestrrat-go/jwx/v3
cache, _ := jwk.NewCache(ctx, httprc.NewClient())
cache.Register(ctx, issuer+"/.well-known/jwks.json")
keys, _ := cache.CachedSet(issuer + "/.well-known/jwks.json")

token, err := jwt.Parse([]byte(raw),
	jwt.WithKeySet(keys),
	jwt.WithIssuer(issuer),
	jwt.WithAudience("https://api.example.com"),
	jwt.WithValidate(true))
```

```python title="Python"
import jwt  # PyJWT

keys = jwt.PyJWKClient(f"{ISSUER}/.well-known/jwks.json")

def authenticate(header: str) -> dict:
    scheme, _, token = header.partition(" ")
    if scheme != "Bearer" or not token:
        raise Unauthorized()
    if jwt.get_unverified_header(token).get("typ") != "at+jwt":
        raise Unauthorized()
    key = keys.get_signing_key_from_jwt(token)
    return jwt.decode(token, key.key, algorithms=["RS256"],
                      audience="https://api.example.com", issuer=ISSUER)
```

### Keys rotate

Keys rotate on their own. A new key is published a day before it signs anything, and an old one stays published for two days after it stops — so a cache that refetches `jwks_uri` every few minutes, and whenever it meets a `kid` it does not know, never rejects a good token. The answer may be cached for five minutes.

When a key may have leaked, an administrator can rotate at once and revoke the old keys; tokens they signed stop checking out the next time your cache refetches.

## Introspection is for an application's own tokens

The introspection endpoint (RFC 7662) answers whether a token is active and what it carries — but only to the confidential application the token was **issued to**. Asked about another application's token, it answers `{"active": false}`, whether or not the token is good: what a token carries about a user is not something any client holding a secret may read.

So introspection is not how a separate API checks the tokens its callers bring: that is the JWT check above. It is for an application asking about tokens it holds itself — whether a refresh token is still alive, say:

```bash title="cURL"
curl -X POST "$ISSUER/oauth2/introspect" \
  -u "$CLIENT_ID:$CLIENT_SECRET" \
  -d "token=$REFRESH_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder(URI.create(ISSUER + "/oauth2/introspect"))
        .header("Authorization", "Basic " + Base64.getEncoder()
                .encodeToString((CLIENT_ID + ":" + CLIENT_SECRET).getBytes(UTF_8)))
        .header("Content-Type", "application/x-www-form-urlencoded")
        .POST(BodyPublishers.ofString("token=" + URLEncoder.encode(refreshToken, UTF_8)))
        .build();
String answer = HttpClient.newHttpClient().send(request, BodyHandlers.ofString()).body();
```

```go title="Go"
form := url.Values{"token": {refreshToken}}
req, _ := http.NewRequest(http.MethodPost, issuer+"/oauth2/introspect", strings.NewReader(form.Encode()))
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
req.SetBasicAuth(clientID, clientSecret)
res, err := http.DefaultClient.Do(req)
```

```python title="Python"
answer = requests.post(
    f"{ISSUER}/oauth2/introspect",
    data={"token": refresh_token},
    auth=(CLIENT_ID, CLIENT_SECRET),
).json()
```

```json
{
	"active": true,
	"token_type": "refresh_token",
	"client_id": "shop-web",
	"sub": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
	"scope": "openid offline_access",
	"iat": 1790000000,
	"exp": 1792592000,
	"iss": "https://id.example.com"
}
```

An access token's answer is `active`, `token_type: "Bearer"` and every claim the token carries. Public clients cannot introspect at all, and each call counts against the token rate limit.

## Roles

With the `roles` scope, on an application that asserts roles, the access token carries `roles` — the user's roles in the application the token was issued to — and `global_roles`. Check scopes for what a client may do, and roles for what the user may do.
