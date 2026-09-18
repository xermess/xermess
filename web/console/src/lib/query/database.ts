import { queryOptions } from '@tanstack/svelte-query';

import { databaseApi, type DatabasePage, type DatabaseTable } from '$lib/api';
import { keys } from './keys';

/** The server's tables, seeded with what the server rendered. */
export function databaseTablesOptions(initial: { tables: DatabaseTable[] }) {
	return queryOptions({
		queryKey: keys.database.tables,
		queryFn: () => databaseApi.tables(),
		initialData: initial
	});
}

/** How many rows of a table one page holds. It is here rather than in the
    page so the key, the request and the pager cannot disagree about it. */
export const DATABASE_PAGE_SIZE = 50;

/** One page of one table, seeded with what the server rendered.

    The offset is part of the key and part of the URL, so paging is a
    navigation: the server fetches the page, the cache keeps it, and going
    back to one already read shows it without asking again. */
export function databaseTableOptions(name: string, offset: number, initial: DatabasePage) {
	return queryOptions({
		queryKey: keys.database.table(name, offset),
		queryFn: () => databaseApi.table(name, { limit: DATABASE_PAGE_SIZE, offset }),
		initialData: initial
	});
}
