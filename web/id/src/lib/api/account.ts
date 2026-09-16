import { request, type Fetch } from './client';

/** What an application's sign-in pages show about it. */
export type Application = {
	name: string;
	logo_uri: string;
	client_uri: string;
	/** The privacy policy, linked under the card and agreed to on registration. */
	policy_uri: string;
	/** The terms of service, the same. */
	tos_uri: string;
	allow_registration: boolean;
};

/** The organisation this server signs users in for: who the account belongs
    to, where to ask for help, and the agreements accepted by making one. An
    application's own name, logo and links come first where it has them; these
    are what is shown for one that has none. Any of it may be empty, and what
    is empty is left out rather than shown blank. */
export type Organization = {
	name: string;
	logo_url: string;
	domain: string;
	support_email: string;
	support_phone: string;
	terms_url: string;
	privacy_url: string;
};

/** A sign-in under way, known to the pages by the handle in `?request=`. */
export type SignInRequest = {
	application: Application;
	login_hint: string;
	expires_at: string;
};

/** What signing in or registering says to do next. */
export type SignedIn = {
	/** Back to the application, with a code. */
	redirect_to?: string;
	/** The password was a temporary one; choose a new one with this token. */
	password_change_required?: boolean;
	reset_token?: string;
};

export type User = {
	id: string;
	email: string;
	email_verified: boolean;
	first_name: string;
	last_name: string;
	created_at: string;
	last_login_at: string | null;
};

export type Session = {
	id: string;
	ip: string;
	user_agent: string;
	signed_in_at: string;
	expires_at: string;
	current: boolean;
};

export type ConnectedApplication = {
	client_id: string;
	name: string;
	logo_uri: string;
	client_uri: string;
	scopes: string[];
	authorized_at: string;
	last_used_at: string;
};

export type RegisterInput = {
	request: string;
	email: string;
	password: string;
	first_name: string;
	last_name: string;
	accept_terms: boolean;
};

const encode = encodeURIComponent;

/** Signing in, registering and resetting a password: no session needed. */
export const signIn = {
	request: (handle: string, fetch?: Fetch) =>
		request<SignInRequest>(`/account/requests/${encode(handle)}`, { fetch }),

	application: (clientId: string, fetch?: Fetch) =>
		request<{ application: Application }>(`/account/applications/${encode(clientId)}`, { fetch }),

	organization: (fetch?: Fetch) =>
		request<{ organization: Organization }>('/account/organization', { fetch }),

	login: (body: { request: string; email: string; password: string }) =>
		request<SignedIn>('/account/login', { method: 'POST', body }),

	register: (body: RegisterInput) =>
		request<SignedIn>('/account/register', { method: 'POST', body }),

	forgotPassword: (body: { request: string; email: string }) =>
		request<{ status: string }>('/account/forgot-password', { method: 'POST', body }),

	checkReset: (token: string, fetch?: Fetch) =>
		request<{ valid: boolean }>(`/account/reset-password?token=${encode(token)}`, { fetch }),

	resetPassword: (body: { token: string; password: string }) =>
		request<{ status: string }>('/account/reset-password', { method: 'POST', body }),

	logout: () => request<{ status: string }>('/account/logout', { method: 'POST' })
};

/** The signed-in user's own account. From a server load, pass its `fetch`:
    the session cookie goes along with it. */
export const account = {
	me: (fetch?: Fetch) => request<{ user: User }>('/account/me', { fetch }),

	updateProfile: (body: { first_name: string; last_name: string }) =>
		request<{ user: User }>('/account/me', { method: 'PATCH', body }),

	changePassword: (body: { current_password: string; new_password: string }) =>
		request<{ status: string }>('/account/password', { method: 'POST', body }),

	sessions: (fetch?: Fetch) => request<{ sessions: Session[] }>('/account/sessions', { fetch }),

	endSession: (id: string) =>
		request<void>(`/account/sessions/${encode(id)}`, { method: 'DELETE' }),

	applications: (fetch?: Fetch) =>
		request<{ applications: ConnectedApplication[] }>('/account/connected-applications', {
			fetch
		}),

	disconnect: (clientId: string) =>
		request<void>(`/account/connected-applications/${encode(clientId)}`, { method: 'DELETE' })
};
