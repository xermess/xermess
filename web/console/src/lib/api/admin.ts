import { api, type Fetch } from './client';
import type {
	API,
	APIApplication,
	APILogEntry,
	APIAccess,
	APIInput,
	TokenPreview,
	TokenPreviewInput,
	Admin,
	ApplicationInput,
	ApplicationPage,
	ApplicationWithSecret,
	AdminInput,
	AdminPage,
	AdminPermission,
	AdminRecord,
	AdminRole,
	AdminRoleInput,
	LoginResult,
	MfaEnrolment,
	MfaStatus,
	SessionState,
	FieldInput,
	FieldRules,
	Organization,
	OrganizationInput,
	Role,
	RoleInput,
	RoleMapping,
	RolePage,
	SetupInput,
	UserField,
	UserInput,
	UserPage,
	UserRecord
} from './types';

/** Setting the panel up: the two calls that work without a session, because
    before the first administrator exists there is nobody to be. */
export const setupApi = {
	status: (fetcher?: Fetch) => api.get<{ required: boolean }>('/admin/setup', fetcher),

	create: (input: SetupInput) => api.post<{ admin: { id: string } }>('/admin/setup', input)
};

/** Signing in and out, from the browser. What the panel *reads* — the
    administrator, the overview, the log — is fetched by the server loads
    instead (lib/server/api.ts), so it is not here. */
export const adminApi = {
	/** A right password signs in, or says which step is still to go. */
	login: (username: string, password: string) =>
		api.post<LoginResult>('/admin/auth/login', { username, password }),

	/** How far this browser's session has got. */
	session: (fetcher?: Fetch) =>
		api.get<{ state: SessionState; mfa_required: boolean }>('/admin/auth/session', fetcher),

	/** Finishes a sign-in waiting for a code, or a recovery code. */
	verifyMfa: (code: string) => api.post<{ admin: Admin }>('/admin/auth/mfa', { code }),

	logout: () => api.post<{ status: string }>('/admin/auth/logout')
};

/** The users an organisation manages, and the shape of their records. */
export const usersApi = {
	/** A page of users. The search and the filter are the same ones the URL
	    carries, so a link and a query key describe the same list. */
	list: (params: { search?: string; verified?: string; role?: string }, fetcher?: Fetch) => {
		const query = new URLSearchParams();
		if (params.search) query.set('search', params.search);
		if (params.verified === 'true' || params.verified === 'false') {
			query.set('verified', params.verified);
		}
		if (params.role) query.set('role', params.role);

		return api.get<UserPage>(`/admin/users?${query}`, fetcher);
	},

	fields: (fetcher?: Fetch) => api.get<{ fields: UserField[] }>('/admin/user-fields', fetcher),

	create: (input: UserInput) => api.post<{ user: UserRecord }>('/admin/users', input),

	/** Every role a user holds: given directly or inherited. */
	roleMappings: (id: string, fetcher?: Fetch) =>
		api.get<{ roles: RoleMapping[] }>(`/admin/users/${id}/role-mappings`, fetcher),

	/** Gives a user roles, global and application roles alike. */
	assignRoles: (id: string, roles: string[]) =>
		api.post<{ roles: RoleMapping[] }>(`/admin/users/${id}/role-mappings`, { roles }),

	/** Takes away a role the user was given directly. */
	unassignRole: (id: string, role: string) =>
		api.delete<{ roles: RoleMapping[] }>(`/admin/users/${id}/role-mappings/${role}`),

	update: (id: string, input: UserInput) =>
		api.patch<{ user: UserRecord }>(`/admin/users/${id}`, input),

	remove: (id: string) => api.delete<void>(`/admin/users/${id}`),

	addField: (field: FieldInput) => api.post<{ field: UserField }>('/admin/user-fields', field),

	updateField: (id: string, rules: FieldRules & { label: string }) =>
		api.patch<{ field: UserField }>(`/admin/user-fields/${id}`, rules),

	removeField: (id: string) => api.delete<void>(`/admin/user-fields/${id}`)
};

/** The organisation this installation belongs to: one record of settings,
    read and written back. */
export const organizationApi = {
	get: (fetcher?: Fetch) => api.get<{ organization: Organization }>('/admin/organization', fetcher),

	update: (input: OrganizationInput) =>
		api.patch<{ organization: Organization }>('/admin/organization', input)
};

/** The roles users hold. */
export const rolesApi = {
	/** A page of roles. `limit` is raised by whatever needs every role at
	    once, such as the pickers in the user and role panels. */
	list: (
		params: {
			scope?: 'global' | 'application';
			application?: string;
			search?: string;
			default?: string;
			limit?: number;
		},
		fetcher?: Fetch
	) => {
		const query = new URLSearchParams();
		if (params.scope) query.set('scope', params.scope);
		if (params.application) query.set('application', params.application);
		if (params.search) query.set('search', params.search);
		if (params.default === 'true' || params.default === 'false') {
			query.set('default', params.default);
		}
		if (params.limit) query.set('limit', String(params.limit));

		return api.get<RolePage>(`/admin/user-roles?${query}`, fetcher);
	},

	create: (input: RoleInput) => api.post<{ role: Role }>('/admin/user-roles', input),

	update: (id: string, input: RoleInput) =>
		api.patch<{ role: Role }>(`/admin/user-roles/${id}`, input),

	remove: (id: string) => api.delete<void>(`/admin/user-roles/${id}`)
};

/** The signed-in administrator's own second factor. */
export const mfaApi = {
	status: (fetcher?: Fetch) => api.get<{ mfa: MfaStatus }>('/admin/mfa', fetcher),

	/** Starts setting up an authenticator. Replacing one takes a code from it. */
	begin: (currentCode?: string) =>
		api.post<{ enrolment: MfaEnrolment }>(
			'/admin/mfa/totp',
			currentCode ? { code: currentCode } : undefined
		),

	/** Finishes it with a code from the new authenticator; answers the recovery codes. */
	confirm: (code: string) =>
		api.post<{ recovery_codes: string[] }>('/admin/mfa/totp/confirm', { code }),

	disable: (code: string) => api.delete<void>('/admin/mfa/totp', { code }),

	recoveryCodes: (code: string) =>
		api.post<{ recovery_codes: string[] }>('/admin/mfa/recovery-codes', { code })
};

/** The panel's administrators, and the roles that say what they may do. Every
    call here is a super admin's alone. */
export const adminsApi = {
	list: (params: { search?: string; status?: string; role?: string }, fetcher?: Fetch) => {
		const query = new URLSearchParams();
		if (params.search) query.set('search', params.search);
		if (params.status) query.set('status', params.status);
		if (params.role) query.set('role', params.role);

		return api.get<AdminPage>(`/admin/admins?${query}`, fetcher);
	},

	create: (input: AdminInput) => api.post<{ admin: AdminRecord }>('/admin/admins', input),

	update: (id: string, input: AdminInput) =>
		api.patch<{ admin: AdminRecord }>(`/admin/admins/${id}`, input),

	remove: (id: string) => api.delete<void>(`/admin/admins/${id}`),

	/** Removes another administrator's second factor and signs them out. */
	resetMfa: (id: string) => api.delete<{ admin: AdminRecord }>(`/admin/admins/${id}/mfa`),

	roles: (search = '', fetcher?: Fetch) =>
		api.get<{ roles: AdminRole[]; total: number }>(
			`/admin/admin-roles${search ? `?search=${encodeURIComponent(search)}` : ''}`,
			fetcher
		),

	createRole: (input: AdminRoleInput) => api.post<{ role: AdminRole }>('/admin/admin-roles', input),

	updateRole: (id: string, input: AdminRoleInput) =>
		api.patch<{ role: AdminRole }>(`/admin/admin-roles/${id}`, input),

	removeRole: (id: string) => api.delete<void>(`/admin/admin-roles/${id}`),

	permissions: (fetcher?: Fetch) =>
		api.get<{ permissions: AdminPermission[] }>('/admin/admin-permissions', fetcher)
};

/** The apps and services that sign their users in here. */
export const applicationsApi = {
	list: (
		params: { search?: string; type?: string; enabled?: string; limit?: number },
		fetcher?: Fetch
	) => {
		const query = new URLSearchParams();
		if (params.search) query.set('search', params.search);
		if (params.type) query.set('type', params.type);
		if (params.enabled === 'true' || params.enabled === 'false')
			query.set('enabled', params.enabled);
		if (params.limit) query.set('limit', String(params.limit));

		return api.get<ApplicationPage>(`/admin/applications?${query}`, fetcher);
	},

	create: (input: ApplicationInput) =>
		api.post<ApplicationWithSecret>('/admin/applications', input),

	update: (id: string, input: ApplicationInput) =>
		api.patch<ApplicationWithSecret>(`/admin/applications/${id}`, input),

	/** Every API, with what the application may do with each. */
	apiAccess: (id: string, fetcher?: Fetch) =>
		api.get<{ apis: APIAccess[] }>(`/admin/applications/${id}/apis`, fetcher),

	/** Authorises the application for an API, with the scopes it may ask for. */
	authorizeAPI: (id: string, apiId: string, scopes: string[]) =>
		api.put<{ apis: APIAccess[] }>(`/admin/applications/${id}/apis/${apiId}`, { scopes }),

	revokeAPI: (id: string, apiId: string) =>
		api.delete<{ apis: APIAccess[] }>(`/admin/applications/${id}/apis/${apiId}`),

	/** What a token request would amount to, without issuing anything. */
	previewToken: (id: string, input: TokenPreviewInput) =>
		api.post<{ preview: TokenPreview }>(`/admin/applications/${id}/token-preview`, input),

	/** Replaces the client secret. The old one stops working at once. */
	rotateSecret: (id: string) => api.post<ApplicationWithSecret>(`/admin/applications/${id}/secret`),

	remove: (id: string) => api.delete<void>(`/admin/applications/${id}`)
};

/** The resource servers tokens are issued for. */
export const apisApi = {
	list: (search = '', fetcher?: Fetch) =>
		api.get<{ apis: API[]; total: number }>(
			`/admin/apis${search ? `?search=${encodeURIComponent(search)}` : ''}`,
			fetcher
		),

	get: (id: string, fetcher?: Fetch) => api.get<{ api: API }>(`/admin/apis/${id}`, fetcher),

	/** The applications the administrator can see, with what each may do
	    with the API. */
	applications: (id: string, fetcher?: Fetch) =>
		api.get<{ applications: APIApplication[] }>(`/admin/apis/${id}/applications`, fetcher),

	logs: (id: string, fetcher?: Fetch) =>
		api.get<{ logs: APILogEntry[] }>(`/admin/apis/${id}/logs`, fetcher),

	create: (input: APIInput) => api.post<{ api: API }>('/admin/apis', input),

	update: (id: string, input: APIInput) => api.patch<{ api: API }>(`/admin/apis/${id}`, input),

	remove: (id: string) => api.delete<void>(`/admin/apis/${id}`)
};
