import { error } from '@sveltejs/kit';
import {
	ApiError,
	signIn,
	type LoginOptions,
	type Organization,
	type SignInRequest,
	type SocialProvider
} from '$lib/api';
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
 * The organisation, the providers people may sign in with and the options of
 * the login flow this sign-in belongs to are loaded whether or not there is a
 * sign-in under way: a reset link opened on its own still names who it is from
 * and whom to ask for help. Failing to load any of them is not worth an error
 * page — the pages show what they have — so they come back empty, and the
 * options fall back to the sign-in these pages have always offered.
 */
export const load: LayoutServerLoad = async ({ url, fetch, setHeaders }) => {
	// A sign-in page is never cached: it is about one sign-in, for one person.
	setHeaders({ 'cache-control': 'no-store' });

	const request = url.searchParams.get('request');

	const [organization, socialProviders, login] = await Promise.all([
		signIn
			.organization(fetch)
			.then(({ organization }): Organization | null => organization)
			.catch(() => null),
		signIn
			.socialProviders(fetch)
			.then(({ providers }): SocialProvider[] => providers)
			.catch((): SocialProvider[] => []),
		signIn
			.loginOptions(request ?? '', fetch)
			.then(({ login }): LoginOptions => login)
			.catch((): LoginOptions => fallback)
	]);

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

	return { request, signInRequest, expired, organization, socialProviders, login };
};

/** What these pages offer when the server cannot say: the sign-in they have
    always offered. A page that hides its way to a new account because one
    request failed would be worse than one that offers a way the server then
    refuses. */
const fallback: LoginOptions = {
	steps: ['identifier', 'password', 'social'],
	allow_registration: true,
	allow_password_reset: true
};
