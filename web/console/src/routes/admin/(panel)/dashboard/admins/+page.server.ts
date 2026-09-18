import type {
	AdminPage,
	AdminPermission,
	AdminRole,
	AdminSecurity,
	ApplicationPage
} from '$lib/api';
import { APPLICATION_CHOICES_LIMIT } from '$lib/query';
import { apiGet, requirePermission } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * A super admin's page. The search and the filters live in the URL, as on the
 * users page. Every admin role and the permission catalog are loaded too, for
 * the role filter and for the panel, which offers the roles to hold.
 */
export const load: PageServerLoad = async ({ fetch, parent, url }) => {
	requirePermission((await parent()).admin, 'super_admin');

	const search = url.searchParams.get('search')?.trim() ?? '';
	const status = url.searchParams.get('status') ?? '';
	const role = url.searchParams.get('role') ?? '';

	const query = new URLSearchParams();
	if (search) query.set('search', search);
	if (status) query.set('status', status);
	if (role) query.set('role', role);

	const [page, roles, catalog, applications, security] = await Promise.all([
		apiGet<AdminPage>(`/admin/admins?${query}`, fetch),
		apiGet<{ roles: AdminRole[] }>('/admin/admin-roles', fetch),
		apiGet<{ permissions: AdminPermission[] }>('/admin/admin-permissions', fetch),
		apiGet<ApplicationPage>(`/admin/applications?limit=${APPLICATION_CHOICES_LIMIT}`, fetch),
		apiGet<AdminSecurity>('/admin/security', fetch)
	]);

	return {
		page,
		roles: roles.roles,
		catalog: catalog.permissions,
		applications: applications.applications,
		security,
		search,
		status,
		role
	};
};
