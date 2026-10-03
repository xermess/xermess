import { redirect } from '@sveltejs/kit';
import { currentUser } from '$lib/server/session';
import { safeNext } from '$lib/utils/next';
import type { PageServerLoad } from './$types';

/**
 * A signed-in user opening the sign-in page directly goes to their account. With a sign-in
 * under way the page shows anyway (the application may have sent prompt=login).
 */
export const load: PageServerLoad = async ({ url, cookies, fetch }) => {
	const next = safeNext(url.searchParams.get('next'));

	if (!url.searchParams.get('request') && (await currentUser(cookies, fetch))) {
		redirect(303, next);
	}

	return { next };
};
