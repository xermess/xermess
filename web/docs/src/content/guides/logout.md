---
title: Signing out
description: End the user's session at your application, at the server, or both.
order: 4
section: Sign in
nav: 'Signing out'
icon: logout-box-r
---

There are two sessions to think about: yours, in your application, and the user's session at {{name}}, which is what lets them into your other applications without a password. Signing out of one does not sign them out of the other.

## Out of your application only

Clear your own session, and [revoke](/guides/refresh-tokens#ending-it) the refresh token if you hold one. The next time the user signs in, {{name}} will not ask for their password if their session there is still alive.

## Out of the server too

Send the browser to `end_session_endpoint` (OpenID Connect RP-Initiated Logout):

```http
GET $ISSUER/oauth2/logout
  ?id_token_hint=eyJhbGciOiJSUzI1NiIs…
  &post_logout_redirect_uri=https://app.example.com/signed-out
  &state=xyz
```

| Parameter                  |                                                                                                                                               |
| -------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `id_token_hint`            | An ID token the server issued this user. It may have expired. It says whose session to end.                                                   |
| `client_id`                | Your client ID, if you send no hint                                                                                                           |
| `post_logout_redirect_uri` | Where to go afterwards. It must be one the application registered, and needs `id_token_hint` or `client_id` to say which application that is. |
| `state`                    | Sent back to `post_logout_redirect_uri` unchanged                                                                                             |

Without `post_logout_redirect_uri`, the browser lands on the sign-in app's signed-out page.

> [!IMPORTANT]
> Send `id_token_hint`. Anyone can put a logout link in a page, so a request that names somebody other than whoever this browser is signed in as — or names nobody — leaves the session alone rather than signing out whoever follows the link.

Signing out of the server does not revoke the refresh tokens your other applications hold: each application ends its own session.
