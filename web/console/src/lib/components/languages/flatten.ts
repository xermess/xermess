/** One language's messages, by dotted key: "login.title". */
export type Messages = Record<string, string>;

/** A translation group as it is written: nested by namespace, text at the leaves. */
export type Nested = { [key: string]: string | Nested };

/**
 * Flattens a nested translation group into dotted keys (`{"login": {"title": …}}` to
 * "login.title"). Flat or mixed groups read the same; anything that is not text, or a repeated
 * key, gives null.
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

/** Writes messages back the way the groups are: nested by namespace, keys in
    the order given. Keys starting with `$` describe the group and stay on top. */
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
