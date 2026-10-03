import { infiniteQueryOptions } from '@tanstack/svelte-query';

import { activityApi, type LogFilter, type LogPage } from '$lib/api';
import { keys } from './keys';

/** How many entries a page of the log shows before "Show more". */
export const LOGS_PAGE_SIZE = 50;

/** The log as keyset pages for a filter, seeded with the server-rendered first page. */
export function logsOptions(filter: LogFilter, first: LogPage) {
	return infiniteQueryOptions({
		queryKey: keys.logs.list(filter),
		queryFn: ({ pageParam }) => activityApi.logs(filter, pageParam, LOGS_PAGE_SIZE),
		initialPageParam: '',
		getNextPageParam: (page: LogPage) => page.next || undefined,
		initialData: { pages: [first], pageParams: [''] }
	});
}
