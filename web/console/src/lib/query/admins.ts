import { infiniteQueryOptions, queryOptions } from '@tanstack/svelte-query';

import { adminsApi, type AdminPage, type AdminPermission, type AdminRole } from '$lib/api';
import { keys } from './keys';
import { LIST_PAGE_SIZE, nextOffset } from './paging';

/** What the list of administrators is asked for, all of it from the URL. */
export type AdminListParams = { search: string; status: string; role: string };

/** The administrators matching a search, as pages, seeded with the first
    one the server rendered. */
export function adminsOptions(params: AdminListParams, first: AdminPage) {
	return infiniteQueryOptions({
		queryKey: keys.admins.list(params),
		queryFn: ({ pageParam }) =>
			adminsApi.list({ ...params, offset: pageParam, limit: LIST_PAGE_SIZE }),
		initialPageParam: 0,
		getNextPageParam: (page: AdminPage) => nextOffset(page, page.admins.length),
		initialData: { pages: [first], pageParams: [0] }
	});
}

/** The admin roles matching a search. An empty search is every role, which
    is what the administrator panel offers to hold. */
export function adminRolesOptions(search: string, initial: AdminRole[]) {
	return queryOptions({
		queryKey: keys.admins.roles(search),
		queryFn: async () => (await adminsApi.roles(search)).roles,
		initialData: initial
	});
}

/** The admin permission catalog. It is defined in the server's code, so it
    only changes with a new release. */
export function adminPermissionsOptions(initial: AdminPermission[]) {
	return queryOptions({
		queryKey: keys.admins.permissions,
		queryFn: async () => (await adminsApi.permissions()).permissions,
		initialData: initial,
		staleTime: Infinity
	});
}
