---
title: OpenAPI and tooling
description: Import the server's OpenAPI document into your tools, or generate a client from it.
order: 1
section: Tools
nav: 'OpenAPI & clients'
icon: braces
---

Every endpoint the server answers is described in an OpenAPI 3.1 document, and the server serves it itself — with its own address filled in, so the tools you point at it call the right place.

| Document   | Served at                              | Who may fetch it          |
| ---------- | -------------------------------------- | ------------------------- |
| Public API | `GET $ISSUER/.well-known/openapi.json` | Anyone                    |
| Admin API  | `GET $PANEL/api/v1/admin/openapi.json` | A signed-in administrator |

The same documents are downloadable here: [public](/openapi/public.json), [admin](/openapi/admin.json).

## Postman and Bruno

Use **Export** in the header to download a ready-made collection — [Public API](/collections/public.postman_collection.json) or [Admin API](/collections/admin.postman_collection.json). It's Postman's Collection v2.1 format, which Bruno imports too.

1. In Postman or Bruno, choose **Import** and pick the file.
2. Fill in the collection's variables:

| Variable                     | Collection | Is                                                                       |
| ---------------------------- | ---------- | ------------------------------------------------------------------------ |
| `issuer`                     | Both       | The public server's address                                              |
| `client_id`, `client_secret` | Public     | Your application's credentials                                           |
| `access_token`               | Public     | A user's access token — for userinfo and the Account API                 |
| `admin_api`                  | Admin      | The panel's address                                                      |
| `admin_cli_secret`           | Admin      | admin-cli's secret — [set it up](/guides/admin-api)                      |
| `admin_session`              | Admin      | An administrator's session cookie, for the few routes a token can't call |

In the Admin API collection, send **Start here → Get an admin-cli token** first: it saves the token to `admin_token`, which every other request uses.

Requests are grouped the way the [reference](/reference/public) groups them, each with an example body.

## Other tools

- **Insomnia**, **Hoppscotch**, or Postman by URL: import `$ISSUER/.well-known/openapi.json`.
- **Swagger UI or Scalar**, to browse it: give either the URL.

## Generate a client

For the OAuth endpoints, use an OpenID Connect library rather than a generated client: it knows about PKCE, discovery and key rotation, which no schema describes. A generated client earns its keep for the account API.

```bash title="cURL"
# Download the document
curl -o openapi.json "$ISSUER/.well-known/openapi.json"
```

```bash title="Java"
# openapi-generator: a java.net.http client, and Jackson models
npx @openapitools/openapi-generator-cli generate \
  -i "$ISSUER/.well-known/openapi.json" -g java \
  --library native -o account-client
```

```bash title="Go"
go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest \
  -generate types,client -package api openapi.json > api/client.go
```

```bash title="Python"
openapi-python-client generate --url "$ISSUER/.well-known/openapi.json"
```

## What the document says beyond OpenAPI

A few things about a route have no OpenAPI keyword, so they are extensions:

| Extension            | Means                                                                                                                                                      |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `x-rate-limit`       | `sign-in` or `token`: which per-address limit the route is held to. Over it, the answer is `429` with `rate_limited`.                                      |
| `x-permissions`      | The admin permissions the route checks; any one is enough.                                                                                                 |
| `x-permission-scope` | `any-application`: holding the permission for one application is enough to reach the route, and the answer is narrowed to the applications it is held for. |
| `x-super-admin`      | Only a super admin may call it.                                                                                                                            |

Every error answer is the `Problem` schema, with one example per code the route can answer with; the provider's own endpoints answer `OAuthError` instead. The [errors](/reference/errors) page lists every code.

## How the reference is made

Nothing in the reference is written by hand. `make docs` reads the server's Go source with the type checker and writes both the OpenAPI documents and the reference pages here:

- **which routes exist, and their guards** — from the route table, `internal/api/server.go`: the session a route needs, the permission it checks, its rate limit;
- **what each one does** — from its handler's doc comment;
- **what it takes** — from the request type it binds, with each field's comment and its `validate` rules;
- **what it answers** — from the values it writes, and the problems its code can return.

A test compares what is committed with what the code would generate, and fails the build when they differ — and another holds the reference to the routes Gin actually mounts. So a route added without its documentation does not get merged.
