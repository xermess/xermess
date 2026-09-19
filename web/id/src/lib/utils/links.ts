import { resolve } from '$app/paths';

/** The sign-in pages that pass the sign-in handle from one to the next. */
export type AuthPage = '/login' | '/register' | '/forgot-password' | '/sso';

/**
 * A link to another sign-in page that keeps the sign-in under way: without the
 * handle, registering or resetting a password could not send the user back to
 * the application that asked.
 */
export function authHref(page: AuthPage, request: string | null): string {
	const path = resolve(page);
	return request ? `${path}?request=${encodeURIComponent(request)}` : path;
}

/**
 * Where to send the browser to sign in through an organisation's identity
 * provider: the server's own address, which sends it on. The sign-in under way
 * and where to land come along, and so does the address typed, so the
 * provider does not ask for it again.
 */
export function ssoHref(
	slug: string,
	options: { request?: string | null; next?: string | null; email?: string }
): string {
	const query = new URLSearchParams();
	if (options.request) query.set('request', options.request);
	if (options.next) query.set('next', options.next);
	if (options.email) query.set('login_hint', options.email);

	const search = query.toString();
	return `/oauth2/sso/${encodeURIComponent(slug)}/start${search ? `?${search}` : ''}`;
}

/** Sends the browser on to where the API said: usually back to an
    application, on another origin, so it is a full navigation. */
export function leaveTo(location: string): void {
	window.location.assign(location);
}
