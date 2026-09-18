import { redirect } from '@sveltejs/kit';
import { signIn, type PublicLanguage } from '$lib/api';
import { COOKIES, LANGUAGE_DEPENDENCY } from '$lib/constants';
import { BASE, chooseLanguage, type Messages } from '$lib/i18n';
import type { LayoutServerLoad } from './$types';

/**
 * The language every page below is drawn in, its text, and the ones somebody
 * may switch to.
 *
 * It is resolved here rather than in each page because it is the same answer
 * for all of them, and on the server because the first response has to arrive
 * already translated — a page that renders in English and corrects itself is
 * a page that flickers.
 *
 * The text comes from the server rather than the build: languages are added
 * and reworded in the admin panel, and the next page anybody opens has it.
 * The picker switches by setting a cookie and invalidating LANGUAGE_DEPENDENCY, which
 * runs this load again and nothing else. `?lang=` does the same for a link
 * followed without JavaScript: it is remembered and then redirected away, so
 * an address left carrying one reader's choice does not carry it to whoever
 * it was shared with.
 *
 * Failing to load any of it is not worth an error page. The pages fall back
 * to the base language this app was built with — the sign-in pages have to
 * work when nothing else does.
 */
export const load: LayoutServerLoad = async ({ cookies, depends, fetch, locals, request, url }) => {
	depends(LANGUAGE_DEPENDENCY);

	const asked = url.searchParams.get('lang') ?? undefined;

	const offered = await signIn
		.languages(fetch)
		.catch(() => ({ languages: [] as PublicLanguage[], default: BASE }));

	const language = chooseLanguage(
		asked ?? cookies.get(COOKIES.language),
		request.headers.get('accept-language'),
		offered.languages.map((one) => one.code),
		offered.default
	);

	if (asked !== undefined) {
		cookies.set(COOKIES.language, language, {
			path: '/',
			// The page reads it while rendering on the server, and the picker
			// writes it from the browser; it is a preference, not a secret.
			httpOnly: false,
			sameSite: 'lax',
			maxAge: 60 * 60 * 24 * 365
		});

		const clean = new URL(url);
		clean.searchParams.delete('lang');

		redirect(303, `${clean.pathname}${clean.search}`);
	}

	const messages: Messages = offered.languages.length
		? await signIn
				.languageText(language, fetch)
				.then((text) => text.messages)
				.catch(() => ({}))
		: {};

	// hooks.server.ts writes this into <html lang> once the page is rendered,
	// so the first byte names the language the page is actually in.
	locals.language = language;

	return { language, languages: offered.languages, messages };
};
