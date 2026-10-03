import { redirect } from '@sveltejs/kit';
import { resolve } from '$app/paths';
import { setupRequired } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * Exists only until the first administrator is created; the API refuses it afterwards anyway.
 */
export const load: PageServerLoad = async ({ fetch }) => {
	if (!(await setupRequired(fetch))) {
		redirect(307, resolve('/admin/login'));
	}

	return {};
};
