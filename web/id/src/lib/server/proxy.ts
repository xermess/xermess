import { env } from '$env/dynamic/private';
import type { HandleFetch } from '@sveltejs/kit';

/**
 * The API paths this app serves on its own origin. In the browser the dev
 * server's proxy, or the reverse proxy in production, routes them to the API.
 */
const API_PATHS = ['/api/v1/account', '/oauth2', '/.well-known'];

/**
 * A fetch made while rendering on the server goes straight to the API at
 * API_URL, rather than out through the public address and back in.
 *
 * Loads call the API by path, the same as the browser does. Here the path is
 * pointed at the API's internal address, and the reader's cookies go with it:
 * a fetch from the server carries none of the browser's on its own.
 */
export const handleFetch: HandleFetch = async ({ event, request, fetch }) => {
	const url = new URL(request.url);

	if (url.origin !== event.url.origin || !API_PATHS.some((path) => url.pathname.startsWith(path))) {
		return fetch(request);
	}

	const headers = new Headers(request.headers);
	const cookie = event.request.headers.get('cookie');
	if (cookie) headers.set('cookie', cookie);
	// Who is really asking, for the API's log. It only believes this from an
	// address in its LOGINER_TRUSTED_PROXIES.
	headers.set('x-forwarded-for', event.getClientAddress());

	const target = new URL(url.pathname + url.search, env.API_URL ?? 'http://localhost:8080');
	const method = request.method;
	const body = method === 'GET' || method === 'HEAD' ? undefined : await request.arrayBuffer();

	return fetch(new Request(target, { method, headers, body, redirect: 'manual' }));
};
