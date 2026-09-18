import type { SocialProvider, SocialSpec } from '$lib/api';
import { apiGet, requirePermission } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * The providers users may sign in with, and the kinds a new one may be.
 *
 * The search box and the filter live in the URL, as on the users page, so the
 * back button walks through them and a filtered list can be linked to. There
 * are few enough providers that the endpoint answers with all of them and the
 * page narrows the list itself — including while it is rendered here, so a
 * shared link arrives already filtered.
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
