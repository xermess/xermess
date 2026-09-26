/**
 * What the project calls itself.
 *
 * The name is written here and nowhere else, and the names built from it — the
 * cookies, which have to match the ones the server sets — are built here too,
 * so the two sides cannot drift apart. The server keeps the same list in
 * internal/brand, and its tests read this file to check the two still agree.
 *
 * So renaming the project is this file, its twin in web/id, the module
 * path on the first line of go.mod, and a `git mv` of the command directory
 * under cmd/. Nothing else may spell the name.
 */
export const BRAND = {
	/** The project as a person reads it: the wordmark beside the mark, and
	    the suffix every page title ends in. */
	name: 'Loginer',

	/** The same name as an identifier — what the cookies and the environment
	    variables are built from. Lower case, because a cookie is not prose. */
	slug: 'loginer',

	/** Where the project itself lives, rather than this installation of it:
	    the documentation, and the source. The header links to both. */
	docsUrl: 'https://loginer.org/docs',
	githubUrl: 'https://github.com/loginer/loginer'
} as const;

/** The names this app and the server have to agree on. A cookie is read in one
    place and written in another — the theme by the hook that renders the page
    and by the toggle that changes it, the session by the server loads — so the
    names live here rather than as a string in each file that happens to spell
    them the same way. */
export const COOKIES = {
	/** The API's session. Set by the API itself, HttpOnly: the panel never
	    reads it, it only passes it on from a server load. */
	session: `${BRAND.slug}_session`,

	/** Light or dark, read while rendering so the page arrives themed. */
	theme: `${BRAND.slug}-theme`,

	/** Whether the dashboard sidebar is folded, read the same way. */
	sidebar: `${BRAND.slug}-sidebar`,

	/** Which of the sidebar's sections the reader has folded away, read the
	    same way so the column arrives as they left it. */
	branches: `${BRAND.slug}-sidebar-closed`
} as const;
