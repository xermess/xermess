import type { AdminPermission, AdminRole } from '$lib/api';
import { apiGet, requirePermission } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/** A super admin's page: the roles administrators hold, and the catalog of
    permissions those roles pick from. The search lives in the URL. */
export const load: PageServerLoad = async ({ fetch, parent, url }) => {
	requirePermission((await parent()).admin, 'super_admin');

	const search = url.searchParams.get('search')?.trim() ?? '';
	const query = search ? `?search=${encodeURIComponent(search)}` : '';

	const [roles, catalog] = await Promise.all([
		apiGet<{ roles: AdminRole[] }>(`/admin/admin-roles${query}`, fetch),
		apiGet<{ permissions: AdminPermission[] }>('/admin/admin-permissions', fetch)
	]);

	return { roles: roles.roles, catalog: catalog.permissions, search };
};
