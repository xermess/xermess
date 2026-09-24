import { browser } from '$app/environment';

import { COOKIES } from '$lib/constants';

/** The typefaces the panel can be set in while one is being chosen. Each
 *  names the text face and the code face that goes with it. */
export const FONTS = [
	{ id: 'roboto', label: 'Roboto', detail: 'with Roboto Mono' },
	{ id: 'product-sans', label: 'Product Sans', detail: 'with Roboto Mono' },
	{ id: 'plex', label: 'IBM Plex Sans', detail: 'with IBM Plex Mono' }
] as const;

export type Font = (typeof FONTS)[number]['id'];

/** Roboto is what tokens.css sets without an attribute, so choosing it
 *  removes the attribute rather than naming the default twice. */
export const DEFAULT_FONT: Font = 'roboto';

const COOKIE = COOKIES.font;
const ATTRIBUTE = 'data-font';
const ONE_YEAR = 60 * 60 * 24 * 365;

export function isFont(value: unknown): value is Font {
	return FONTS.some((font) => font.id === value);
}

/** Kept in a cookie for the same reason as the theme: hooks.server.ts reads
 *  it and writes the attribute, so the page is painted in the chosen face
 *  from its first frame. This reads back what the server wrote. */
function initial(): Font {
	if (!browser) return DEFAULT_FONT;

	const attribute = document.documentElement.getAttribute(ATTRIBUTE);

	return isFont(attribute) ? attribute : DEFAULT_FONT;
}

class FontState {
	current = $state<Font>(initial());

	set(font: Font) {
		this.current = font;
		if (!browser) return;

		if (font === DEFAULT_FONT) {
			document.documentElement.removeAttribute(ATTRIBUTE);
		} else {
			document.documentElement.setAttribute(ATTRIBUTE, font);
		}

		document.cookie = `${COOKIE}=${font}; path=/; max-age=${ONE_YEAR}; samesite=lax`;
	}
}

export const font = new FontState();
