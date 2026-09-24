import { redirect } from '@sveltejs/kit';
import { resolve } from '$app/paths';
import { adminApi, mfaApi } from '$lib/api';
import type { PageServerLoad } from './$types';

/**
 * Setting up an authenticator: the one thing an administrator who has to have
 * one can do before anything else, and what a signed-in one without one comes
 * here for. Anyone else has nothing to set up here.
 */
export const load: PageServerLoad = async ({ fetch }) => {
	const { state } = await adminApi.session(fetch).catch(() => ({ state: 'none' as const }));

	if (state === 'none' || state === 'mfa') redirect(307, resolve('/admin/login'));

	const { mfa } = await mfaApi.status(fetch);
	if (mfa.enabled && state === 'signed_in') redirect(307, resolve('/admin/dashboard'));

	return { required: mfa.required, enrolling: state === 'enroll' };
};
