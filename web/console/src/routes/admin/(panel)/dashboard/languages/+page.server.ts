import type { LanguageList } from '$lib/api';
import { apiGet, requirePermission } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * Every language this installation has, how much of each app it translates,
 * and the shipped languages it does not have. The text itself is only asked
 * for when a language's editor is opened.
 */
export const load: PageServerLoad = async ({ fetch, parent }) => {
	requirePermission((await parent()).admin, 'languages.read');

	const { languages, apps, shipped } = await apiGet<LanguageList>('/admin/languages', fetch);

	return { languages, apps, shipped };
};
