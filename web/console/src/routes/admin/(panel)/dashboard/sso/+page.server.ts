import type { Role, SSOConnection } from '$lib/api';
import { canAnywhere } from '$lib/permissions';
import { apiGet, requirePermission } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * The organisations' identity providers, and the roles a group at one can be
 * mapped to.
 *
 * The roles are only there for an administrator who may read them: the rest
 * see a connection's mappings by the role's id, and cannot change them into
 * something they could not see anyway.
 */
export const load: PageServerLoad = async ({ fetch, parent }) => {
	const { admin } = await parent();
	requirePermission(admin, 'sso.read');

	const mayReadRoles = canAnywhere(admin, 'users.read') || canAnywhere(admin, 'applications.read');

	const [{ connections }, roles] = await Promise.all([
		apiGet<{ connections: SSOConnection[] }>('/admin/sso-connections', fetch),
		mayReadRoles
			? apiGet<{ roles: Role[] }>('/admin/user-roles?limit=200', fetch).then((page) => page.roles)
			: Promise.resolve([] as Role[])
	]);

	return { connections, roles, mayReadRoles };
};
