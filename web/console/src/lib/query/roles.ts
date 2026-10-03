import { infiniteQueryOptions, queryOptions } from '@tanstack/svelte-query';

import { rolesApi, type Role, type RolePage } from '$lib/api';
import { keys } from './keys';
import { LIST_PAGE_SIZE, nextOffset } from './paging';

/** What the list is asked for, all of it from the URL: the tab (global or
    application roles), the application chip, the search box and the default
    filter. An empty application means every application. */
export type RoleListParams = {
	scope: 'global' | 'application';
	application: string;
	search: string;
	isDefault: string;
};

/** How many roles the pickers ask for: the most the API gives in one page. */
export const ROLE_CHOICES_LIMIT = 500;

/**
 * Roles matching a search as pages; `first` is the server-rendered page, so it is not fetched
 * twice.
 */
export function rolesOptions(params: RoleListParams, first: RolePage) {
	return infiniteQueryOptions({
		queryKey: keys.roles.list(params),
		queryFn: ({ pageParam }) =>
			rolesApi.list({
				scope: params.scope,
				application: params.application || undefined,
				search: params.search,
				default: params.isDefault,
				offset: pageParam,
				limit: LIST_PAGE_SIZE
			}),
		initialPageParam: 0,
		getNextPageParam: (page: RolePage) => nextOffset(page, page.roles.length),
		initialData: { pages: [first], pageParams: [0] }
	});
}

/** Every role of every application the administrator can see, whatever the
    list on screen is filtered to: the user panel offers them to hold, and the
    role panel offers an application's to inherit. */
export function roleChoicesOptions(initial: Role[]) {
	return queryOptions({
		queryKey: keys.roles.choices,
		queryFn: async () => (await rolesApi.list({ limit: ROLE_CHOICES_LIMIT })).roles,
		initialData: initial
	});
}
