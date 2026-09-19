import { getContext, setContext } from 'svelte';
import { ApiError } from '$lib/api/client';
import { BASE, base, type Messages } from './messages';

export { BASE, type Messages } from './messages';

/** Looks one message up. Anything in `{braces}` in the text is replaced by
    the parameter of that name. `has` says whether there is text for a key at
    all, rather than the key itself coming back. */
export type Translate = ((key: string, params?: Record<string, string | number>) => string) & {
	has: (key: string) => boolean;
};

/**
 * Builds a translator over whatever text `messages` returns.
 *
 * It reads `messages()` on every lookup rather than once, so a component that
 * calls `t(...)` while it renders re-renders when the language changes — the
 * same way it would for any other state it read.
 *
 * The server sends the text with every key already filled in, so the
 * fallbacks here are for when it could not be asked: the base language as
 * this build shipped it, and then the key itself. Neither is worth an error —
 * a half-translated page is usable, and a page that throws is not.
 */
export function translator(messages: () => Messages): Translate {
	const lookup = (key: string) => messages()[key] || base[key];

	return Object.assign(
		(key: string, params?: Record<string, string | number>) => fill(lookup(key) || key, params),
		{ has: (key: string) => Boolean(lookup(key)) }
	);
}

/**
 * What to tell an administrator about an error, in their language.
 *
 * The server names every problem with a code, and the panel says
 * `error.<code>` from its own catalog. A code the panel has no sentence for —
 * one of the admin API's not yet given one — falls back to the server's
 * English, and anything that is not an answer at all to `fallback`, or
 * "something went wrong".
 */
export function messageOf(err: unknown, t: Translate, fallback?: string): string {
	if (!(err instanceof ApiError)) return fallback ?? t('error.unknown');

	const key = `error.${err.code}`;
	if (t.has(key)) return t(key, err.params);

	return err.message || fallback || t('error.unknown');
}

/** Replaces `{name}` with the parameter of that name. A parameter nobody
    passed is left as it was written, which shows up in the page rather than
    silently emptying the sentence. */
function fill(text: string, params?: Record<string, string | number>): string {
	if (!params) return text;

	return text.replace(/\{(\w+)\}/g, (whole, name: string) =>
		name in params ? String(params[name]) : whole
	);
}

const KEY = Symbol('i18n');

/** Put the translator where every component below can reach it. The root
    layout does this once; nothing else should. */
export function provideTranslator(translate: Translate) {
	setContext(KEY, translate);
}

/**
 * The translator, for a component that has text in it:
 *
 * ```svelte
 * const t = useTranslator();
 * ...
 * <h1>{t('login.title')}</h1>
 * ```
 *
 * It has to be called while the component is being created, as every context
 * does. A component rendered outside the layout — a test, a story — gets the
 * base language rather than nothing.
 */
export function useTranslator(): Translate {
	return getContext<Translate | undefined>(KEY) ?? translator(() => base);
}

/**
 * Which language to show an administrator.
 *
 * In order: the one they chose, then the ones their browser asks for in
 * Accept-Language — by exact tag first, then by its language part, so a
 * browser asking for ru-RU is served ru — and then the base language.
 *
 * `available` is the languages the panel is shown in — English and Russian,
 * as the server lists them — not what the Languages page offers users: which
 * of those an administrator reads the panel in is their own business.
 */
export function chooseLanguage(
	chosen: string | undefined,
	accepted: string | null,
	available: string[]
): string {
	const has = (code: string) => available.includes(code);

	if (chosen && has(chosen)) return chosen;

	for (const tag of parseAccepted(accepted)) {
		if (has(tag)) return tag;

		const language = tag.split('-')[0];
		if (has(language)) return language;
	}

	return BASE;
}

/** The tags of an Accept-Language header, best first. */
function parseAccepted(header: string | null): string[] {
	if (!header) return [];

	return header
		.split(',')
		.map((part) => {
			const [tag, ...rest] = part.trim().split(';');
			const quality = rest
				.map((piece) => piece.trim())
				.find((piece) => piece.startsWith('q='))
				?.slice(2);

			return { tag: tag.trim(), quality: quality === undefined ? 1 : Number(quality) || 0 };
		})
		.filter((entry) => entry.tag !== '' && entry.tag !== '*' && entry.quality > 0)
		.sort((a, b) => b.quality - a.quality)
		.map((entry) => entry.tag);
}
