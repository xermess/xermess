/** The shapes the API returns. They mirror the Go responses in
    internal/server/admin.go; change them together. */

/** The signed-in administrator. */
export type Admin = {
	id: string;
	username: string;
	email: string;
	full_name: string;
	status: string;
	/** The names of the roles held for the whole panel. */
	roles: string[];
	last_login_at?: string;
	/** What the whole-panel roles allow. The panel shows and hides its
	    controls by these; the server checks them regardless. */
	permissions: AdminPermissionName[];
	/** What roles held for one application add there, by application id. */
	scoped_permissions: Record<string, AdminPermissionName[]>;
	/** Whether they may manage administrators and admin roles. */
	is_super_admin: boolean;
	/** Whether they sign in with a second factor. */
	mfa_enabled: boolean;
};

/** How far the session this browser carries has got. `mfa` is waiting for a
    code; `enroll` has to set up an authenticator before anything else. */
export type SessionState = 'none' | 'signed_in' | 'mfa' | 'enroll';

/** What signing in with a password led to: signed in, or a step to go. */
export type LoginResult =
	{ admin: Admin; next?: undefined } | { next: 'mfa' | 'enroll'; admin?: undefined };

/** The signed-in administrator's own second factor. */
export type MfaStatus = {
	enabled: boolean;
	/** Whether every administrator must have one: it cannot be turned off. */
	required: boolean;
	confirmed_at: string | null;
	last_used_at: string | null;
	recovery_codes_left: number;
};

/** An authenticator being set up: the secret, and the otpauth URI for the QR code. */
export type MfaEnrolment = { secret: string; uri: string };

/** The names in the admin permission catalog. They mirror the constants in
    internal/model/admin_permission.go; change them together. */
export type AdminPermissionName =
	| 'activity.read'
	| 'users.read'
	| 'users.write'
	| 'user_fields.write'
	| 'apis.read'
	| 'apis.write'
	| 'organization.read'
	| 'organization.write'
	| 'social.read'
	| 'social.write'
	| 'login_flows.read'
	| 'login_flows.write'
	| 'sso.read'
	| 'sso.write'
	| 'languages.read'
	| 'languages.write'
	| 'applications.read'
	| 'applications.write'
	| 'user_roles.write'
	| 'role_assignments.write';

/** One entry of the admin permission catalog. */
export type AdminPermission = {
	name: AdminPermissionName;
	group: string;
	description: string;
	/** Whether a role held for one application grants this there. */
	scopable: boolean;
};

/** The status of an administrator's account. Only an active one may sign in. */
export type AdminStatus = 'active' | 'suspended' | 'disabled' | 'invited';

/** One role an administrator holds: for the whole panel when `application`
    is null, or for that one application. */
export type AdminAssignment = {
	id: string;
	role: Reference;
	application: Reference | null;
};

/** An administrator, as a super admin manages them. */
export type AdminRecord = {
	id: string;
	username: string;
	email: string;
	first_name: string;
	last_name: string;
	full_name: string;
	status: AdminStatus;
	assignments: AdminAssignment[];
	permissions: AdminPermissionName[];
	scoped_permissions: Record<string, AdminPermissionName[]>;
	is_super_admin: boolean;
	/** Whether they sign in with a second factor. */
	mfa_enabled: boolean;
	last_login_at: string | null;
	last_login_ip: string;
	created_at: string;
	updated_at: string;
};

/** What the panel sends for one assignment. */
export type AdminAssignmentInput = {
	role_id: string;
	/** Null for the whole panel. */
	application_id: string | null;
};

/** What the panel sends when creating or updating an administrator. The
    address is the account, so there is no username. */
export type AdminInput = {
	email: string;
	first_name: string;
	last_name: string;
	status: Exclude<AdminStatus, 'invited'>;
	/** Required when creating. Left out on an update, the password is kept. */
	password?: string;
	confirm_password?: string;
	/** Every role the administrator holds, replacing what they had. */
	assignments: AdminAssignmentInput[];
};

export type AdminPage = {
	admins: AdminRecord[];
	total: number;
	limit: number;
	offset: number;
};

/** A role administrators hold. */
export type AdminRole = {
	id: string;
	name: string;
	description: string;
	permissions: AdminPermissionName[];
	/** super_admin: grants everything, and cannot be changed or removed. */
	builtin: boolean;
	admin_count: number;
	created_at: string;
	updated_at: string;
};

export type AdminRoleInput = {
	name: string;
	description: string;
	permissions: AdminPermissionName[];
};

/** The kinds of OAuth client: a server-rendered web app, a single-page app, a
    native app on a device, or a service with no user. */
export type ApplicationType = 'web' | 'spa' | 'native' | 'm2m';

export type AuthMethod = 'client_secret_basic' | 'client_secret_post' | 'none';

export type GrantType = 'authorization_code' | 'refresh_token' | 'client_credentials';

export type Scope = 'openid' | 'profile' | 'email' | 'offline_access' | 'roles';

/** An app or service that signs its users in here with OAuth 2.0 and OpenID
    Connect. The names follow RFC 7591's client metadata. */
export type Application = {
	id: string;
	name: string;
	description: string;
	type: ApplicationType;
	logo_uri: string;
	client_uri: string;
	/** The privacy policy and terms of service its sign-in pages link to. */
	policy_uri: string;
	tos_uri: string;
	client_id: string;
	client_id_issued_at: number;
	/** Whether a secret exists. The secret itself is only ever shown once. */
	has_secret: boolean;
	/** The last four characters of the secret, to tell which one is in use. */
	secret_hint: string;
	secret_created_at: string | null;
	token_endpoint_auth_method: AuthMethod;
	grant_types: GrantType[];
	response_types: string[];
	redirect_uris: string[];
	post_logout_redirect_uris: string[];
	scopes: Scope[];
	require_pkce: boolean;
	access_token_lifetime: number;
	id_token_lifetime: number;
	refresh_token_lifetime: number;
	/** Put the user's roles in this application into its tokens. */
	assert_roles: boolean;
	/** Only let users holding one of its roles sign in to it. */
	require_role_assignment: boolean;
	enabled: boolean;
	/** Offer "Create an account" on its sign-in page. */
	allow_registration: boolean;
	/** The login flow it signs people in with, null for the default one. */
	login_flow_id: string | null;
	role_count: number;
	created_at: string;
	updated_at: string;
};

/** What the panel sends when registering or changing an application. `type`
    is only read when registering. */
export type ApplicationInput = Pick<
	Application,
	| 'name'
	| 'description'
	| 'type'
	| 'logo_uri'
	| 'client_uri'
	| 'policy_uri'
	| 'tos_uri'
	| 'token_endpoint_auth_method'
	| 'grant_types'
	| 'redirect_uris'
	| 'post_logout_redirect_uris'
	| 'scopes'
	| 'require_pkce'
	| 'access_token_lifetime'
	| 'id_token_lifetime'
	| 'refresh_token_lifetime'
	| 'assert_roles'
	| 'require_role_assignment'
	| 'enabled'
	| 'allow_registration'
> & {
	/** The flow to sign people in with: an id, or "" for the default one. */
	login_flow_id?: string;
};

/** An application, with the client secret that was just made for it. The
    secret is only ever in this one answer. */
export type ApplicationWithSecret = {
	application: Application;
	client_secret?: string;
};

export type ApplicationPage = {
	applications: Application[];
	total: number;
	limit: number;
	offset: number;
};

/** What the first administrator is made from. There is no username: the
    address is the account. */
export type SetupInput = {
	email: string;
	password: string;
	first_name: string;
	last_name: string;
};

/** What a log entry happened to. `name` is missing when the record is gone,
    or when the administrator's roles do not let them see what it is called. */
export type ActivityTarget = {
	type: string;
	id: string;
	name?: string;
};

export type ActivityEvent = {
	id: string;
	action: string;
	actor: string;
	ip: string;
	target: ActivityTarget | null;
	/** What else is worth a line: why a sign-in was refused, which API. */
	detail?: string;
	created_at: string;
};

export type LogEntry = ActivityEvent & {
	user_agent: string;
};

/** The dashboard. Counts are totals now; `new_users`, `recent_events`, the
    sign-ins and the busiest administrators cover the last seven days. */
export type Overview = {
	counts: {
		users: number;
		active_users: number;
		new_users: number;
		applications: number;
		enabled_applications: number;
		apis: number;
		api_scopes: number;
		user_roles: number;
		admins: number;
		locked_admins: number;
		active_sessions: number;
		events: number;
		recent_events: number;
	};
	sign_ins: {
		since: string;
		succeeded: number;
		failed: number;
		blocked: number;
	};
	/** One entry a day, oldest first, today last. */
	daily: { day: string; events: number; failures: number }[];
	top_actors: { actor: string; events: number }[];
	activity: ActivityEvent[];
};

export type AdminSession = {
	id: string;
	ip: string;
	user_agent: string;
	created_at: string;
	expires_at: string;
	active: boolean;
};

/** The kinds of value a user field can hold. */
export type FieldType = 'text' | 'number' | 'bool' | 'email' | 'date';

/** The rules a field can put on its values. `min` and `max` bound a number's
    value or the length of text; `starts_with` is a prefix text must begin
    with. A rule that is not set is null. */
export type FieldRules = {
	required: boolean;
	unique: boolean;
	min: number | null;
	max: number | null;
	starts_with: string;
};

/** One field of a user record.
 *
 *  `builtin` says which kind it is: a built-in field is a column of the
 *  record — every installation has it, and it cannot be changed or removed —
 *  while an additional one was added in the panel and its values live in
 *  `data`. The API returns both in one list, built-ins first. */
export type UserField = FieldRules & {
	/** The row this field is. A built-in field is a column rather than a row,
	    so it has none. */
	id?: string;
	name: string;
	label: string;
	type: FieldType;
	position: number;
	builtin: boolean;
};

/** What the panel sends when adding a field. A field's name and type are
    fixed once records hold values under them, so an update sends the rules
    only. */
export type FieldInput = FieldRules & {
	name: string;
	label: string;
	type: FieldType;
};

/** The built-in fields of a user record: the columns every installation has. */
export type UserBuiltins = {
	email: string;
	email_verified: boolean;
	first_name: string;
	last_name: string;
	is_active: boolean;
	/** The password is one an administrator set, to be replaced by the user. */
	is_temporary_password: boolean;
};

/** One provider a user signs in with, as their record shows it. */
export type SocialAccount = {
	/** The connection's own id, which disconnecting it names. */
	id: string;
	provider: string;
	slug: string;
	kind: SocialKind;
	email: string;
	connected_at: string;
	last_login_at: string | null;
};

/** A user. The built-in fields are its own properties; everything an
    organisation added lives in `data`, keyed by field name. */
export type UserRecord = UserBuiltins & {
	id: string;
	/** Whether a password is set. The password itself never leaves the server. */
	has_password: boolean;
	/** The roles given to this user directly: global roles, and roles of
	    every application the administrator can see. What they add up to is
	    the role mapping, `usersApi.roleMappings`. */
	roles: UserRoleRef[];
	/** The providers this user signs in with, if any. */
	social_accounts: SocialAccount[];
	data: Record<string, unknown> | null;
	created_at: string;
	updated_at: string;
};

/** What the panel sends when creating or updating a user. */
export type UserInput = UserBuiltins & {
	/** Required when creating a user. Left out on an update, the user keeps
	    the password they have. */
	password?: string;
	/** The password typed again; the server checks the two match. */
	confirm_password?: string;
	data: Record<string, unknown>;
};

export type UserPage = {
	users: UserRecord[];
	total: number;
	limit: number;
	offset: number;
};

/** Another row, named: enough to show it and to send it back. */
export type Reference = {
	id: string;
	name: string;
};

/** A role, and its scope: the application it belongs to, or null for a
    global role. */
export type UserRoleRef = Reference & {
	application_id: string | null;
};

/** One role a user holds, as the role mapping lists it. */
export type RoleMapping = {
	id: string;
	name: string;
	description: string;
	/** The application the role belongs to, or null for a global role. */
	application: { id: string; name: string; client_id: string } | null;
	/** Whether the role includes other roles. */
	composite: boolean;
	/** Whether the role was given to the user directly. */
	assigned: boolean;
	/** The directly given roles it also comes through. A role with `via`
	    and not `assigned` is inherited. */
	via: Reference[];
};

/** A role users hold. What it lets them do is up to each application that
    signs its users in here. */
export type Role = {
	id: string;
	/** The application the role belongs to, or null for a global role. Its
	    name is unique within that scope. */
	application_id: string | null;
	name: string;
	description: string;
	/** Given to every new user, so they can sign in to the application. */
	is_default: boolean;
	/** The roles this one includes directly. */
	inherits: UserRoleRef[];
	/** Every role it includes with inheritance followed, sorted. */
	inherited_roles: UserRoleRef[];
	/** The API scopes the role grants directly. */
	api_scopes: APIScopeRef[];
	/** How many users hold the role directly. */
	user_count: number;
	created_at: string;
	updated_at: string;
};

/** What the panel sends when creating or updating a role. `inherits` holds
    ids and replaces what the role had. */
export type RoleInput = {
	/** The application, or null for a global role. Read when creating only:
	    a role keeps its scope. */
	application_id: string | null;
	name: string;
	description: string;
	is_default: boolean;
	inherits: string[];
	/** Ids of the API scopes the role grants. Left out, the role keeps its
	    grants. */
	api_scopes?: string[];
};

export type RolePage = {
	roles: Role[];
	total: number;
	limit: number;
	offset: number;
};

/** A resource server: a service applications ask for access tokens to call.
    Its identifier is a token's audience. */
export type API = {
	id: string;
	name: string;
	identifier: string;
	description: string;
	/** Grant a user only the scopes their roles grant. */
	enforce_roles: boolean;
	scopes: APIScope[];
	signing_algorithm: SigningAlgorithm;
	/** Seconds; 0 leaves it to each application. */
	token_lifetime: number;
	/** Whether applications may get refresh tokens for it. */
	allow_offline_access: boolean;
	application_count: number;
	/** How many roles grant at least one of its scopes. */
	role_count: number;
	/** The iss an API validating its tokens expects. */
	issuer: string;
	/** Where the keys its tokens are signed with are published. */
	jwks_uri: string;
	created_at: string;
	updated_at: string;
};

/** Asymmetric only: an API checks tokens with public keys. */
export type SigningAlgorithm = 'RS256' | 'PS256' | 'ES256';

export type APIScope = {
	id: string;
	name: string;
	description: string;
	/** Added to every token for the API, when allowed and granted. */
	default: boolean;
};

/** One application, and what it may do with an API. */
export type APIApplication = {
	id: string;
	name: string;
	type: ApplicationType;
	client_id: string;
	enabled: boolean;
	authorized: boolean;
	/** Ids of the API's scopes the application may ask for. */
	allowed: string[];
};

/** One thing that happened to or involving an API. */
export type APILogEntry = {
	id: string;
	action: string;
	actor: string;
	ip: string;
	metadata: Record<string, unknown> | null;
	created_at: string;
};

/** An API scope, and the API it belongs to. */
export type APIScopeRef = {
	id: string;
	name: string;
	api_id: string;
};

/** What the panel sends for an API. A scope sent with its id is kept, with
    everything granting it; one without is new; one left out is removed. The
    identifier is only read when registering. */
export type APIInput = {
	name: string;
	identifier: string;
	description: string;
	enforce_roles: boolean;
	signing_algorithm: SigningAlgorithm;
	token_lifetime: number;
	allow_offline_access: boolean;
	scopes: { id?: string; name: string; description: string; default: boolean }[];
};

/** What one application may do with one API. */
export type APIAccess = {
	id: string;
	name: string;
	identifier: string;
	enforce_roles: boolean;
	authorized: boolean;
	scopes: (APIScope & { allowed: boolean })[];
};

/** A token request to evaluate. */
export type TokenPreviewInput = {
	/** Null for a token the application asks for as itself. */
	user_id: string | null;
	audience: string;
	scope: string;
};

export type ScopeDecision = {
	scope: string;
	kind: 'openid' | 'api';
	granted: boolean;
	reason: string;
	/** Not asked for, but added as one of the API's default scopes. */
	default?: boolean;
};

/** What a token request amounts to. */
export type TokenPreview = {
	issued: boolean;
	reason?: string;
	decisions: ScopeDecision[];
	access_token_header?: Record<string, unknown>;
	access_token?: Record<string, unknown>;
	id_token?: Record<string, unknown>;
};

/** The organisation this installation belongs to. There is one, so it is a
    record of settings rather than a list: it is read and written back.

    It is also what the server shows the outside world: the terms and privacy
    links are published in the discovery document, and the sign-in pages fall
    back to all of it for an application that carries none of its own. */
export type Organization = {
	name: string;
	slug: string;
	/** A host name on its own, or empty. */
	domain: string;
	logo_url: string;
	support_email: string;
	support_phone: string;
	terms_url: string;
	privacy_url: string;
	created_at: string;
};

/** The settings the panel may change: the record without its own bookkeeping. */
export type OrganizationSettings = Omit<Organization, 'created_at'>;

/** What the panel sends for it. Every field is optional because the endpoint
    is a PATCH: what is left out keeps the value it has. */
export type OrganizationInput = Partial<OrganizationSettings>;

/** Which provider a social login record is for. A kind whose provider this
    server knows carries its own endpoints; the two custom ones ask for them. */
export type SocialKind = 'google' | 'apple' | 'facebook' | 'yandex' | 'vk' | 'oidc' | 'oauth2';

/** How the client secret is presented at a provider's token endpoint. */
export type SocialTokenAuth = 'basic' | 'post';

/** What this server knows about a kind of provider, which is what the form
    fills in and what it asks for. */
export type SocialSpec = {
	kind: SocialKind;
	label: string;
	authorize_url: string;
	token_url: string;
	userinfo_url: string;
	scopes: string[];
	/** Whether the endpoints come from the record rather than from the kind. */
	custom: boolean;
	/** Whether the secret is a key this server signs with, as Apple's is. */
	signed_secret: boolean;
	token_auth: SocialTokenAuth;
	/** Where an administrator registers this server with the provider. */
	docs: string;
};

/** A configured provider. The secrets are never sent back: what is stored is
    only ever reported as stored. */
export type SocialProvider = {
	id: string;
	kind: SocialKind;
	slug: string;
	name: string;
	client_id: string;
	has_client_secret: boolean;
	team_id: string;
	key_id: string;
	has_private_key: boolean;
	scopes: string[];
	token_auth: SocialTokenAuth | '';
	/** What that comes to once the kind's default is applied. */
	token_auth_used: SocialTokenAuth;
	authorize_url: string;
	token_url: string;
	userinfo_url: string;
	enabled: boolean;
	link_verified_emails: boolean;
	allow_registration: boolean;
	position: number;
	/** What to register with the provider as the redirect URI. */
	callback_url: string;
	/** How many users sign in with it. */
	identities: number;
	created_at: string;
};

/** What the panel sends for one.

    Everything is optional because the endpoint is a PATCH: a request that
    does not mention a setting leaves it as it is — so turning a provider off
    is `{ enabled: false }` and nothing else. The kind and the slug are read
    when it is registered and never again. */
export type SocialProviderInput = {
	kind?: SocialKind;
	slug?: string;
	name?: string;
	client_id?: string;
	client_secret?: string;
	private_key?: string;
	team_id?: string;
	key_id?: string;
	scopes?: string[];
	token_auth?: SocialTokenAuth | '';
	authorize_url?: string;
	token_url?: string;
	userinfo_url?: string;
	enabled?: boolean;
	link_verified_emails?: boolean;
	allow_registration?: boolean;
};

/** How administrators are made to sign in, and what that means for the ones
    there are. */
export type AdminSecurity = {
	mfa_required: boolean;
	administrators: number;
	with_mfa: number;
};

/** One step a login flow can be made of. The names mirror the constants in
    internal/model/login_flow.go; change them together. */
export type LoginStep =
	'identifier' | 'password' | 'social' | 'email_code' | 'totp' | 'terms' | 'consent';

/** What the server knows about a step: what it is called, and whether the
    sign-in pages run it yet. */
export type LoginStepSpec = {
	step: LoginStep;
	label: string;
	description: string;
	/** Fixed steps cannot be removed or moved. */
	fixed: boolean;
	/** False for a step a flow may name but the server does not run yet. */
	implemented: boolean;
};

/** A login flow: what somebody is taken through when they sign in. */
export type LoginFlow = {
	id: string;
	name: string;
	slug: string;
	description: string;
	/** The flow every application that names none falls back to. */
	is_default: boolean;
	enabled: boolean;
	steps: LoginStep[];
	allow_registration: boolean;
	allow_password_reset: boolean;
	require_verified_email: boolean;
	session_lifetime_hours: number;
	/** How many applications name this flow. */
	applications: number;
	/** The steps it names that the sign-in pages do not run yet. */
	planned: LoginStep[];
	created_at: string;
	updated_at: string;
};

/** What the panel sends for one. Everything is optional because the endpoint
    is a PATCH: a request that mentions one setting changes one setting. */
export type LoginFlowInput = {
	name?: string;
	slug?: string;
	description?: string;
	is_default?: boolean;
	enabled?: boolean;
	steps?: LoginStep[];
	allow_registration?: boolean;
	allow_password_reset?: boolean;
	require_verified_email?: boolean;
	session_lifetime_hours?: number;
};

/** A session still signing a user in, as the Sessions page lists it. */
export type UserSessionRecord = {
	id: string;
	user: { id: string; email: string; name: string };
	ip: string;
	user_agent: string;
	signed_in_at: string;
	expires_at: string;
};

/** A page of sessions, and the session to continue after when there are
    more. There is no total: counting every session on every visit is what
    the page is built to avoid. */
export type UserSessionPage = {
	sessions: UserSessionRecord[];
	next?: string;
};

/** Which of the two apps a translation is for: the sign-in pages, or this
    panel. */
export type LocaleApp = 'id' | 'console';

/** One language this installation has: its settings, and how much of each app
    it translates. */
export type Language = {
	code: string;
	/** The language in English, and in itself. */
	name: string;
	native: string;
	/** The apps this language is translated for: the sign-in pages always,
	    and the admin panel only for English and Russian. */
	apps: LocaleApp[];
	/** How much of each of those apps is translated, as a percentage of the
	    base language's keys, and how many keys each is short. An app the
	    language is not for has no entry. */
	coverage: Partial<Record<LocaleApp, number>>;
	missing: Partial<Record<LocaleApp, number>>;
	/** Whether the sign-in pages offer it, and whether it is the one somebody
	    gets before they have chosen. */
	enabled: boolean;
	is_default: boolean;
	/** Where it comes in the picker, lowest first. */
	position: number;
	/** True for the language every other is a translation of: it cannot be
	    turned off or removed. */
	base: boolean;
	/** True when the server ships a translation of it, so a key a release
	    adds reaches it on the next start. */
	shipped: boolean;
	updated_at: string;
};

/** A language the server ships with that this installation does not have,
    which the panel offers to bring back. */
export type ShippedLanguage = {
	code: string;
	name: string;
	native: string;
};

/** Everything the Languages page lists. */
export type LanguageList = {
	languages: Language[];
	apps: LocaleApp[];
	shipped: ShippedLanguage[];
};

/** What the panel sends to change one. Everything is optional because the
    endpoint is a PATCH: a request that mentions one setting changes one
    setting. */
export type LanguageInput = {
	name?: string;
	native?: string;
	enabled?: boolean;
	is_default?: boolean;
	position?: number;
};

/** What the panel sends to add one. `copy_from` names a language — here, or
    one the server ships — whose text the new one starts as; without it the
    new language starts with nothing translated. */
export type NewLanguageInput = {
	code: string;
	name: string;
	native: string;
	enabled?: boolean;
	is_default?: boolean;
	copy_from?: string;
};

/** One language's text for one app, as the editor needs it: every key there
    is, the base language's text for each, and what this language has. */
export type Translation = {
	app: LocaleApp;
	keys: string[];
	base: Record<string, string>;
	messages: Record<string, string>;
};

/** Which protocol an SSO connection speaks. */
export type SSOProtocol = 'oidc' | 'saml';

/** One group at the identity provider, given one role here. */
export type SSORoleMapping = { group: string; role_id: string };

/** An organisation's own identity provider, as the panel shows it: what is
    stored, never its secrets; how many people sign in through it; what to give
    the provider; and, for SAML, what the provider's metadata says. */
export type SSOConnection = {
	id: string;
	slug: string;
	name: string;
	protocol: SSOProtocol;
	enabled: boolean;
	/** The email domains it signs people in for — and, enforced, the only way
	    in for them. */
	domains: string[];
	enforce_domains: boolean;
	show_on_login: boolean;

	issuer: string;
	client_id: string;
	has_client_secret: boolean;
	scopes: string[];

	metadata_url: string;
	metadata: string;
	name_id_format: 'email' | 'persistent' | 'unspecified';
	sign_requests: boolean;

	/** What to do with an address that already has an account: link it, or
	    refuse. */
	matching: 'link' | 'deny';
	create_users: boolean;
	sync_profile: boolean;
	email_attribute: string;
	first_name_attribute: string;
	last_name_attribute: string;
	groups_attribute: string;
	role_mappings: SSORoleMapping[];
	sync_roles: boolean;

	/** How many people have signed in through it. */
	users: number;
	service_provider: {
		callback_url?: string;
		acs_url?: string;
		entity_id?: string;
		metadata_url?: string;
		certificate?: string;
	};
	identity_provider?: SSOIdentityProvider;
	created_at: string;
	updated_at: string;
};

/** What a SAML provider's metadata says about it. */
export type SSOIdentityProvider = {
	entity_id: string;
	sso_url: string;
	certificates: number;
	certificate_expires?: string;
};

/** What the panel sends to make or change a connection. Everything is
    optional because an update is a PATCH; the client secret is sent only to
    replace it. */
export type SSOConnectionInput = Partial<
	Omit<
		SSOConnection,
		| 'id'
		| 'has_client_secret'
		| 'users'
		| 'service_provider'
		| 'identity_provider'
		| 'created_at'
		| 'updated_at'
	>
> & { client_secret?: string };

/** What trying a provider found. */
export type SSOTestResult = {
	issuer?: string;
	authorization_endpoint?: string;
	token_endpoint?: string;
	jwks_uri?: string;
	unsupported_scopes?: string[];
	identity_provider?: SSOIdentityProvider;
	metadata?: string;
};
