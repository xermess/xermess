import type { SocialProvider, SocialSpec } from '$lib/api';
import { apiGet, requirePermission } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * Social providers and the kinds a new one may be. There are few, so all are loaded and
 * filtered on the page (also during SSR); search and filter live in the URL.
 */
export const load: PageServerLoad = async ({ fetch, parent, url }) => {
	requirePermission((await parent()).admin, 'social.read');

	const { providers, kinds } = await apiGet<{
		providers: SocialProvider[];
		kinds: SocialSpec[];
	}>('/admin/social-providers', fetch);

	return {
		providers,
		kinds,
		search: url.searchParams.get('search')?.trim() ?? '',
		status: url.searchParams.get('status') ?? ''
	};
};
