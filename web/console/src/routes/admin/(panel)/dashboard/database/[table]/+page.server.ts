import { error } from '@sveltejs/kit';
import { ApiError, type DatabasePage } from '$lib/api';
import { apiGet, requirePermission } from '$lib/server/api';
import { DATABASE_PAGE_SIZE } from '$lib/query';
import type { PageServerLoad } from './$types';

/**
 * One table: its columns, and the page of rows the URL asks for.
 *
 * The offset is in the URL rather than in the page's own state, so paging is
 * a navigation — the back button walks back through the pages, and a row
 * somebody found can be linked to.
 */
export const load: PageServerLoad = async ({ fetch, params, parent, url }) => {
	requirePermission((await parent()).admin, 'database.read');

	const offset = Math.max(Number(url.searchParams.get('offset') ?? 0) || 0, 0);

	const query = new URLSearchParams({
		limit: String(DATABASE_PAGE_SIZE),
		offset: String(offset)
	});

	try {
		const page = await apiGet<DatabasePage>(
			`/admin/database/tables/${encodeURIComponent(params.table)}?${query}`,
			fetch
		);

		return { page, offset };
	} catch (err) {
		if (err instanceof ApiError && err.status === 404) {
			error(404, `There is no table called ${params.table}.`);
		}

		throw err;
	}
};
