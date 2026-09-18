import { getContext, setContext } from 'svelte';
import { BASE, base, type Messages } from './messages';

export { BASE, type Messages } from './messages';

/** Looks one message up. Anything in `{braces}` in the text is replaced by
    the parameter of that name. */
export type Translate = (key: string, params?: Record<string, string | number>) => string;

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
	return (key, params) => fill(messages()[key] || base[key] || key, params);
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
 * Which language to show somebody, given what they have asked for and what is
 * offered.
 *
 * In order: the language they chose, then the ones their browser asks for in
 * Accept-Language — by exact tag first, then by its language part, so a
 * browser asking for ru-RU is served ru — and then the installation's
 * default. Only what is offered counts wherever it appears: a saved choice of
 * a language that has since been turned off or removed is passed over, not
 * honoured into an error.
 */
export function chooseLanguage(
	chosen: string | undefined,
	accepted: string | null,
	offered: string[],
	fallback: string
): string {
	const usable = (code: string | undefined | null) =>
		code && offered.includes(code) ? code : null;

	const wanted = usable(chosen);
	if (wanted) return wanted;

	for (const tag of parseAccepted(accepted)) {
		const exact = usable(tag);
		if (exact) return exact;

		const language = usable(tag.split('-')[0]);
		if (language) return language;
	}

	return usable(fallback) ?? offered[0] ?? BASE;
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
