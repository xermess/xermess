import { error } from '@sveltejs/kit';
import { ApiError, signIn, type Organization, type SignInRequest } from '$lib/api';
import type { LayoutServerLoad } from './$types';

/**
 * The sign-in under way, when the page was given one, and the organisation
 * these pages always speak for.
 *
 * The authorization endpoint sends the browser here with `?request=<handle>`,
 * and every sign-in page passes it on. Loading it once, here, is what lets
 * each page show the application's name, logo and links in the first response.
 * A handle that expired or was used is not an error page: the page says so.
 *
 * The organisation is loaded whether or not there is a sign-in under way: a
 * reset link opened on its own still names who it is from and whom to ask for
 * help. Failing to load it is not worth an error page either — the pages show
 * what they have — so it comes back null.
 */
export const load: LayoutServerLoad = async ({ url, fetch, setHeaders }) => {
	// A sign-in page is never cached: it is about one sign-in, for one person.
	setHeaders({ 'cache-control': 'no-store' });

	const request = url.searchParams.get('request');

	const organization: Organization | null = await signIn
		.organization(fetch)
		.then(({ organization }) => organization)
		.catch(() => null);

	let signInRequest: SignInRequest | null = null;
	let expired = false;

	if (request) {
		try {
			signInRequest = await signIn.request(request, fetch);
		} catch (err) {
			if (err instanceof ApiError && (err.status === 410 || err.status === 404)) {
				expired = true;
			} else if (err instanceof ApiError && err.status === 0) {
				error(503, err.message);
			} else {
				throw err;
			}
		}
	}

	return { request, signInRequest, expired, organization };
};
