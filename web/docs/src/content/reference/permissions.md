---
title: "Admin permissions"
description: "What each permission an admin role can grant allows, and the routes that check it."
order: 11
icon: shield-keyhole
generated: true
---

<!-- Written by `make docs` from the Go source. Edit the code, not this file. -->

An administrator's roles grant permissions from this list. A role assigned for one application grants its *scopable* permissions for that application only. Managing administrators, their roles, mail, one-time codes, signing keys and the cache is not here: those belong to the super admin role alone.

## Activity

### activity.read {#activity-read}

See the dashboard counts and the activity log.

Checked by:

- [`GET /api/v1/admin/overview`](/reference/admin/activity#get-api-v1-admin-overview) — Is the panel's front page for a range of days: totals, sign-in outcomes, daily activity, the busiest people and the latest entries.
- [`GET /api/v1/admin/logs`](/reference/admin/activity#get-api-v1-admin-logs) — Lists one page of the activity log, newest first, narrowed by the filters parseFilter reads, with the cursor the next page starts from.
- [`GET /api/v1/admin/logs/export`](/reference/admin/activity#get-api-v1-admin-logs-export) — Writes matching entries as CSV, newest first, up to exportLimit, hiding exactly what the logs page hides.

## Users

### users.read {#users-read}

See users and their fields.

Checked by:

- [`GET /api/v1/admin/users`](/reference/admin/users#get-api-v1-admin-users) — Returns a page of users, newest first. `search` matches the email and any field text.
- [`GET /api/v1/admin/users/:id`](/reference/admin/users#get-api-v1-admin-users-id) — Returns one user.
- [`GET /api/v1/admin/users/:id/roles`](/reference/admin/users#get-api-v1-admin-users-id-roles) — Returns the roles a user holds as a token would carry them: global roles and each application's, direct and inherited. `client_id` narrows to one application; applications the administrator cannot see are left out.
- [`GET /api/v1/admin/users/:id/role-mappings`](/reference/admin/users#get-api-v1-admin-users-id-role-mappings) — Returns every role a user holds, direct and inherited, with what each comes through.
- [`GET /api/v1/admin/user-fields`](/reference/admin/fields#get-api-v1-admin-user-fields) — Returns the fields a user record has, in the order they are shown.
- [`GET /api/v1/admin/user-sessions`](/reference/admin/sessions#get-api-v1-admin-user-sessions) — Returns a page of active sessions, newest first.
- [`GET /api/v1/admin/user-roles`](/reference/admin/roles#get-api-v1-admin-user-roles) — Returns a page of roles sorted by name, each with its member count and every role it includes. `scope` is "global", "application" or empty for both; `application` narrows to one application.
- [`GET /api/v1/admin/user-roles/:id`](/reference/admin/roles#get-api-v1-admin-user-roles-id) — Returns one role.

### users.write {#users-write}

Create, edit and delete users.

Checked by:

- [`POST /api/v1/admin/users`](/reference/admin/users#post-api-v1-admin-users) — Adds a user.
- [`PATCH /api/v1/admin/users/:id`](/reference/admin/users#patch-api-v1-admin-users-id) — Replaces a user's email, verified flag and fields, and the password too when a new one is given.
- [`DELETE /api/v1/admin/users/:id`](/reference/admin/users#delete-api-v1-admin-users-id) — Removes a user for good.
- [`DELETE /api/v1/admin/users/:id/social-accounts/:identity`](/reference/admin/users#delete-api-v1-admin-users-id-social-accounts-identity) — Removes a provider a user signs in with; their account and other ways in remain.
- [`DELETE /api/v1/admin/users/:id/sessions`](/reference/admin/sessions#delete-api-v1-admin-users-id-sessions) — Ends every session a user has and revokes every refresh token their applications hold, so they are signed out everywhere at once.
- [`DELETE /api/v1/admin/user-sessions/:id`](/reference/admin/sessions#delete-api-v1-admin-user-sessions-id) — Signs one session out.

### user_fields.write {#user_fields-write}

Add, change and remove user fields.

Checked by:

- [`POST /api/v1/admin/user-fields`](/reference/admin/fields#post-api-v1-admin-user-fields) — Adds a field to every user record.
- [`PATCH /api/v1/admin/user-fields/:id`](/reference/admin/fields#patch-api-v1-admin-user-fields-id) — Changes a field's rules.
- [`DELETE /api/v1/admin/user-fields/:id`](/reference/admin/fields#delete-api-v1-admin-user-fields-id) — Removes a field; stored values are dropped when each user is next saved.

## Organization

### organization.read {#organization-read}

See the organisation this installation belongs to.

Checked by:

- [`GET /api/v1/admin/organization`](/reference/admin/organization#get-api-v1-admin-organization) — Returns the organisation.

### organization.write {#organization-write}

Change the organisation's name, domain, support address and logo.

Checked by:

- [`PATCH /api/v1/admin/organization`](/reference/admin/organization#patch-api-v1-admin-organization) — Changes the settings.

## Authentication

### social.read {#social-read}

See the providers users can sign in with.

Checked by:

- [`GET /api/v1/admin/social-providers`](/reference/admin/social#get-api-v1-admin-social-providers) — Returns every configured provider, and the kinds a new one may be.
- [`GET /api/v1/admin/social-providers/:id`](/reference/admin/social#get-api-v1-admin-social-providers-id) — Returns one provider.

### social.write {#social-write}

Register providers users can sign in with, and change their keys: whoever holds this decides which accounts elsewhere reach this server.

Checked by:

- [`GET /api/v1/admin/social-providers/:id/secret`](/reference/admin/social#get-api-v1-admin-social-providers-id-secret) — Returns the stored client secret so it can be checked against the provider's console.
- [`POST /api/v1/admin/social-providers`](/reference/admin/social#post-api-v1-admin-social-providers) — Registers a provider.
- [`PATCH /api/v1/admin/social-providers/:id`](/reference/admin/social#patch-api-v1-admin-social-providers-id) — Changes a provider's settings.
- [`DELETE /api/v1/admin/social-providers/:id`](/reference/admin/social#delete-api-v1-admin-social-providers-id) — Removes a provider and its identities; the accounts stay, though some may lose their only way in.

### login_flows.read {#login_flows-read}

See the login flows applications sign their users in with.

Checked by:

- [`GET /api/v1/admin/login-flows`](/reference/admin/flows#get-api-v1-admin-login-flows) — Returns every flow, the steps one can be made of, and how many applications each flow signs people in for.
- [`GET /api/v1/admin/login-flows/:id`](/reference/admin/flows#get-api-v1-admin-login-flows-id) — Returns one flow.

### login_flows.write {#login_flows-write}

Write login flows and decide which is the default: whoever holds this decides what a sign-in asks for.

Checked by:

- [`POST /api/v1/admin/login-flows`](/reference/admin/flows#post-api-v1-admin-login-flows) — Adds a flow, disabled until an administrator enables it.
- [`PATCH /api/v1/admin/login-flows/:id`](/reference/admin/flows#patch-api-v1-admin-login-flows-id) — Changes a flow.
- [`DELETE /api/v1/admin/login-flows/:id`](/reference/admin/flows#delete-api-v1-admin-login-flows-id) — Removes a flow.

### sso.read {#sso-read}

See the organisations' identity providers, their domains and how they map roles.

Checked by:

- [`GET /api/v1/admin/sso-connections`](/reference/admin/sso#get-api-v1-admin-sso-connections) — Returns every connection, with how many people sign in through each.
- [`GET /api/v1/admin/sso-connections/:id`](/reference/admin/sso#get-api-v1-admin-sso-connections-id) — Returns one connection.

### sso.write {#sso-write}

Connect, change and remove identity providers: whoever holds this decides who may sign in, as whom, and with which roles.

Checked by:

- [`POST /api/v1/admin/sso-connections`](/reference/admin/sso#post-api-v1-admin-sso-connections) — Adds a connection, disabled until an administrator turns it on.
- [`POST /api/v1/admin/sso-connections/test`](/reference/admin/sso#post-api-v1-admin-sso-connections-test) — Tries a provider before it is relied on: an OpenID Connect issuer's discovery, or a SAML provider's metadata, from its address or as pasted.
- [`PATCH /api/v1/admin/sso-connections/:id`](/reference/admin/sso#patch-api-v1-admin-sso-connections-id) — Changes a connection.
- [`DELETE /api/v1/admin/sso-connections/:id`](/reference/admin/sso#delete-api-v1-admin-sso-connections-id) — Removes a connection and the identities held at it.
- [`POST /api/v1/admin/sso-connections/:id/refresh-metadata`](/reference/admin/sso#post-api-v1-admin-sso-connections-id-refresh-metadata) — Reads a SAML provider's metadata again from its address — for when it has rolled its certificate over.

## Languages

### languages.read {#languages-read}

See the languages and read their text.

Checked by:

- [`GET /api/v1/admin/languages`](/reference/admin/languages#get-api-v1-admin-languages) — Returns every language, with how much of each app it translates, and the shipped languages this installation does not have.
- [`GET /api/v1/admin/languages/:code/translations/:app`](/reference/admin/languages#get-api-v1-admin-languages-code-translations-app) — Returns one language's text for one app, for the editor.

### languages.write {#languages-write}

Add, translate and remove languages, and decide which are offered and the default.

Checked by:

- [`POST /api/v1/admin/languages`](/reference/admin/languages#post-api-v1-admin-languages) — Adds a language, with the text it starts from (createRequest).
- [`PATCH /api/v1/admin/languages/:code`](/reference/admin/languages#patch-api-v1-admin-languages-code) — Changes a language's names and whether and where it is offered.
- [`DELETE /api/v1/admin/languages/:code`](/reference/admin/languages#delete-api-v1-admin-languages-code) — Removes a language and its text.
- [`PUT /api/v1/admin/languages/:code/translations/:app`](/reference/admin/languages#put-api-v1-admin-languages-code-translations-app) — Replaces one language's text for one app.

## APIs

### apis.read {#apis-read}

See APIs, their scopes, and which applications may use them.

Checked by:

- [`GET /api/v1/admin/apis`](/reference/admin/apis#get-api-v1-admin-apis) — Returns every API matching the search, sorted by name, with its scopes and how many applications may use it.
- [`GET /api/v1/admin/apis/:id`](/reference/admin/apis#get-api-v1-admin-apis-id) — Returns one API.
- [`GET /api/v1/admin/apis/:id/applications`](/reference/admin/apis#get-api-v1-admin-apis-id-applications) — Lists visible applications with their access to the API.
- [`GET /api/v1/admin/apis/:id/logs`](/reference/admin/apis#get-api-v1-admin-apis-id-logs) — Lists changes to the API and applications gaining or losing access, newest first.

### apis.write {#apis-write}

Register, change and remove APIs and their scopes.

Checked by:

- [`POST /api/v1/admin/apis`](/reference/admin/apis#post-api-v1-admin-apis) — Registers an API with its scopes.
- [`PATCH /api/v1/admin/apis/:id`](/reference/admin/apis#patch-api-v1-admin-apis-id) — Replaces an API's name, description, role enforcement and scopes.
- [`DELETE /api/v1/admin/apis/:id`](/reference/admin/apis#delete-api-v1-admin-apis-id) — Removes an API with its scopes, application authorisations and role grants.

## Applications

### applications.read {#applications-read}

See applications and the roles they define.
Can be granted for a single application.

Checked by:

- [`GET /api/v1/admin/applications`](/reference/admin/applications#get-api-v1-admin-applications) — Returns a page of the applications the administrator can see, sorted by name, each with how many roles it defines.
- [`GET /api/v1/admin/applications/:id`](/reference/admin/applications#get-api-v1-admin-applications-id) — Returns one application.
- [`PATCH /api/v1/admin/applications/:id`](/reference/admin/applications#patch-api-v1-admin-applications-id) — Replaces an application's settings.
- [`POST /api/v1/admin/applications/:id/secret`](/reference/admin/applications#post-api-v1-admin-applications-id-secret) — Replaces a confidential client's secret.
- [`GET /api/v1/admin/applications/:id/apis`](/reference/admin/applications#get-api-v1-admin-applications-id-apis) — Lists every API with what the application may do with it: whether it is authorised to ask for tokens for it, and which scopes.
- [`PUT /api/v1/admin/applications/:id/apis/:api`](/reference/admin/applications#put-api-v1-admin-applications-id-apis-api) — Lets the application request tokens for an API and replaces the scopes it may ask for: the ceiling on its tokens.
- [`DELETE /api/v1/admin/applications/:id/apis/:api`](/reference/admin/applications#delete-api-v1-admin-applications-id-apis-api) — Stops the application asking for tokens for an API.
- [`POST /api/v1/admin/applications/:id/token-preview`](/reference/admin/applications#post-api-v1-admin-applications-id-token-preview) — Runs the token endpoint's evaluation without issuing anything: whether a token is issued, each scope decision, and the claims.
- [`GET /api/v1/admin/user-roles`](/reference/admin/roles#get-api-v1-admin-user-roles) — Returns a page of roles sorted by name, each with its member count and every role it includes. `scope` is "global", "application" or empty for both; `application` narrows to one application.
- [`GET /api/v1/admin/user-roles/:id`](/reference/admin/roles#get-api-v1-admin-user-roles-id) — Returns one role.

### applications.write {#applications-write}

Change an application's settings and rotate its secret; granted for every application, also create and delete applications.
Can be granted for a single application.

Checked by:

- [`POST /api/v1/admin/applications`](/reference/admin/applications#post-api-v1-admin-applications) — Registers an application.
- [`DELETE /api/v1/admin/applications/:id`](/reference/admin/applications#delete-api-v1-admin-applications-id) — Removes an application, the roles it defines, and everyone's hold on them.
- [`GET /api/v1/admin/apis`](/reference/admin/apis#get-api-v1-admin-apis) — Returns every API matching the search, sorted by name, with its scopes and how many applications may use it.
- [`GET /api/v1/admin/apis/:id`](/reference/admin/apis#get-api-v1-admin-apis-id) — Returns one API.
- [`GET /api/v1/admin/apis/:id/applications`](/reference/admin/apis#get-api-v1-admin-apis-id-applications) — Lists visible applications with their access to the API.

## Roles

### user_roles.write {#user_roles-write}

Create, edit and delete roles: an application's when held for it, and the global roles too when held for the whole panel.
Can be granted for a single application.

Checked by:

- [`POST /api/v1/admin/user-roles`](/reference/admin/roles#post-api-v1-admin-user-roles) — Adds a role: a global one when application_id is null, otherwise one of that application's.
- [`PATCH /api/v1/admin/user-roles/:id`](/reference/admin/roles#patch-api-v1-admin-user-roles-id) — Replaces a role's name, description, default flag and inheritance.
- [`DELETE /api/v1/admin/user-roles/:id`](/reference/admin/roles#delete-api-v1-admin-user-roles-id) — Removes a role for good.

### role_assignments.write {#role_assignments-write}

Give users roles and take them away: an application's when held for it, and the global roles too when held for the whole panel.
Can be granted for a single application.

Checked by:

- [`POST /api/v1/admin/users/:id/role-mappings`](/reference/admin/users#post-api-v1-admin-users-id-role-mappings) — Gives a user roles directly, keeping existing ones.
- [`DELETE /api/v1/admin/users/:id/role-mappings/:role`](/reference/admin/users#delete-api-v1-admin-users-id-role-mappings-role) — Removes a directly held role; an inherited one is removed by removing what it comes through.
