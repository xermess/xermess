---
title: Tokens
description: What the ID token, the access token and userinfo carry, and how to check them.
order: 2
section: Sign in
nav: 'Tokens'
icon: key-2
---

A token response can hold three tokens, each for a different reader:

| Token           | For                                        | Format                                                                      |
| --------------- | ------------------------------------------ | --------------------------------------------------------------------------- |
| `id_token`      | Your application: who signed in            | JWT, `typ: JWT`, always RS256                                               |
| `access_token`  | The API named in `audience`, or `userinfo` | JWT, `typ: at+jwt` (RFC 9068), RS256 unless the API chose another algorithm |
| `refresh_token` | The token endpoint, and nobody else        | Opaque                                                                      |

Never send the ID token to an API, and never read the access token in your application to find out who the user is — that is what the ID token and `userinfo` are for. An access token's contents are an agreement between {{name}} and the API.

## The ID token

```json
{
	"iss": "https://id.example.com",
	"sub": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
	"aud": "shop-web",
	"azp": "shop-web",
	"iat": 1790000000,
	"exp": 1790003600,
	"auth_time": 1789999950,
	"nonce": "n-0S6_WzA2Mj",
	"sid": "0199a3c2-8a41-7b55-a1e3-5e6f7a8b9c0d",
	"at_hash": "77QmUPtjPfzWtF2AnpK9RQ",
	"name": "Ada Lovelace",
	"given_name": "Ada",
	"family_name": "Lovelace",
	"email": "ada@example.com",
	"email_verified": true,
	"roles": ["editor"],
	"global_roles": ["staff"]
}
```

| Claim                               |                                                                                                                         |
| ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `sub`                               | The user's ID. Key your accounts on this, not on the email, which can change.                                           |
| `aud`, `azp`                        | Your client ID                                                                                                          |
| `auth_time`                         | When the user last entered their password                                                                               |
| `nonce`                             | What you sent; check it                                                                                                 |
| `sid`                               | Their session at {{name}}                                                                                               |
| `name`, `given_name`, `family_name` | With the `profile` scope                                                                                                |
| `email`, `email_verified`           | With the `email` scope                                                                                                  |
| `roles`, `global_roles`             | With the `roles` scope, when the application asserts roles: the user's roles in this application, and their global ones |

### Checking one

1. Fetch the keys from `jwks_uri`, and cache them — the answer may be cached for five minutes. When a token names a `kid` you don't have, fetch again: keys rotate, and a new one is published a day before it signs anything.
2. Check the signature with the key its `kid` names, and that `alg` is `RS256`.
3. Check `iss` is your issuer, `aud` is your client ID, and `exp` is in the future. Allow a minute of clock skew, not more.
4. Check `nonce` is the one you kept for this sign-in.

Every JOSE library does steps 1–3 in one call:

```java title="Java"
// com.nimbusds:nimbus-jose-jwt
var keys = JWKSourceBuilder.create(new URL(discovery.jwksUri())).build();
var processor = new DefaultJWTProcessor<SecurityContext>();
processor.setJWSKeySelector(new JWSVerificationKeySelector<>(JWSAlgorithm.RS256, keys));
processor.setJWTClaimsSetVerifier(new DefaultJWTClaimsVerifier<>(
        CLIENT_ID,
        new JWTClaimsSet.Builder().issuer(discovery.issuer()).build(),
        Set.of("sub", "exp", "iat")));

JWTClaimsSet claims = processor.process(idToken, null);
if (!session.nonce().equals(claims.getStringClaim("nonce"))) {
    throw new BadJOSEException("the ID token is not for this sign-in");
}
```

```go title="Go"
// github.com/coreos/go-oidc/v3
provider, err := oidc.NewProvider(ctx, issuer)
verifier := provider.Verifier(&oidc.Config{ClientID: clientID})
idToken, err := verifier.Verify(ctx, rawIDToken)
if err != nil || idToken.Nonce != nonce {
	return errors.New("the ID token is not for this sign-in")
}
```

```python title="Python"
import jwt  # PyJWT

keys = jwt.PyJWKClient(discovery["jwks_uri"])
key = keys.get_signing_key_from_jwt(id_token)
claims = jwt.decode(id_token, key.key, algorithms=["RS256"],
                    audience=CLIENT_ID, issuer=discovery["issuer"])
if claims["nonce"] != session["nonce"]:
    raise ValueError("the ID token is not for this sign-in")
```

## The access token

```json
{
	"iss": "https://id.example.com",
	"sub": "0199a3c2-7f10-7c1e-9b0e-3f1d2a4b5c6d",
	"aud": ["https://api.example.com"],
	"azp": "shop-web",
	"client_id": "shop-web",
	"scope": "openid orders:read",
	"iat": 1790000000,
	"exp": 1790003600,
	"jti": "0199a3c2-9b12-7d3e-8f4a-1b2c3d4e5f60",
	"auth_time": 1789999950
}
```

`aud` is the API named in `audience`, and empty without one. `sub` is the user — or, for [client credentials](/guides/machine-to-machine), the application's client ID. `roles` and `global_roles` are here too under the same conditions as in the ID token. How an API checks one is in [Protect an API](/guides/protecting-an-api).

## Userinfo

`userinfo_endpoint` answers the same claims as the ID token would carry _now_, for the scopes the access token was granted — so it reflects a name or an email changed since the sign-in.

```bash title="cURL"
curl "$ISSUER/oauth2/userinfo" -H "Authorization: Bearer $ACCESS_TOKEN"
```

```java title="Java"
HttpRequest request = HttpRequest.newBuilder(URI.create(discovery.userinfoEndpoint()))
        .header("Authorization", "Bearer " + accessToken)
        .build();
String claims = HttpClient.newHttpClient().send(request, BodyHandlers.ofString()).body();
```

```go title="Go"
req, _ := http.NewRequest(http.MethodGet, discovery.UserInfoEndpoint, nil)
req.Header.Set("Authorization", "Bearer "+accessToken)
res, err := http.DefaultClient.Do(req)
```

```python title="Python"
claims = requests.get(
    discovery["userinfo_endpoint"],
    headers={"Authorization": f"Bearer {access_token}"},
).json()
```

A token that is missing, expired or not from this server is answered with `401` and a `WWW-Authenticate: Bearer error="invalid_token"` header.

## Lifetimes

Access and ID tokens last an hour unless the application — or, for an access token, the API — sets another lifetime, between a minute and a day. Refresh tokens last thirty days by default, up to a year.
