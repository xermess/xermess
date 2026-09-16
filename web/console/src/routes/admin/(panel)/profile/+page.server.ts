import type { AdminSession, MfaStatus } from '$lib/api';
import { apiGet } from '$lib/server/api';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => {
	const [{ sessions }, { mfa }] = await Promise.all([
		apiGet<{ sessions: AdminSession[] }>('/admin/sessions', fetch),
		apiGet<{ mfa: MfaStatus }>('/admin/mfa', fetch)
	]);

	return { sessions, mfa };
};
