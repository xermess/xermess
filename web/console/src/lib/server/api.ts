import { error, redirect } from '@sveltejs/kit';
import { resolve } from '$app/paths';

import type { Admin, AdminPermissionName } from '$lib/api';
import { can, canAnywhere } from '$lib/permissions';

/**
 * Calls the API as the administrator making this request.
 *
 * It takes the load's own fetch and nothing else: handleFetch in
 * hooks.server.ts is what points that fetch at the admin API's internal
 * address and sends the reader's cookie with it. Everything the panel renders
 * goes through here, which is what lets the pages be rendered on the server
 * instead of assembled in the browser afterwards.
 */
async function call(path: string, fetch: typeof globalThis.fetch) {
	try {
		return await fetch(`/api/v1${path}`);
	} catch {
		error(503, 'Could not reach the admin API. Is it running, and is API_URL right?');
	}
}

/**
 * Reads from the API, sending anyone without a usable session to sign in.
 * Every page behind the panel needs that same treatment, so it lives here.
 */
export async function apiGet<T>(path: string, fetch: typeof globalThis.fetch): Promise<T> {
	const response = await call(path, fetch);

	if (response.status === 401) {
		redirect(307, resolve('/admin/login'));
	}

	if (response.status === 403) {
		error(403, 'Your roles do not allow you to see this page.');
	}

	if (!response.ok) {
		error(response.status, 'The server could not answer that');
	}

	return response.json() as Promise<T>;
}

/**
 * Whether the panel still has to be set up — that is, whether the API has no
 * administrator yet.
 *
 * The sign-in page asks before it draws itself: with no account to sign in
 * to, the only useful thing to show is the form that makes one.
 */
export async function setupRequired(fetch: typeof globalThis.fetch): Promise<boolean> {
	const response = await call('/admin/setup', fetch);

	if (!response.ok) return false;

	const { required } = (await response.json()) as { required: boolean };

	return required;
}

/**
 * Stops a page load for an administrator whose roles do not allow the page,
 * before it asks the API for anything it would only refuse. Pass
 * 'super_admin' for the pages that only a super admin may open.
 */
export function requirePermission(
	admin: Admin,
	permission: AdminPermissionName | 'super_admin'
): void {
	const allowed = permission === 'super_admin' ? admin.is_super_admin : can(admin, permission);

	if (!allowed) {
		error(403, 'Your roles do not allow you to see this page.');
	}
}

/**
 * Stops a page load unless the administrator holds the permission for the
 * whole panel or for at least one application: the pages whose lists are
 * narrowed to the applications an administrator's roles reach.
 */
export function requireAnywhere(admin: Admin, permission: AdminPermissionName): void {
	if (!canAnywhere(admin, permission)) {
		error(403, 'Your roles do not allow you to see this page.');
	}
}
