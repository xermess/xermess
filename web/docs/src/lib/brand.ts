/**
 * What the project calls itself, and the names built from it. The server checks this matches
 * internal/brand. Guides write {{name}} and the other PLACEHOLDERS keys instead of the name.
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
