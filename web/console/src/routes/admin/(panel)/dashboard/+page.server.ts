import type { AdminSession, Overview } from '$lib/api';
import { can } from '$lib/permissions';
import { apiGet } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * Both calls are independent, so they go out together. The overview is only
 * asked for when the administrator's roles allow reading activity: this is
 * everyone's landing page, so it shows what it can rather than refusing.
 */
export const load: PageServerLoad = async ({ fetch, parent }) => {
	const { admin } = await parent();

	const [overview, sessions] = await Promise.all([
		can(admin, 'activity.read')
			? apiGet<Overview>('/admin/overview', fetch).then(withLists)
			: Promise.resolve(null),
		apiGet<{ sessions: AdminSession[] | null }>('/admin/sessions', fetch)
	]);

	return { overview, sessions: sessions.sessions ?? [] };
};

/** The overview with every list a list: an API from before these fields
    existed, or one that says null for an empty list, still renders. */
function withLists(overview: Overview): Overview {
	return {
		...overview,
		daily: overview.daily ?? [],
		top_actors: overview.top_actors ?? [],
		activity: overview.activity ?? []
	};
}
