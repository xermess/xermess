/**
 * The post-sign-in destination from `?next=`, accepted only if it stays on this site (no open
 * redirect). Browsers strip tabs and newlines and treat \ as /, so the value is parsed like a
 * browser would against a stand-in origin, and the parsed path is returned.
 */
export function safeNext(next: string | null, fallback = '/'): string {
	if (!next || !next.startsWith('/') || hasControlCharacter(next)) {
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

function hasControlCharacter(value: string): boolean {
	for (let i = 0; i < value.length; i++) {
		const code = value.charCodeAt(i);
		if (code <= 0x1f || code === 0x7f) return true;
	}
	return false;
}
