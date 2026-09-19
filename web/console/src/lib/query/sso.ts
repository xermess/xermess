import { queryOptions } from '@tanstack/svelte-query';

import { ssoApi, type SSOConnection } from '$lib/api';
import { keys } from './keys';

/** Every SSO connection, seeded with what the server rendered. */
export function ssoOptions(initial: { connections: SSOConnection[] }) {
	return queryOptions({
		queryKey: keys.sso.list,
		queryFn: () => ssoApi.list(),
		initialData: initial
	});
}
