import type { Handle } from '@sveltejs/kit';

import { COOKIES } from '$lib/brand';

export { handleFetch } from '$lib/server/proxy';

/**
 * Writes the reader's chosen theme into <html> before it is sent, so there is no
 * light-then-dark flash; with no choice, prefers-color-scheme in tokens.css decides. <html
 * lang> is always English: the panel is not translated.
 */
export const handle: Handle = async ({ event, resolve }) => {
	const saved = event.cookies.get(COOKIES.theme);
	const theme = saved === 'dark' || saved === 'light' ? saved : null;
	const attribute = theme ? `data-theme="${theme}"` : '';

	const response = await resolve(event, {
		transformPageChunk: ({ html }) => html.replace('__THEME__', attribute).replace('__LANG__', 'en')
	});

	// The panel takes an administrator's password and acts on their behalf,
	// so no other site may frame it: a transparent frame over a fake page is
	// how clickjacking gets a click or a password it should not.
	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set('Content-Security-Policy', "frame-ancestors 'none'");

	return response;
};
