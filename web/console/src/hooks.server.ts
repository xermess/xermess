import type { Handle } from '@sveltejs/kit';

import { COOKIES } from '$lib/constants';
import { DEFAULT_FONT, isFont } from '$lib/state/font.svelte';

export { handleFetch } from '$lib/server/proxy';

/**
 * Puts the reader's theme into the HTML before it is sent.
 *
 * The panel renders in the browser, so for the first frames there is no
 * stylesheet and no JavaScript: whatever `<html>` carries is what gets
 * painted. Deciding it here means a reader who chose dark is served dark,
 * rather than being shown light and corrected a moment later.
 *
 * A reader who has not chosen yet gets no attribute at all, which leaves the
 * prefers-color-scheme rules in tokens.css to decide — also without a flash,
 * because that happens in CSS rather than after it.
 *
 * `<html lang>` is written here too — it is what a screen reader reads the
 * page's words with — and it is always English: the panel is written in
 * English, in its own markup. The sign-in pages are the app an installation
 * translates.
 *
 * The typeface chosen in the header's picker is written beside the theme,
 * for the same reason: text painted in one face and redrawn in another is a
 * visible jump.
 */
export const handle: Handle = async ({ event, resolve }) => {
	const saved = event.cookies.get(COOKIES.theme);
	const theme = saved === 'dark' || saved === 'light' ? saved : null;

	const chosen = event.cookies.get(COOKIES.font);
	const font = isFont(chosen) && chosen !== DEFAULT_FONT ? chosen : null;

	const attributes = [theme && `data-theme="${theme}"`, font && `data-font="${font}"`]
		.filter(Boolean)
		.join(' ');

	const response = await resolve(event, {
		transformPageChunk: ({ html }) =>
			html.replace('__THEME__', attributes).replace('__LANG__', 'en')
	});

	// The panel takes an administrator's password and acts on their behalf,
	// so no other site may frame it: a transparent frame over a fake page is
	// how clickjacking gets a click or a password it should not.
	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set('Content-Security-Policy', "frame-ancestors 'none'");

	return response;
};
