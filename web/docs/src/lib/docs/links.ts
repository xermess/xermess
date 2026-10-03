import { resolve } from '$app/paths';

/** Prefixes a docs path with the base path; the one place a stored address becomes a link. */
export function href(path: string): string {
	return resolve(path as `/${string}`);
}
