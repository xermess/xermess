/**
 * Where to go after signing in, from a `?next=` parameter.
 *
 * Only a path on this site is accepted: `next=https://evil.example` or
 * `next=//evil.example` would turn the sign-in page into a way to send someone
 * anywhere with this site's name on the link.
 *
 * Looking at the first characters is not enough. A browser strips tabs and
 * newlines out of an address before it looks at it, so `/\t/evil.example` is
 * `//evil.example` to the browser and a path to a prefix check; and a
 * backslash is a slash. So the value is parsed the way a browser parses it,
 * against a stand-in origin, and accepted only if it stayed there — and
 * what is returned is the parsed path, not the text that came in.
 */
export function safeNext(next: string | null, fallback = '/'): string {
	if (!next || !next.startsWith('/') || /[\u0000-\u001f\u007f]/.test(next)) {
		return fallback;
	}

	const base = 'https://next.invalid';
	let parsed: URL;
	try {
		parsed = new URL(next, base);
	} catch {
		return fallback;
	}

	if (parsed.origin !== base) {
		return fallback;
	}

	return `${parsed.pathname}${parsed.search}${parsed.hash}`;
}
