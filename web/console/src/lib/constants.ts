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

	/** Which typeface the panel is set in, read the same way. */
	font: 'xermess-font',

	/** Whether the dashboard sidebar is folded, read the same way. */
	sidebar: 'xermess-sidebar',

	/** Which of the sidebar's sections the reader has folded away, read the
	    same way so the column arrives as they left it. */
	branches: 'xermess-sidebar-closed'
} as const;

/** Where the project itself lives, rather than this installation of it: the
    documentation, and the source. The header links to both and so does the
    account menu, so the addresses are written once. */
export const DOCS_URL = 'https://xermess.org/docs';
export const GITHUB_URL = 'https://github.com/xermess/xermess';

/** The shortest password an administrator may have: the server's
    model.MinAdminPasswordLength, said here too so a form can say so before
    anything is sent. */
export const MIN_ADMIN_PASSWORD = 10;

/** What a load that reads the signed-in administrator depends on, so a
    change to their own account can refresh that alone rather than every
    load on the page. */
export const ADMIN_DEPENDENCY = 'app:admin';
