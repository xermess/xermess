import { queryOptions } from '@tanstack/svelte-query';

import { organizationApi, type Organization } from '$lib/api';
import { keys } from './keys';

/** The organisation, seeded with what the server rendered. There is one of
    it, so the key takes no parameters. */
export function organizationOptions(initial: { organization: Organization }) {
	return queryOptions({
		queryKey: keys.organization.settings,
		queryFn: () => organizationApi.get(),
		initialData: initial
	});
}
