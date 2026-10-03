import { infiniteQueryOptions } from '@tanstack/svelte-query';

import { sessionsApi, type UserSessionPage } from '$lib/api';
import { keys } from './keys';

/** How many sessions a page shows before "Show more". */
export const SESSIONS_PAGE_SIZE = 50;

export type SessionListParams = { search: string; user: string };

/**
 * Active sessions as keyset pages, seeded with the server-rendered first page. There is no
 * total, so no page count.
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
