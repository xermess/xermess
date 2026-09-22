import type { Handle } from '@sveltejs/kit';

import { COOKIES } from '$lib/constants';
import { BASE } from '$lib/i18n';

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
 * The language goes in the same way and for the same reason, and `<html lang>`
 * is what a screen reader reads the page's words with. The root layout's load
 * decides it, from the languages the API has, and leaves it in `locals`; the
 * page is only transformed once the loads have run.
 */
export const handle: Handle = async ({ event, resolve }) => {
	const saved = event.cookies.get(COOKIES.theme);
	const theme = saved === 'dark' || saved === 'light' ? saved : null;

	const attributes = theme ? `data-theme="${theme}"` : '';

	const response = await resolve(event, {
		transformPageChunk: ({ html }) =>
			html.replace('__THEME__', attributes).replace('__LANG__', event.locals.language ?? BASE)
	});

	// The panel takes an administrator's password and acts on their behalf,
	// so no other site may frame it: a transparent frame over a fake page is
	// how clickjacking gets a click or a password it should not.
	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set('Content-Security-Policy', "frame-ancestors 'none'");

	return response;
};
