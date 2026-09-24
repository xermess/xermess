import { queryOptions } from '@tanstack/svelte-query';

import { mailApi, type MailContent, type MailResponse } from '$lib/api';
import { keys } from './keys';

/** How this installation sends email, seeded with what the server rendered.
    There is one record of it, so the key takes no parameters. */
export function mailOptions(initial: MailResponse) {
	return queryOptions({
		queryKey: keys.mail.settings,
		queryFn: () => mailApi.get(),
		initialData: initial
	});
}

/** The words of every email, in every language the installation has. */
export function mailContentOptions(initial: MailContent) {
	return queryOptions({
		queryKey: keys.mail.content,
		queryFn: () => mailApi.content(),
		initialData: initial
	});
}
