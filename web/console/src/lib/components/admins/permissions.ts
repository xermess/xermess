import type { AdminPermission, AdminPermissionName, AdminRole } from '$lib/api';

/** The catalog split by group, in the order the server lists it. */
export function byGroup(catalog: AdminPermission[]): [string, AdminPermission[]][] {
	const groups = new Map<string, AdminPermission[]>();

	for (const permission of catalog) {
		const group = groups.get(permission.group) ?? [];
		group.push(permission);
		groups.set(permission.group, group);
	}

	return [...groups.entries()];
}

/** Every permission holding these admin roles adds up to, in catalog order.
    super_admin grants the whole catalog. With `scoped`, only the permissions
    that can be scoped count, as for a role held for one application. */
export function grantedBy(
	held: string[],
	roles: AdminRole[],
	catalog: AdminPermission[],
	scoped = false
): AdminPermissionName[] {
	const chosen = roles.filter((role) => held.includes(role.id));
	const names = new Set(chosen.flatMap((role) => role.permissions));
	const everything = chosen.some((role) => role.is_builtin);

	return catalog
		.filter((permission) => !scoped || permission.scopable)
		.map((permission) => permission.name)
		.filter((name) => everything || names.has(name));
}
