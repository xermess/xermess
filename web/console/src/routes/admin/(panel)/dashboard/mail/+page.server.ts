import type { MailContent, MailResponse } from '$lib/api';
import { apiGet, requirePermission } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/** How this installation sends email, and the words of every message it
    sends. A super admin's page: the settings carry the mail server's
    password, and the words are what lands in a user's inbox. */
export const load: PageServerLoad = async ({ fetch, parent }) => {
	requirePermission((await parent()).admin, 'super_admin');

	const [mail, content] = await Promise.all([
		apiGet<MailResponse>('/admin/mail', fetch),
		apiGet<MailContent>('/admin/mail/content', fetch)
	]);

	return { mail, content };
};
