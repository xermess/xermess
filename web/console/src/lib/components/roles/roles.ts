import type { Admin, Application, Role, UserRoleRef } from '$lib/api';
import { can } from '$lib/permissions';

/**
 * Role scopes and inheritance on the client, so a form can show what a change would do before
 * saving. The server resolves the same rules.
 */

/** Every role reached from these ids by following inheritance, the starting
    roles included. An id with no role behind it is skipped. */
export function reachable(from: string[], roles: Role[]): Set<string> {
	const byId = new Map(roles.map((role) => [role.id, role]));
	const seen = new Set<string>();
	const stack = [...from];

	while (stack.length > 0) {
		const id = stack.pop()!;
		const role = byId.get(id);

		if (!role || seen.has(id)) continue;
		seen.add(id);

		for (const inherited of role.inherits) stack.push(inherited.id);
	}

	return seen;
}

/** Whether `role` including `candidate` would make a loop: the candidate is
    the role itself, or already includes it somewhere down the line. */
export function wouldCycle(role: string | undefined, candidate: string, roles: Role[]): boolean {
	if (!role) return false;

	return candidate === role || reachable([candidate], roles).has(role);
}

/** Whether a role of this scope may include `other`, as the server holds it:
    a global role may include any role, and an application role global roles
    and its own application's. */
export function mayInherit(scope: string | null, other: Pick<Role, 'application_id'>): boolean {
	return scope === null || other.application_id === null || other.application_id === scope;
}

/** A role's scope as a heading: "Global", or the application's name. */
export function scopeName(role: Pick<UserRoleRef, 'application_id'>, apps: Application[]): string {
	if (role.application_id === null) return 'Global';

	return apps.find((app) => app.id === role.application_id)?.name ?? 'Unknown application';
}

/** Whether the administrator may give and take away a role: role
    assignments for its application, or for the whole panel when it is
    global. */
export function canAssign(admin: Admin, role: Pick<Role, 'application_id'>): boolean {
	return role.application_id === null
		? can(admin, 'role_assignments.write')
		: can(admin, 'role_assignments.write', role.application_id);
}

/** Whether the administrator may change roles of a scope. */
export function canEditRoles(admin: Admin, scope: string | null): boolean {
	return scope === null ? can(admin, 'user_roles.write') : can(admin, 'user_roles.write', scope);
}

/** Roles grouped by scope — global first, then each application by name —
    for the pickers that list roles of several scopes. */
export function byScope(
	roles: Role[],
	apps: Application[]
): { scope: string | null; name: string; roles: Role[] }[] {
	const groups = new Map<string | null, Role[]>();

	for (const role of roles) {
		const group = groups.get(role.application_id) ?? [];
		group.push(role);
		groups.set(role.application_id, group);
	}

	return [...groups.entries()]
		.map(([scope, group]) => ({
			scope,
			name: scopeName({ application_id: scope }, apps),
			roles: group.sort((a, b) => a.name.localeCompare(b.name))
		}))
		.sort((a, b) => (a.scope === null ? -1 : b.scope === null ? 1 : a.name.localeCompare(b.name)));
}

/** A role name as the server will store it: trimmed and lower case. */
export function tidyName(value: string): string {
	return value.trim().toLowerCase();
}
