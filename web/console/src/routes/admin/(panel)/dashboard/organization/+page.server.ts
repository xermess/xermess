import type { Organization } from '$lib/api';
import { apiGet, requirePermission } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/** The organisation this installation belongs to. There is one, so the page
    loads a record of settings rather than a list. */
export const load: PageServerLoad = async ({ fetch, parent }) => {
	requirePermission((await parent()).admin, 'organization.read');

	return apiGet<{ organization: Organization }>('/admin/organization', fetch);
};
