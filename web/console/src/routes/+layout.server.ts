import { COOKIES, LANGUAGE_DEPENDENCY } from '$lib/constants';
import { chooseLanguage, type Messages } from '$lib/i18n';
import type { LanguageChoice } from '$lib/state/language.svelte';
import type { LayoutServerLoad } from './$types';

/**
 * The language this panel is drawn in, its text, and the languages it can be
 * drawn in.
 *
 * The language is one administrator's own preference on one machine — the
 * Languages page decides what users are offered, not what an administrator
 * reads the panel in — so it is a cookie rather than a setting. It is
 * resolved on the server so the first response arrives already translated,
 * and the text comes from the API rather than the build, so a language added
 * or reworded on the Languages page is on the next page.
 *
 * Both endpoints take no session, since the sign-in page is drawn in the same
 * language. When they cannot be reached the panel falls back to the base
 * language it was built with, rather than to an error page.
 */
export const load: LayoutServerLoad = async ({ cookies, depends, fetch, locals, request }) => {
	depends(LANGUAGE_DEPENDENCY);

	const panelLanguages = await read<{ languages: LanguageChoice[] }>(
		fetch,
		'/api/v1/admin/panel/languages'
	).then((body) => body?.languages ?? []);

	const language = chooseLanguage(
		cookies.get(COOKIES.language),
		request.headers.get('accept-language'),
		panelLanguages.map((one) => one.code)
	);

	const messages: Messages = panelLanguages.length
		? ((
				await read<{ messages: Messages }>(
					fetch,
					`/api/v1/admin/panel/languages/${encodeURIComponent(language)}`
				)
			)?.messages ?? {})
		: {};

	// hooks.server.ts writes this into <html lang> once the page is rendered.
	locals.language = language;

	return { language, panelLanguages, messages };
};

/** A GET that answers null rather than throwing: nothing here is worth an
    error page. */
async function read<T>(fetch: typeof globalThis.fetch, path: string): Promise<T | null> {
	try {
		const response = await fetch(path);

		return response.ok ? ((await response.json()) as T) : null;
	} catch {
		return null;
	}
}
