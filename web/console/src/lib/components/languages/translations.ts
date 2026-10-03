/**
 * What the Languages page does with translation files and language tags,
 * apart from drawing them.
 */
import { flatten } from './flatten';

/** The `{name}` parameters a text has, which the app fills in when it shows
    it — `{app}`, `{count}`. */
function parameters(text: string): Set<string> {
	return new Set([...text.matchAll(/\{(\w+)\}/g)].map((match) => `{${match[1]}}`));
}

/** The parameters the English has and a translation lost. A translation
    without `{app}` still reads, but never says which application — which is
    easy to miss, and worth saying beside the text. */
export function missingParameters(english: string, translation: string): string[] {
	const kept = parameters(translation);

	return [...parameters(english)].filter((name) => !kept.has(name));
}

export type ReadResult =
	{ ok: true; messages: Record<string, string>; skipped: number } | { ok: false };

/**
 * Reads an imported translation file, nested or flat. Unknown keys are skipped and counted,
 * `$name`/`$native` are ignored, and empty texts are left out so a partial file never erases
 * anything.
 */
export async function readTranslationFile(file: File, known: string[]): Promise<ReadResult> {
	let parsed: unknown;
	try {
		parsed = JSON.parse(await file.text());
	} catch {
		return { ok: false };
	}

	const flat = flatten(parsed);
	if (!flat) return { ok: false };

	const keys = new Set(known);
	const messages: Record<string, string> = {};
	let skipped = 0;

	for (const [key, value] of Object.entries(flat)) {
		if (key.startsWith('$')) continue;

		if (!keys.has(key)) {
			skipped++;
		} else if (value.trim() !== '') {
			messages[key] = value;
		}
	}

	return { ok: true, messages, skipped };
}

/** Hands the browser a file to save. */
export function download(filename: string, text: string) {
	const url = URL.createObjectURL(new Blob([text], { type: 'application/json' }));

	const link = document.createElement('a');
	link.href = url;
	link.download = filename;
	link.click();

	URL.revokeObjectURL(url);
}

/** A language tag as the server accepts one: a two- or three-letter
    language, then optionally a script or a region. */
export const languageCode = /^[a-z]{2,3}(-[A-Za-z0-9]{2,8})*$/i;

/**
 * A language's English and native names from its tag, using the browser's Intl data; null when
 * the browser cannot name it.
 */
export function namesOf(code: string): { name: string; native: string } | null {
	try {
		const [tag] = Intl.getCanonicalLocales(code);
		const name = new Intl.DisplayNames(['en'], { type: 'language', fallback: 'none' }).of(tag);
		const native = new Intl.DisplayNames([tag], { type: 'language', fallback: 'none' }).of(tag);

		if (!name) return null;

		return { name, native: native ? capitalise(native, tag) : name };
	} catch {
		return null;
	}
}

/** Many languages write their own name in lower case — "русский",
    "español" — which reads as a mistake at the start of a line in a picker. */
function capitalise(text: string, tag: string): string {
	return text.charAt(0).toLocaleUpperCase(tag) + text.slice(1);
}
