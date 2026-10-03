/**
 * The base language as shipped, merged from its groups. Real text comes from the database with
 * the page; this copy is only the fallback when that request fails. vite.config.ts allows
 * reading the directory in development.
 */
import { flatten, type Messages } from './flatten';

export type { Messages } from './flatten';

/** The language every other is a translation of: the keys it has are the keys
    there are, and its text is what a missing translation falls back to. */
export const BASE = 'en';

const groupModules = import.meta.glob('../../../../../i18n/id/en/*.json', {
	eager: true,
	import: 'default'
});

/** The base language's text, as this build shipped it. */
export const base: Messages = Object.assign(
	{},
	...Object.values(groupModules).map((file) => flatten(file) ?? {})
);
