---
title: Concepts
description: The few words the rest of the docs use.
order: 2
icon: book-open
---

## Application

Anything that signs users in through {{name}} — OAuth calls it a _client_. It has a **client ID**, the **redirect URIs** it may be sent back to, and, if it runs on a server, a **client secret**. Apps that can't keep a secret (single-page, mobile) use PKCE instead.

## Issuer

The server's address, such as `https://id.example.com`. Every endpoint is under it, and every token names it in `iss`.

## Tokens

| Token         | Tells you              | Goes to                            |
| ------------- | ---------------------- | ---------------------------------- |
| ID token      | Who signed in          | Your application                   |
| Access token  | What the caller may do | Your API, or userinfo              |
| Refresh token | —                      | The token endpoint, for new tokens |

See [Tokens](/guides/tokens).

## Scopes

What an application asks for. What it gets is in the token's `scope`.

| Scope            | Gives                               |
| ---------------- | ----------------------------------- |
| `openid`         | An ID token                         |
| `profile`        | `name`, `given_name`, `family_name` |
| `email`          | `email`, `email_verified`           |
| `offline_access` | A refresh token                     |
| `roles`          | `roles` and `global_roles`          |

## API and audience

An **API** is a service your apps call with an access token. It has an **identifier** — usually its URL — and its own scopes, like `orders:read`. Name it in `audience` to get a token for it; an administrator decides which applications may.

## Roles

Each application defines its own roles, and there are **global roles** shared by all. With the `roles` scope, a token carries the user's roles in that application as `roles`, and their global ones as `global_roles`.

## Sessions

Signing in gives the user a session at {{name}}, which is what lets them into your other applications without signing in again. Your application keeps its own session too — see [Signing out](/guides/logout).
