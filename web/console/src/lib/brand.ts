/**
 * What the project calls itself, and the cookie names built from it. The server keeps the same
 * list in internal/brand and its tests check the two agree; nothing else may spell the name.
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

/** Cookie names shared by the code that writes and the code that reads each cookie. */
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
