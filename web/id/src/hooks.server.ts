import type { Handle } from '@sveltejs/kit';

import { COOKIES } from '$lib/constants';
import { BASE } from '$lib/i18n';

export { handleFetch } from '$lib/server/proxy';

/**
 * Two things every response of this app gets.
 *
 * The reader's theme goes into the HTML before it is sent, so a reader who
 * chose dark is served dark rather than shown light and corrected a moment
 * later. A reader who has not chosen gets no attribute, and the
 * prefers-color-scheme rules in tokens.css decide — also before the first
 * paint.
 *
 * The language goes in the same way, so `<html lang>` is right in the first
 * byte — which is what a screen reader reads the page's words with. Which
 * language that is depends on what this installation offers, so the root
 * layout's load decides it and leaves it in `locals`; the page is only
 * transformed once the loads have run.
 *
 * And no other site may frame any page: every one of them takes a password or
 * acts for a signed-in user, and a transparent frame over a fake page is how
 * clickjacking gets either.
 */
export const handle: Handle = async ({ event, resolve }) => {
	const saved = event.cookies.get(COOKIES.theme);
	const theme = saved === 'dark' || saved === 'light' ? saved : null;

	const attributes = theme ? `data-theme="${theme}"` : '';

	const response = await resolve(event, {
		transformPageChunk: ({ html }) =>
			html.replace('__THEME__', attributes).replace('__LANG__', event.locals.language ?? BASE)
	});

	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set('Content-Security-Policy', "frame-ancestors 'none'");
	response.headers.set('Referrer-Policy', 'no-referrer');
	response.headers.set('X-Content-Type-Options', 'nosniff');

	return response;
};
