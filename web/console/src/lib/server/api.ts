import { error, redirect } from '@sveltejs/kit';
import { resolve } from '$app/paths';

import type { Admin, AdminPermissionName } from '$lib/api';
import { can, canAnywhere } from '$lib/permissions';

/**
 * Calls the API as the administrator making this request, using the load's fetch, which
 * handleFetch points at the internal admin API with the reader's cookie.
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
 * Whether the panel still needs its first administrator; the sign-in page shows the setup form
 * instead.
 */
export async function setupRequired(fetch: typeof globalThis.fetch): Promise<boolean> {
	const response = await call('/admin/setup', fetch);

	if (!response.ok) return false;

	const { required } = (await response.json()) as { required: boolean };

	return required;
}

/**
 * Stops a page load the administrator's roles do not allow, before calling the API. Pass
 * 'super_admin' for super-admin pages.
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
 * Stops a page load unless the administrator holds the permission panel-wide or for at least
 * one application.
 */
export function requireAnywhere(admin: Admin, permission: AdminPermissionName): void {
	if (!canAnywhere(admin, permission)) {
		error(403, 'Your roles do not allow you to see this page.');
	}
}
