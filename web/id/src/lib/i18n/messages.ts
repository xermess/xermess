/**
 * The one language this app is built with: the base, as the server ships it.
 *
 * Every other language — and the base language's own current wording — is in
 * the database and arrives with the page (the root layout asks for it), so a
 * language added or reworded in the admin panel needs no build. This copy is
 * only what is left when that request fails: the sign-in pages have to draw
 * themselves in something when nothing else works.
 *
 * `vite.config.ts` allows the directory to be read in development.
 */
import english from '../../../../../locales/id/en.json';

/** One language's messages, by key. */
export type Messages = Record<string, string>;

/** The language every other is a translation of: the keys it has are the keys
    there are, and its text is what a missing translation falls back to. */
export const BASE = 'en';

/** The base language's text, as this build shipped it. */
export const base: Messages = english;
