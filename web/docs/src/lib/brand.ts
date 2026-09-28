/**
 * What the project calls itself.
 *
 * The name is written here and nowhere else in this app, and the names built
 * from it — the cookies the examples send — are built here too. The server
 * keeps the same list in internal/brand, and its tests read this file to check
 * the two still agree.
 *
 * The guides never spell the name either: they write {{name}}, {{slug}} and
 * the other keys of PLACEHOLDERS below, and the page fills them in as it
 * renders. So renaming the project is this file, its twins in web/console and
 * web/id, the module path in go.mod, and a `git mv` of the command under cmd/.
 */
export const BRAND = {
	/** The project as a person reads it: the wordmark and the page titles. */
	name: 'Loginer',

	/** The same name as an identifier, as cookies and variables are built from it. */
	slug: 'loginer',

	/** Where these docs are published, and where the source lives. */
	docsURL: 'https://loginer.org/docs',
	githubURL: 'https://github.com/loginer/loginer'
} as const;

/** The cookies the examples send. The server sets both. */
export const COOKIES = {
	/** A user's session at the provider, from the sign-in pages. */
	userSession: `${BRAND.slug}_user_session`,

	/** An administrator's session, from the panel. */
	adminSession: `${BRAND.slug}_session`
} as const;

/** Light or dark, remembered in the reader's browser. */
export const THEME_KEY = `${BRAND.slug}-docs-theme`;

/** The language the reader last chose for the examples: Java, Go or Python. */
export const LANGUAGE_KEY = `${BRAND.slug}-docs-language`;

/** What {{key}} in a page stands for. */
export const PLACEHOLDERS: Record<string, string> = {
	name: BRAND.name,
	slug: BRAND.slug,
	envPrefix: `${BRAND.slug.toUpperCase()}_`,
	userSessionCookie: COOKIES.userSession,
	adminSessionCookie: COOKIES.adminSession,
	githubURL: BRAND.githubURL,
	/** The audiences of the server's own two APIs, as internal/brand builds them. */
	adminAPI: `urn:${BRAND.slug}:admin-api`,
	accountAPI: `urn:${BRAND.slug}:account-api`
};
