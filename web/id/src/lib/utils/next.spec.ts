import { describe, expect, it } from 'vitest';
import { safeNext } from './next';

describe('safeNext', () => {
	it('keeps a path on this site', () => {
		expect(safeNext('/security?tab=sessions')).toBe('/security?tab=sessions');
		expect(safeNext('/security#sessions')).toBe('/security#sessions');
	});

	it('refuses anywhere else', () => {
		for (const next of [
			'https://evil.example',
			'//evil.example',
			'/\\evil.example',
			'security',
			''
		]) {
			expect(safeNext(next)).toBe('/');
		}
		expect(safeNext(null, '/applications')).toBe('/applications');
	});

	it('refuses what a browser would read as another host', () => {
		// A browser drops tabs and newlines before parsing, and reads a
		// backslash as a slash: each of these is //host to it.
		for (const next of ['/\t/evil.example', '/\n/evil.example', '/\r/evil.example', '/\\/evil.example', '/%09/x']) {
			expect(new URL(next.replace(/%09/g, '\t'), 'https://id.example.com/login').origin).not.toBe(
				'https://id.example.com'
			);
			expect(safeNext(next.replace(/%09/g, '\t'))).toBe('/');
		}
	});
});
