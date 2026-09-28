import { resolve } from '$app/paths';

/**
 * An address inside the docs, with the base path in front. Every page is one
 * rest route, so an address is only known as a string — the navigation, the
 * search index and the Markdown all hold them that way — and this is the one
 * place that turns one into a link.
 */
export function href(path: string): string {
	return resolve(path as `/${string}`);
}
