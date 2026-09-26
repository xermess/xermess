import type { AdminSession, Overview, OverviewDays } from '$lib/api';
import { can } from '$lib/permissions';
import { apiGet } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * Both calls are independent, so they go out together. The overview is only
 * asked for when the administrator's roles allow reading activity: this is
 * everyone's landing page, so it shows what it can rather than refusing.
 */
export const load: PageServerLoad = async ({ fetch, parent, url }) => {
	const { admin } = await parent();

	// The range the page is looked at over, from the address so it survives
	// a reload and can be shared; anything else is the default fortnight.
	const asked = Number(url.searchParams.get('days'));
	const days: OverviewDays = asked === 7 || asked === 30 || asked === 90 ? asked : 14;

	const [overview, sessions] = await Promise.all([
		can(admin, 'activity.read')
			? apiGet<Overview>(`/admin/overview?days=${days}`, fetch).then(withLists)
			: Promise.resolve(null),
		apiGet<{ sessions: AdminSession[] | null }>('/admin/sessions', fetch)
	]);

	return { days, overview, sessions: sessions.sessions ?? [] };
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
