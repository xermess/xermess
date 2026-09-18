import { queryOptions } from '@tanstack/svelte-query';

import { socialApi, type SocialProvider, type SocialSpec } from '$lib/api';
import { keys } from './keys';

/** The configured providers and the kinds one may be, seeded with what the
    server rendered. */
export function socialProvidersOptions(initial: {
	providers: SocialProvider[];
	kinds: SocialSpec[];
}) {
	return queryOptions({
		queryKey: keys.social.providers,
		queryFn: () => socialApi.list(),
		initialData: initial
	});
}
