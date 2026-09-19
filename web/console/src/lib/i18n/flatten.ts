/** One language's messages, by dotted key: "login.title". */
export type Messages = Record<string, string>;

/** A translation file as it is written: nested by screen, text at the leaves. */
export type Nested = { [key: string]: string | Nested };

/**
 * Reads a translation file into messages by dotted key.
 *
 * The files under locales/ are nested by screen — `{"login": {"title": …}}` —
 * so a translator sees one page's text together; everything that looks text
 * up speaks of "login.title". A flat file, or one that mixes the two, reads
 * the same. It answers null for anything that is not a file of text: a
 * number, a list, or a key said twice.
 */
export function flatten(file: unknown): Messages | null {
	if (!isObject(file)) return null;

	const out: Messages = {};

	const walk = (node: Record<string, unknown>, prefix: string): boolean => {
		for (const [key, value] of Object.entries(node)) {
			const full = prefix ? `${prefix}.${key}` : key;

			if (typeof value === 'string') {
				if (full in out) return false;
				out[full] = value;
			} else if (isObject(value)) {
				if (!walk(value, full)) return false;
			} else {
				return false;
			}
		}

		return true;
	};

	return walk(file, '') ? out : null;
}

/** Writes messages back the way the files are: nested by screen, keys in the
    order given. Keys starting with `$` describe the file and stay on top. */
export function nest(messages: Messages): Nested {
	const out: Nested = {};

	for (const [key, value] of Object.entries(messages)) {
		if (key.startsWith('$')) {
			out[key] = value;
			continue;
		}

		const parts = key.split('.');
		let node = out;

		for (const part of parts.slice(0, -1)) {
			const next = node[part];
			if (typeof next === 'string') break;
			node = (node[part] ??= {}) as Nested;
		}

		node[parts[parts.length - 1]] = value;
	}

	return out;
}

function isObject(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null && !Array.isArray(value);
}
