/**
 * What the project calls itself.
 *
 * The name is written here and nowhere else, and the names built from it — the
 * cookies, which have to match the ones the server sets — are built here too,
 * so the two sides cannot drift apart. The server keeps the same list in
 * internal/brand, and its tests read this file to check the two still agree.
 *
 * So renaming the project is this file, its twin in web/console, the module
 * path on the first line of go.mod, and a `git mv` of the command directory
 * under cmd/. Nothing else may spell the name.
 *
 * The product's own name comes from here rather than from a translation: a
 * product is not called something else in another language.
 */
export const BRAND = {
	/** The project as a person reads it: what the footer's shield names, and
	    what an organization a fresh installation starts with is called. */
	name: 'Loginer',

	/** The same name as an identifier — what the cookies and the environment
	    variables are built from. Lower case, because a cookie is not prose. */
	slug: 'loginer'
} as const;

/** The names this app and the server have to agree on. */
export const COOKIES = {
	/** The user's session at the server. Set by the API, HttpOnly: this app
	    never reads it, it only passes it on from a server load. */
	session: `${BRAND.slug}_user_session`,

	/** Light or dark, read while rendering so the page arrives themed. Its own
	    name, so the users' choice and the admin panel's do not overwrite each
	    other when both run on localhost. */
	theme: `${BRAND.slug}-account-theme`,

	/** The language chosen in the picker, read while rendering so the page
	    arrives in it. Its own name, for the same reason as the theme's. */
	language: `${BRAND.slug}-account-language`
} as const;

/** What the root layout's load depends on, so switching language re-runs it —
    and nothing else — through invalidate(). */
export const LANGUAGE_DEPENDENCY = `${BRAND.slug}:language`;
