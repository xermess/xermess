/**
 * The names both sides of the app have to agree on.
 *
 * A cookie is read in one place and written in another — the theme by the
 * hook that renders the page and by the toggle that changes it, the session
 * by the server loads — so the names live here rather than as a string in
 * each file that happens to spell them the same way.
 */
export const COOKIES = {
	/** The API's session. Set by the API itself, HttpOnly: the panel never
	    reads it, it only passes it on from a server load. */
	session: 'xermess_session',

	/** Light or dark, read while rendering so the page arrives themed. */
	theme: 'xermess-theme',

	/** Whether the dashboard sidebar is folded, read the same way. */
	sidebar: 'xermess-sidebar',

	/** The language this panel is shown in, read while rendering so the page
	    arrives in it. It is one administrator's own preference on one
	    machine, not a setting of the installation. */
	language: 'xermess-language'
} as const;

/** What the root layout's load depends on, so switching language re-runs it —
    and nothing else — through invalidate(). */
export const LANGUAGE_DEPENDENCY = 'xermess:language';
