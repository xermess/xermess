import type { OTPResponse } from '$lib/api';
import { apiGet, requirePermission } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/** The one-time codes the server emails as people sign in. A super admin's
    page: how short or long-lived a code is decides how hard this server is to
    get into. */
export const load: PageServerLoad = async ({ fetch, parent }) => {
	requirePermission((await parent()).admin, 'super_admin');

	return { otp: await apiGet<OTPResponse>('/admin/otp', fetch) };
};
