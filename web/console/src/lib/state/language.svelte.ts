import { invalidate } from '$app/navigation';

import { COOKIES, LANGUAGE_DEPENDENCY } from '$lib/constants';

/** One language the panel can be shown in, as the picker lists it. */
export type LanguageChoice = {
	code: string;
	/** What the language calls itself, which is what somebody looking for
	    their own language scans the list for. */
	native: string;
	/** And what it is called in English, for whoever is setting it up. */
	name: string;
};

const ONE_YEAR = 60 * 60 * 24 * 365;

/**
 * Remembers a language and re-renders the panel in it.
 *
 * The choice is a cookie rather than localStorage so the server can read it
 * and send the page already translated; invalidating afterwards is what makes
 * the current page follow, since the text it draws with comes from the root
 * layout's load.
 */
export async function setLanguage(code: string) {
	document.cookie = `${COOKIES.language}=${encodeURIComponent(code)}; path=/; max-age=${ONE_YEAR}; samesite=lax`;

	await reloadText();
}

/** Asks the server for the panel's text again: after a language is switched,
    or after its text was edited on the Languages page. */
export function reloadText() {
	return invalidate(LANGUAGE_DEPENDENCY);
}
