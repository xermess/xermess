import { redirect } from '@sveltejs/kit';
import { currentUser } from '$lib/server/session';
import { safeNext } from '$lib/utils/next';
import type { PageServerLoad } from './$types';

/**
 * "Sign in with SSO" is the sign-in page's other door, so it behaves like it:
 * somebody already signed in, with no application waiting, goes to their
 * account, and `next` is kept for where to land afterwards.
 */
export const load: PageServerLoad = async ({ url, cookies, fetch }) => {
	const next = safeNext(url.searchParams.get('next'));

	if (!url.searchParams.get('request') && (await currentUser(cookies, fetch))) {
		redirect(303, next);
	}

	return { next };
};
