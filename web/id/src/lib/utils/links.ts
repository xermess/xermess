import { resolve } from '$app/paths';

/** The sign-in pages that pass the sign-in handle from one to the next. */
export type AuthPage = '/login' | '/register' | '/forgot-password' | '/sso';

/**
 * A link to another sign-in page that keeps the sign-in handle, so the user can still return to
 * the application.
 */
export function authHref(page: AuthPage, request: string | null): string {
	const path = resolve(page);
	return request ? `${path}?request=${encodeURIComponent(request)}` : path;
}

/**
 * Where to send the browser for SSO: the server's start address, carrying the sign-in handle,
 * the landing page and the typed address.
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
