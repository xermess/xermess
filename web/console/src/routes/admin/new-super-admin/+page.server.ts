import { redirect } from '@sveltejs/kit';
import { resolve } from '$app/paths';
import { setupRequired } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * This page exists only until the panel has an administrator. Once it has
 * one, the way in is the sign-in page — and this form would be a way to make
 * a second super admin without being one, which the API refuses anyway.
 */
export const load: PageServerLoad = async ({ fetch }) => {
	if (!(await setupRequired(fetch))) {
		redirect(307, resolve('/admin/login'));
	}

	return {};
};
