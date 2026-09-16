import type { ApplicationPage } from '$lib/api';
import { apiGet, requireAnywhere } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * The applications this administrator can see. A role held for one
 * application is enough to open the page; the list is narrowed to what their
 * roles reach. The search and the type filter live in the URL.
 */
export const load: PageServerLoad = async ({ fetch, parent, url }) => {
	requireAnywhere((await parent()).admin, 'applications.read');

	const search = url.searchParams.get('search')?.trim() ?? '';
	const type = url.searchParams.get('type') ?? '';

	const query = new URLSearchParams();
	if (search) query.set('search', search);
	if (type) query.set('type', type);

	const page = await apiGet<ApplicationPage>(`/admin/applications?${query}`, fetch);

	return { page, search, type };
};
