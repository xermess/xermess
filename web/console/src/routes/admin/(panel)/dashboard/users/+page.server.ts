import type { Application, ApplicationPage, RolePage, UserField, UserPage } from '$lib/api';
import { APPLICATION_CHOICES_LIMIT, ROLE_CHOICES_LIMIT } from '$lib/query';
import { canAnywhere } from '$lib/permissions';
import { apiGet, requirePermission } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * The search and the filter live in the URL, so the server can render the
 * result directly, the browser's back button works, and a filtered list can
 * be linked to.
 *
 * The roles the administrator can see are loaded too — the global roles, and
 * the roles of any application their roles reach — with those applications:
 * the table names each role's scope, and the role mapping offers the roles.
 */
export const load: PageServerLoad = async ({ fetch, parent, url }) => {
	const { admin } = await parent();
	requirePermission(admin, 'users.read');

	const search = url.searchParams.get('search')?.trim() ?? '';
	const verified = url.searchParams.get('verified') ?? '';
	const role = url.searchParams.get('role') ?? '';

	const query = new URLSearchParams();
	if (search) query.set('search', search);
	if (verified === 'true' || verified === 'false') query.set('verified', verified);
	if (role) query.set('role', role);

	const seesApplications = canAnywhere(admin, 'applications.read');

	const [page, fields, applications, roles] = await Promise.all([
		apiGet<UserPage>(`/admin/users?${query}`, fetch),
		apiGet<{ fields: UserField[] }>('/admin/user-fields', fetch),
		seesApplications
			? apiGet<ApplicationPage>(
					`/admin/applications?limit=${APPLICATION_CHOICES_LIMIT}`,
					fetch
				).then((it) => it.applications)
			: Promise.resolve<Application[]>([]),
		apiGet<RolePage>(`/admin/user-roles?limit=${ROLE_CHOICES_LIMIT}`, fetch).then((it) => it.roles)
	]);

	return { page, fields: fields.fields, applications, roles, search, verified, role };
};
