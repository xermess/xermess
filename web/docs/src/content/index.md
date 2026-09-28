---
title: Introduction
description: Sign your users in with {{name}}, an OAuth 2.0 and OpenID Connect server.
order: 1
icon: home-5
---

# Sign your users in with {{name}}

{{name}} is an OAuth 2.0 and OpenID Connect server. Your application sends users to it to sign in, and gets back tokens that say who they are. It speaks the standards, so any OpenID Connect library works with it.

<div class="cards">
  <a href="/start/quickstart"><strong>Quickstart</strong><span>Sign your first user in, step by step.</span></a>
  <a href="/guides/authorization-code"><strong>Sign users in</strong><span>The flow every application uses.</span></a>
  <a href="/guides/protecting-an-api"><strong>Protect an API</strong><span>Check the tokens your API receives.</span></a>
  <a href="/guides/admin-api"><strong>Manage users</strong><span>Create and update users from your code, with admin-cli.</span></a>
  <a href="/guides/account-api"><strong>User self-service</strong><span>Let signed-in users edit their own account in your app.</span></a>
  <a href="/reference/public"><strong>API reference</strong><span>Every endpoint, with examples.</span></a>
</div>

## How it works

1. **Redirect** the user to `/oauth2/authorize`. They sign in on {{name}}'s pages.
2. **Get a code** back on your redirect URI.
3. **Exchange** the code at `/oauth2/token` for an ID token (who they are) and an access token (what they may call).

## Endpoints at a glance

Everything is under your **issuer** — the server's address, such as `https://id.example.com`. The examples write it `$ISSUER`.

| Endpoint                                | For                                                   |
| --------------------------------------- | ----------------------------------------------------- |
| `GET /.well-known/openid-configuration` | Discovery: every other address, and what is supported |
| `GET /oauth2/authorize`                 | Sending the user to sign in                           |
| `POST /oauth2/token`                    | Getting tokens                                        |
| `GET /oauth2/userinfo`                  | Who an access token belongs to                        |
| `GET /.well-known/jwks.json`            | The keys to check token signatures                    |
| `GET /oauth2/logout`                    | Signing the user out                                  |

> [!TIP]
> Give your library `$ISSUER/.well-known/openid-configuration` and let it find the rest.
