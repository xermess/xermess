import { invalidate } from '$app/navigation';
import { tick } from 'svelte';

import { COOKIES, LANGUAGE_DEPENDENCY } from '$lib/brand';

const ONE_YEAR = 60 * 60 * 24 * 365;

/**
 * Switches language in place: stores the choice in a cookie the server reads, then reloads only
 * the text, so typed input and scroll position survive. Cross-fades where supported, unless
 * reduced motion is requested.
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
