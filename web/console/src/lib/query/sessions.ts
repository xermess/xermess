import { infiniteQueryOptions } from '@tanstack/svelte-query';

import { sessionsApi, type UserSessionPage } from '$lib/api';
import { keys } from './keys';

/** How many sessions a page shows before "Show more". */
export const SESSIONS_PAGE_SIZE = 50;

export type SessionListParams = { search: string; user: string };

/**
 * The active sessions as pages, seeded with the first one the server
 * rendered. "Show more" asks for the page after the last session shown —
 * there is no page count to jump to, because there is no total.
 */
export function sessionsOptions(params: SessionListParams, first: UserSessionPage) {
	return infiniteQueryOptions({
		queryKey: keys.sessions.list(params),
		queryFn: ({ pageParam }) =>
			sessionsApi.list({ ...params, after: pageParam, limit: SESSIONS_PAGE_SIZE }),
		initialPageParam: '',
		getNextPageParam: (page: UserSessionPage) => page.next || undefined,
		initialData: { pages: [first], pageParams: [''] }
	});
}
