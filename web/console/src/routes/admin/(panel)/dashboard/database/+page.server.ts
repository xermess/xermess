import type { DatabaseTable } from '$lib/api';
import { apiGet, requirePermission } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * The tables this server keeps its records in, with how big each one is.
 *
 * It is a browser and nothing more: there is no endpoint behind this page
 * that writes a row, and the columns holding a password, a key or a token are
 * never read at all.
 */
export const load: PageServerLoad = async ({ fetch, parent }) => {
	requirePermission((await parent()).admin, 'database.read');

	const { tables } = await apiGet<{ tables: DatabaseTable[] }>('/admin/database/tables', fetch);

	return { tables };
};
