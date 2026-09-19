/**
 * The one language this panel is built with: the base, as the server ships it.
 *
 * Every other language — and the base language's own current wording — is in
 * the database and arrives with the page (the root layout asks for it), so a
 * language added or reworded on the Languages page needs no build. This copy
 * is only what is left when that request fails, so the panel still draws
 * itself when the API is down.
 *
 * The file is nested by screen, as every file under locales/ is, and read
 * here into dotted keys. `vite.config.ts` allows the directory to be read in
 * development.
 */
import english from '../../../../../locales/console/en.json';
import { flatten, type Messages } from './flatten';

export type { Messages } from './flatten';

/** The language every other is a translation of: the keys it has are the keys
    there are, and its text is what a missing translation falls back to. */
export const BASE = 'en';

/** The base language's text, as this build shipped it. */
export const base: Messages = flatten(english) ?? {};
