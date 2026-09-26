import { queryOptions } from '@tanstack/svelte-query';

import { adminApi, mfaApi } from '$lib/api';
import { keys } from './keys';

/** Where the signed-in administrator is signed in. Asked for when the
    account's settings are opened, not before: nothing renders it on the
    server. */
export function profileSessionsOptions() {
	return queryOptions({
		queryKey: keys.profile.sessions,
		queryFn: async () => (await adminApi.profileSessions()).sessions
	});
}

/** Whether the signed-in administrator has a second factor, and how many
    recovery codes are left. */
export function mfaStatusOptions() {
	return queryOptions({
		queryKey: keys.profile.mfa,
		queryFn: async () => (await mfaApi.status()).mfa
	});
}
