/**
 * The names the server and the browser have to agree on.
 */
export const COOKIES = {
	/** The user's session at xermess. Set by the API, HttpOnly: this app never
	    reads it, it only passes it on from a server load. */
	session: 'xermess_user_session',

	/** Light or dark, read while rendering so the page arrives themed. Its own
	    name, so the users' choice and the admin panel's do not overwrite each
	    other when both run on localhost. */
	theme: 'xermess-account-theme',

	/** The language chosen in the picker, read while rendering so the page
	    arrives in it. Its own name, for the same reason as the theme's. */
	language: 'xermess-account-language'
} as const;

/** What the root layout's load depends on, so switching language re-runs it —
    and nothing else — through invalidate(). */
export const LANGUAGE_DEPENDENCY = 'xermess:language';
