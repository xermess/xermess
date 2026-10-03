---
title: "Admin API"
description: "Manage users and settings from your own code, with admin-cli, or as the panel does."
order: 1
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

> [!TIP]
> To call it from your code, use [admin-cli](/guides/admin-api): a client credentials token whose scopes are the permissions below. Signing users in to your application is the [public API](/reference/public).

| | |
| --- | --- |
| Base URL | The panel's address — the examples write it `$ADMIN_API` |
| Authentication | An [admin-cli token](/guides/admin-api) (`Authorization: Bearer`), or the panel's `loginer_session` session cookie |
| Permissions | Most routes check one — see [Admin permissions](/reference/permissions) |
| OpenAPI | [Download](/openapi/admin.json), or `GET $ADMIN_API/api/v1/admin/openapi.json` when signed in |

## [Health](/reference/admin/health)

Whether the server is up, for load balancers and uptime checks.

| Endpoint | |
| --- | --- |
| [`GET /healthz`](/reference/admin/health#get-healthz) | Says the server is up. |

## [First administrator](/reference/admin/setup)

Creating the first administrator of a new installation.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/setup`](/reference/admin/setup#get-api-v1-admin-setup) | Says whether the panel still needs its first administrator. |
| [`POST /api/v1/admin/setup`](/reference/admin/setup#post-api-v1-admin-setup) | Makes the first administrator: a super admin, with the address and password whoever is setting the panel up chose. |

## [Signing in](/reference/admin/auth)

Signing an administrator in and out, and who is signed in.

| Endpoint | |
| --- | --- |
| [`POST /api/v1/admin/auth/login`](/reference/admin/auth#post-api-v1-admin-auth-login) | Checks the credentials and sets the session cookie. |
| [`GET /api/v1/admin/auth/session`](/reference/admin/auth#get-api-v1-admin-auth-session) | Reports how far this browser's session has got (password, code, enrolment). |
| [`POST /api/v1/admin/auth/mfa`](/reference/admin/auth#post-api-v1-admin-auth-mfa) | Finishes a sign-in waiting for a second factor. |
| [`POST /api/v1/admin/auth/logout`](/reference/admin/auth#post-api-v1-admin-auth-logout) | Revokes the session and clears the cookie. |
| [`GET /api/v1/admin/me`](/reference/admin/auth#get-api-v1-admin-me) | Returns the signed-in administrator, and tells the browser whether its session is still alive. |
| [`PATCH /api/v1/admin/me`](/reference/admin/auth#patch-api-v1-admin-me) | Changes the caller's own name and address, nothing else, so every administrator may use it. |
| [`POST /api/v1/admin/me/password`](/reference/admin/auth#post-api-v1-admin-me-password) | Sets the caller's own password, given the one they have. |
| [`GET /api/v1/admin/sessions`](/reference/admin/auth#get-api-v1-admin-sessions) | Lists the caller's own sessions, so they can see where they are signed in. |
| [`DELETE /api/v1/admin/sessions/:id`](/reference/admin/auth#delete-api-v1-admin-sessions-id) | Signs one of the administrator's other browsers out. |
| [`DELETE /api/v1/admin/sessions`](/reference/admin/auth#delete-api-v1-admin-sessions) | Signs the administrator out everywhere but here: what to do after using a shared computer, or losing a laptop. |

## [Second factor](/reference/admin/mfa)

An administrator's authenticator app and recovery codes.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/mfa`](/reference/admin/mfa#get-api-v1-admin-mfa) | Describes the administrator's second factor. |
| [`POST /api/v1/admin/mfa/totp`](/reference/admin/mfa#post-api-v1-admin-mfa-totp) | Starts setting up an authenticator app, and answers the secret and the otpauth URI to show as a QR code. |
| [`POST /api/v1/admin/mfa/totp/confirm`](/reference/admin/mfa#post-api-v1-admin-mfa-totp-confirm) | Finishes setting up with a code from the app, and answers the recovery codes — the only time they are shown. |
| [`DELETE /api/v1/admin/mfa/totp`](/reference/admin/mfa#delete-api-v1-admin-mfa-totp) | Turns two-factor sign-in off. |
| [`POST /api/v1/admin/mfa/recovery-codes`](/reference/admin/mfa#post-api-v1-admin-mfa-recovery-codes) | Replaces the recovery codes. |

## [OpenAPI document](/reference/admin/reference)

The server's own OpenAPI document, for your tooling.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/openapi.json`](/reference/admin/reference#get-api-v1-admin-openapi-json) | Answers the OpenAPI document. |

## [Activity](/reference/admin/activity)

The dashboard's counts and the activity log.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/overview`](/reference/admin/activity#get-api-v1-admin-overview) | Is the panel's front page for a range of days: totals, sign-in outcomes, daily activity, the busiest people and the latest entries. |
| [`GET /api/v1/admin/logs`](/reference/admin/activity#get-api-v1-admin-logs) | Lists one page of the activity log, newest first, narrowed by the filters parseFilter reads, with the cursor the next page starts from. |
| [`GET /api/v1/admin/logs/export`](/reference/admin/activity#get-api-v1-admin-logs-export) | Writes matching entries as CSV, newest first, up to exportLimit, hiding exactly what the logs page hides. |

## [Users](/reference/admin/users)

The people your organisation manages.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/users`](/reference/admin/users#get-api-v1-admin-users) | Returns a page of users, newest first. `search` matches the email and any field text. |
| [`GET /api/v1/admin/users/:id`](/reference/admin/users#get-api-v1-admin-users-id) | Returns one user. |
| [`GET /api/v1/admin/users/:id/roles`](/reference/admin/users#get-api-v1-admin-users-id-roles) | Returns the roles a user holds as a token would carry them: global roles and each application's, direct and inherited. `client_id` narrows to one application; applications the administrator cannot see are left out. |
| [`GET /api/v1/admin/users/:id/role-mappings`](/reference/admin/users#get-api-v1-admin-users-id-role-mappings) | Returns every role a user holds, direct and inherited, with what each comes through. |
| [`POST /api/v1/admin/users`](/reference/admin/users#post-api-v1-admin-users) | Adds a user. |
| [`PATCH /api/v1/admin/users/:id`](/reference/admin/users#patch-api-v1-admin-users-id) | Replaces a user's email, verified flag and fields, and the password too when a new one is given. |
| [`DELETE /api/v1/admin/users/:id`](/reference/admin/users#delete-api-v1-admin-users-id) | Removes a user for good. |
| [`DELETE /api/v1/admin/users/:id/social-accounts/:identity`](/reference/admin/users#delete-api-v1-admin-users-id-social-accounts-identity) | Removes a provider a user signs in with; their account and other ways in remain. |
| [`POST /api/v1/admin/users/:id/role-mappings`](/reference/admin/users#post-api-v1-admin-users-id-role-mappings) | Gives a user roles directly, keeping existing ones. |
| [`DELETE /api/v1/admin/users/:id/role-mappings/:role`](/reference/admin/users#delete-api-v1-admin-users-id-role-mappings-role) | Removes a directly held role; an inherited one is removed by removing what it comes through. |

## [User fields](/reference/admin/fields)

The custom fields a user record is made of.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/user-fields`](/reference/admin/fields#get-api-v1-admin-user-fields) | Returns the fields a user record has, in the order they are shown. |
| [`POST /api/v1/admin/user-fields`](/reference/admin/fields#post-api-v1-admin-user-fields) | Adds a field to every user record. |
| [`PATCH /api/v1/admin/user-fields/:id`](/reference/admin/fields#patch-api-v1-admin-user-fields-id) | Changes a field's rules. |
| [`DELETE /api/v1/admin/user-fields/:id`](/reference/admin/fields#delete-api-v1-admin-user-fields-id) | Removes a field; stored values are dropped when each user is next saved. |

## [User sessions](/reference/admin/sessions)

Who is signed in, and signing them out.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/user-sessions`](/reference/admin/sessions#get-api-v1-admin-user-sessions) | Returns a page of active sessions, newest first. |
| [`DELETE /api/v1/admin/users/:id/sessions`](/reference/admin/sessions#delete-api-v1-admin-users-id-sessions) | Ends every session a user has and revokes every refresh token their applications hold, so they are signed out everywhere at once. |
| [`DELETE /api/v1/admin/user-sessions/:id`](/reference/admin/sessions#delete-api-v1-admin-user-sessions-id) | Signs one session out. |

## [Organization](/reference/admin/organization)

The organisation's name, branding and agreements.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/organization`](/reference/admin/organization#get-api-v1-admin-organization) | Returns the organisation. |
| [`PATCH /api/v1/admin/organization`](/reference/admin/organization#patch-api-v1-admin-organization) | Changes the settings. |

## [Social providers](/reference/admin/social)

Signing in with an account at another provider.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/social-providers`](/reference/admin/social#get-api-v1-admin-social-providers) | Returns every configured provider, and the kinds a new one may be. |
| [`GET /api/v1/admin/social-providers/:id`](/reference/admin/social#get-api-v1-admin-social-providers-id) | Returns one provider. |
| [`GET /api/v1/admin/social-providers/:id/secret`](/reference/admin/social#get-api-v1-admin-social-providers-id-secret) | Returns the stored client secret so it can be checked against the provider's console. |
| [`POST /api/v1/admin/social-providers`](/reference/admin/social#post-api-v1-admin-social-providers) | Registers a provider. |
| [`PATCH /api/v1/admin/social-providers/:id`](/reference/admin/social#patch-api-v1-admin-social-providers-id) | Changes a provider's settings. |
| [`DELETE /api/v1/admin/social-providers/:id`](/reference/admin/social#delete-api-v1-admin-social-providers-id) | Removes a provider and its identities; the accounts stay, though some may lose their only way in. |

## [Single sign-on](/reference/admin/sso)

Signing in through a customer's own identity provider, over OpenID Connect or SAML.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/sso-connections`](/reference/admin/sso#get-api-v1-admin-sso-connections) | Returns every connection, with how many people sign in through each. |
| [`GET /api/v1/admin/sso-connections/:id`](/reference/admin/sso#get-api-v1-admin-sso-connections-id) | Returns one connection. |
| [`POST /api/v1/admin/sso-connections`](/reference/admin/sso#post-api-v1-admin-sso-connections) | Adds a connection, disabled until an administrator turns it on. |
| [`POST /api/v1/admin/sso-connections/test`](/reference/admin/sso#post-api-v1-admin-sso-connections-test) | Tries a provider before it is relied on: an OpenID Connect issuer's discovery, or a SAML provider's metadata, from its address or as pasted. |
| [`PATCH /api/v1/admin/sso-connections/:id`](/reference/admin/sso#patch-api-v1-admin-sso-connections-id) | Changes a connection. |
| [`DELETE /api/v1/admin/sso-connections/:id`](/reference/admin/sso#delete-api-v1-admin-sso-connections-id) | Removes a connection and the identities held at it. |
| [`POST /api/v1/admin/sso-connections/:id/refresh-metadata`](/reference/admin/sso#post-api-v1-admin-sso-connections-id-refresh-metadata) | Reads a SAML provider's metadata again from its address — for when it has rolled its certificate over. |

## [Login flows](/reference/admin/flows)

The steps a sign-in goes through, per application.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/login-flows`](/reference/admin/flows#get-api-v1-admin-login-flows) | Returns every flow, the steps one can be made of, and how many applications each flow signs people in for. |
| [`GET /api/v1/admin/login-flows/:id`](/reference/admin/flows#get-api-v1-admin-login-flows-id) | Returns one flow. |
| [`POST /api/v1/admin/login-flows`](/reference/admin/flows#post-api-v1-admin-login-flows) | Adds a flow, disabled until an administrator enables it. |
| [`PATCH /api/v1/admin/login-flows/:id`](/reference/admin/flows#patch-api-v1-admin-login-flows-id) | Changes a flow. |
| [`DELETE /api/v1/admin/login-flows/:id`](/reference/admin/flows#delete-api-v1-admin-login-flows-id) | Removes a flow. |

## [Languages](/reference/admin/languages)

The sign-in pages' languages and their text.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/languages`](/reference/admin/languages#get-api-v1-admin-languages) | Returns every language, with how much of each app it translates, and the shipped languages this installation does not have. |
| [`GET /api/v1/admin/languages/:code/translations/:app`](/reference/admin/languages#get-api-v1-admin-languages-code-translations-app) | Returns one language's text for one app, for the editor. |
| [`POST /api/v1/admin/languages`](/reference/admin/languages#post-api-v1-admin-languages) | Adds a language, with the text it starts from (createRequest). |
| [`PATCH /api/v1/admin/languages/:code`](/reference/admin/languages#patch-api-v1-admin-languages-code) | Changes a language's names and whether and where it is offered. |
| [`DELETE /api/v1/admin/languages/:code`](/reference/admin/languages#delete-api-v1-admin-languages-code) | Removes a language and its text. |
| [`PUT /api/v1/admin/languages/:code/translations/:app`](/reference/admin/languages#put-api-v1-admin-languages-code-translations-app) | Replaces one language's text for one app. |

## [Applications](/reference/admin/applications)

The applications that sign users in: their settings, secrets and API access.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/applications`](/reference/admin/applications#get-api-v1-admin-applications) | Returns a page of the applications the administrator can see, sorted by name, each with how many roles it defines. |
| [`GET /api/v1/admin/applications/:id`](/reference/admin/applications#get-api-v1-admin-applications-id) | Returns one application. |
| [`PATCH /api/v1/admin/applications/:id`](/reference/admin/applications#patch-api-v1-admin-applications-id) | Replaces an application's settings. |
| [`POST /api/v1/admin/applications/:id/secret`](/reference/admin/applications#post-api-v1-admin-applications-id-secret) | Replaces a confidential client's secret. |
| [`GET /api/v1/admin/applications/:id/apis`](/reference/admin/applications#get-api-v1-admin-applications-id-apis) | Lists every API with what the application may do with it: whether it is authorised to ask for tokens for it, and which scopes. |
| [`PUT /api/v1/admin/applications/:id/apis/:api`](/reference/admin/applications#put-api-v1-admin-applications-id-apis-api) | Lets the application request tokens for an API and replaces the scopes it may ask for: the ceiling on its tokens. |
| [`DELETE /api/v1/admin/applications/:id/apis/:api`](/reference/admin/applications#delete-api-v1-admin-applications-id-apis-api) | Stops the application asking for tokens for an API. |
| [`POST /api/v1/admin/applications/:id/token-preview`](/reference/admin/applications#post-api-v1-admin-applications-id-token-preview) | Runs the token endpoint's evaluation without issuing anything: whether a token is issued, each scope decision, and the claims. |
| [`POST /api/v1/admin/applications`](/reference/admin/applications#post-api-v1-admin-applications) | Registers an application. |
| [`DELETE /api/v1/admin/applications/:id`](/reference/admin/applications#delete-api-v1-admin-applications-id) | Removes an application, the roles it defines, and everyone's hold on them. |

## [User roles](/reference/admin/roles)

The roles users hold, globally and in each application.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/user-roles`](/reference/admin/roles#get-api-v1-admin-user-roles) | Returns a page of roles sorted by name, each with its member count and every role it includes. `scope` is "global", "application" or empty for both; `application` narrows to one application. |
| [`GET /api/v1/admin/user-roles/:id`](/reference/admin/roles#get-api-v1-admin-user-roles-id) | Returns one role. |
| [`POST /api/v1/admin/user-roles`](/reference/admin/roles#post-api-v1-admin-user-roles) | Adds a role: a global one when application_id is null, otherwise one of that application's. |
| [`PATCH /api/v1/admin/user-roles/:id`](/reference/admin/roles#patch-api-v1-admin-user-roles-id) | Replaces a role's name, description, default flag and inheritance. |
| [`DELETE /api/v1/admin/user-roles/:id`](/reference/admin/roles#delete-api-v1-admin-user-roles-id) | Removes a role for good. |

## [APIs](/reference/admin/apis)

The APIs access tokens are issued for, and their scopes.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/apis`](/reference/admin/apis#get-api-v1-admin-apis) | Returns every API matching the search, sorted by name, with its scopes and how many applications may use it. |
| [`GET /api/v1/admin/apis/:id`](/reference/admin/apis#get-api-v1-admin-apis-id) | Returns one API. |
| [`GET /api/v1/admin/apis/:id/applications`](/reference/admin/apis#get-api-v1-admin-apis-id-applications) | Lists visible applications with their access to the API. |
| [`GET /api/v1/admin/apis/:id/logs`](/reference/admin/apis#get-api-v1-admin-apis-id-logs) | Lists changes to the API and applications gaining or losing access, newest first. |
| [`POST /api/v1/admin/apis`](/reference/admin/apis#post-api-v1-admin-apis) | Registers an API with its scopes. |
| [`PATCH /api/v1/admin/apis/:id`](/reference/admin/apis#patch-api-v1-admin-apis-id) | Replaces an API's name, description, role enforcement and scopes. |
| [`DELETE /api/v1/admin/apis/:id`](/reference/admin/apis#delete-api-v1-admin-apis-id) | Removes an API with its scopes, application authorisations and role grants. |

## [Administrators](/reference/admin/admins)

The panel's administrators.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/security`](/reference/admin/admins#get-api-v1-admin-security) | Returns the administrators' sign-in settings and how many have an authenticator. |
| [`PATCH /api/v1/admin/security`](/reference/admin/admins#patch-api-v1-admin-security) | Changes those settings. |
| [`GET /api/v1/admin/admins`](/reference/admin/admins#get-api-v1-admin-admins) | Returns a page of administrators, newest first. |
| [`POST /api/v1/admin/admins`](/reference/admin/admins#post-api-v1-admin-admins) | Adds an administrator, with the password and roles the super admin chose for them. |
| [`GET /api/v1/admin/admins/:id`](/reference/admin/admins#get-api-v1-admin-admins-id) | Returns one administrator. |
| [`PATCH /api/v1/admin/admins/:id`](/reference/admin/admins#patch-api-v1-admin-admins-id) | Replaces an administrator's details, status, roles and optionally password. |
| [`DELETE /api/v1/admin/admins/:id`](/reference/admin/admins#delete-api-v1-admin-admins-id) | Removes an administrator for good. |
| [`DELETE /api/v1/admin/admins/:id/mfa`](/reference/admin/admins#delete-api-v1-admin-admins-id-mfa) | Removes another administrator's second factor and signs them out everywhere. |

## [Mail](/reference/admin/mail)

How email is sent, and what each message says.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/mail`](/reference/admin/mail#get-api-v1-admin-mail) | Returns the mail settings. |
| [`PATCH /api/v1/admin/mail`](/reference/admin/mail#patch-api-v1-admin-mail) | Changes the settings; omitted fields stay, including the password unless a new one is typed. |
| [`POST /api/v1/admin/mail/test`](/reference/admin/mail#post-api-v1-admin-mail-test) | Sends one message with the form's settings, so a server can be tried before saving. |
| [`GET /api/v1/admin/mail/content`](/reference/admin/mail#get-api-v1-admin-mail-content) | Returns every email's text in each offered language, with English beside it. |
| [`PUT /api/v1/admin/mail/content/:code`](/reference/admin/mail#put-api-v1-admin-mail-content-code) | Writes one language's email text, merging only the email keys. |

## [One-time codes](/reference/admin/otp)

How the one-time codes sent by email behave.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/otp`](/reference/admin/otp#get-api-v1-admin-otp) | Returns the settings and the login flows that use emailed codes. |
| [`PATCH /api/v1/admin/otp`](/reference/admin/otp#patch-api-v1-admin-otp) | Changes the settings. |

## [Signing keys](/reference/admin/keys)

The keys tokens are signed with, and rotating them.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/signing-keys`](/reference/admin/keys#get-api-v1-admin-signing-keys) | Describes every published key: the one signing, the next one waiting, and the retired ones still published. |
| [`POST /api/v1/admin/signing-keys/rotate`](/reference/admin/keys#post-api-v1-admin-signing-keys-rotate) | Makes new keys now. |

## [Cache](/reference/admin/caching)

What the Redis cache holds, and clearing it.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/cache`](/reference/admin/caching#get-api-v1-admin-cache) | Sums up both databases: how many keys of each kind, each group's generation and what it holds, and the Redis server they share. |
| [`GET /api/v1/admin/cache/:database/keys`](/reference/admin/caching#get-api-v1-admin-cache-database-keys) | Lists a page of one database's keys. |
| [`GET /api/v1/admin/cache/:database/key`](/reference/admin/caching#get-api-v1-admin-cache-database-key) | Answers one key and what it holds. |
| [`PUT /api/v1/admin/cache/:database/key`](/reference/admin/caching#put-api-v1-admin-cache-database-key) | Replaces a cached value by hand. |
| [`DELETE /api/v1/admin/cache/:database/key`](/reference/admin/caching#delete-api-v1-admin-cache-database-key) | Removes one key. |
| [`POST /api/v1/admin/cache/:database/groups/:group/clear`](/reference/admin/caching#post-api-v1-admin-cache-database-groups-group-clear) | Forgets everything cached in one group, for every server process at once. |
| [`DELETE /api/v1/admin/cache/:database`](/reference/admin/caching#delete-api-v1-admin-cache-database) | Removes every key this server keeps in one database. |

## [Administrator roles](/reference/admin/adminroles)

What administrators may do in the panel.

| Endpoint | |
| --- | --- |
| [`GET /api/v1/admin/admin-permissions`](/reference/admin/adminroles#get-api-v1-admin-admin-permissions) | Returns the catalog roles pick from, in the order the panel lists it. |
| [`GET /api/v1/admin/admin-roles`](/reference/admin/adminroles#get-api-v1-admin-admin-roles) | Returns every admin role matching the search, sorted by name, with how many administrators hold each. |
| [`POST /api/v1/admin/admin-roles`](/reference/admin/adminroles#post-api-v1-admin-admin-roles) | Adds an admin role. |
| [`PATCH /api/v1/admin/admin-roles/:id`](/reference/admin/adminroles#patch-api-v1-admin-admin-roles-id) | Replaces an admin role's name, description and permissions. |
| [`DELETE /api/v1/admin/admin-roles/:id`](/reference/admin/adminroles#delete-api-v1-admin-admin-roles-id) | Removes an admin role. |
