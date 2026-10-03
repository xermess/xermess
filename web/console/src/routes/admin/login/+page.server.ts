import { redirect } from '@sveltejs/kit';
import { resolve } from '$app/paths';
import { adminApi } from '$lib/api';
import { setupRequired } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * Routes by sign-in state: signed in goes to the panel, MFA setup or code entry goes there, and
 * a panel without administrators goes to setup.
 */
export const load: PageServerLoad = async ({ fetch }) => {
	const { state } = await adminApi.session(fetch).catch(() => ({ state: 'none' as const }));

	if (state === 'signed_in') redirect(307, resolve('/admin/dashboard'));
	if (state === 'enroll') redirect(307, resolve('/admin/mfa-setup'));

	if (state === 'none' && (await setupRequired(fetch))) {
		redirect(307, resolve('/admin/new-super-admin'));
	}

	return { waitingForCode: state === 'mfa' };
};
