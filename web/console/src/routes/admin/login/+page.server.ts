import { redirect } from '@sveltejs/kit';
import { resolve } from '$app/paths';
import { adminApi } from '$lib/api';
import { setupRequired } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * Where the reader is in signing in decides what this page shows: someone
 * already signed in goes to the panel, someone who has to set up two-factor
 * sign-in goes to do that, someone waiting for a code is asked for it, and a
 * panel with no administrator yet sends the reader to make the first one.
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
