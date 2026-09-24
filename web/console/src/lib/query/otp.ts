import { queryOptions } from '@tanstack/svelte-query';

import { otpApi, type OTPResponse } from '$lib/api';
import { keys } from './keys';

/** The one-time codes the server emails, seeded with what the server
    rendered. There is one record of it, so the key takes no parameters. */
export function otpOptions(initial: OTPResponse) {
	return queryOptions({
		queryKey: keys.otp.settings,
		queryFn: () => otpApi.get(),
		initialData: initial
	});
}
