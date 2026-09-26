import { invalidate } from '$app/navigation';
import { tick } from 'svelte';

import { COOKIES, LANGUAGE_DEPENDENCY } from '$lib/brand';

const ONE_YEAR = 60 * 60 * 24 * 365;

/**
 * Remembers a language and redraws the page in it, where the reader is.
 *
 * Nothing is reloaded: the root layout asks the server for the new text and
 * every component re-renders with it, so a half-typed email address and the
 * scroll position both survive. The choice is a cookie rather than
 * localStorage because the server reads it to send the next page already
 * translated.
 *
 * Where the browser can, the two languages cross-fade rather than one
 * snapping into the other — unless the reader has asked for less motion.
 */
export async function switchLanguage(code: string) {
	document.cookie = `${COOKIES.language}=${encodeURIComponent(code)}; path=/; max-age=${ONE_YEAR}; samesite=lax`;

	const redraw = async () => {
		await invalidate(LANGUAGE_DEPENDENCY);
		await tick();
	};

	const still = matchMedia('(prefers-reduced-motion: reduce)').matches;
	if (!document.startViewTransition || still || document.hidden) {
		await redraw();
		return;
	}

	// A transition the browser skips — the tab went to the background, say —
	// still redraws; only the fade is lost, which is nothing to report.
	const transition = document.startViewTransition(redraw);
	transition.ready.catch(() => {});
	await transition.updateCallbackDone;
}
