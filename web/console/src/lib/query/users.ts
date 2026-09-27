import { infiniteQueryOptions, queryOptions } from '@tanstack/svelte-query';

import { usersApi, type UserField, type UserPage } from '$lib/api';
import { keys } from './keys';
import { LIST_PAGE_SIZE, nextOffset } from './paging';

/** What the list is asked for: the search box, the verified filter and the
    role filter, all of which live in the URL. */
export type UserListParams = { search: string; verified: string; role: string };

/**
 * The users matching a search, as pages: "Show more" asks for the one after
 * the rows already shown.
 *
 * `first` is the page the server already rendered, so opening the page does
 * not fetch the same rows a second time. It belongs to this key only: a new
 * search is a new key, and a new page load, which brings its own.
 */
export function usersOptions(params: UserListParams, first: UserPage) {
	return infiniteQueryOptions({
		queryKey: keys.users.list(params),
		queryFn: ({ pageParam }) =>
			usersApi.list({ ...params, offset: pageParam, limit: LIST_PAGE_SIZE }),
		initialPageParam: 0,
		getNextPageParam: (page: UserPage) => nextOffset(page, page.users.length),
		initialData: { pages: [first], pageParams: [0] }
	});
}

/** The fields every user record has. They change rarely, and everything on
    the page needs them, so they are one query the whole page shares. */
export function userFieldsOptions(initial: UserField[]) {
	return queryOptions({
		queryKey: keys.users.fields,
		queryFn: async () => (await usersApi.fields()).fields,
		initialData: initial
	});
}
