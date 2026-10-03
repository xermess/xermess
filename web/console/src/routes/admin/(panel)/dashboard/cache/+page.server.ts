import type { CacheDatabaseName, CacheKeyPage, CacheKind, CacheOverview } from '$lib/api';
import { CACHE_PAGE_SIZE } from '$lib/query';
import { apiGet, requirePermission } from '$lib/server/api';
import type { PageServerLoad } from './$types';

const kinds: CacheKind[] = ['entry', 'stale', 'generation', 'session', 'ratelimit', 'other'];

/**
 * A super admin's page showing the two Redis databases. The selected database and key filters
 * live in the URL; the first page of keys renders with the page.
 */
export const load: PageServerLoad = async ({ fetch, parent, url }) => {
	requirePermission((await parent()).admin, 'super_admin');

	const database: CacheDatabaseName =
		url.searchParams.get('database') === 'sessions' ? 'sessions' : 'cache';
	const asked = url.searchParams.get('kind') ?? '';
	const kind: CacheKind | '' = kinds.includes(asked as CacheKind) ? (asked as CacheKind) : '';
	const group = url.searchParams.get('group')?.trim() ?? '';
	const search = url.searchParams.get('search')?.trim() ?? '';

	const overview = await apiGet<CacheOverview>('/admin/cache', fetch);

	let first: CacheKeyPage = { keys: [], cursor: '' };
	if (overview.configured) {
		const query = new URLSearchParams({ limit: String(CACHE_PAGE_SIZE) });
		if (kind) query.set('kind', kind);
		if (group) query.set('group', group);
		if (search) query.set('search', search);

		first = await apiGet<CacheKeyPage>(`/admin/cache/${database}/keys?${query}`, fetch);
	}

	return { overview, first, database, kind, group, search };
};
