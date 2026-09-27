import type { CacheDatabaseName, CacheKind } from '$lib/api';
import type { ColorPalette } from '$lib/components/ui';

/** What each database is called, and what it is for, as the page says it. */
export const databases: Record<CacheDatabaseName, { label: string; description: string }> = {
	cache: {
		label: 'Cache',
		description:
			'What every page reads and anybody may see: the languages and their text, the organisation, the login flows, the sign-in buttons. Clearing it only makes the next pages slower.'
	},
	sessions: {
		label: 'Sessions',
		description:
			'Who is signed in and what they may do: the sessions behind the cookies, the administrators with their roles, the applications by client id, and the rate limit’s counts. Nothing here is edited by hand; removing it signs nobody out.'
	}
};

/** What each kind of key is called, and the colour it is shown in. */
export const kinds: Record<CacheKind, { label: string; tone: ColorPalette }> = {
	entry: { label: 'Value', tone: 'success' },
	stale: { label: 'Stale', tone: 'neutral' },
	generation: { label: 'Generation', tone: 'info' },
	session: { label: 'Session', tone: 'warning' },
	ratelimit: { label: 'Rate limit', tone: 'danger' },
	other: { label: 'Other', tone: 'neutral' }
};

/** The kinds a listing can be narrowed to, in each database. */
export const kindFilters: Record<CacheDatabaseName, { value: CacheKind | ''; label: string }[]> = {
	cache: [
		{ value: '', label: 'Every kind' },
		{ value: 'entry', label: 'Values' },
		{ value: 'generation', label: 'Generations' }
	],
	sessions: [
		{ value: '', label: 'Every kind' },
		{ value: 'session', label: 'Sessions' },
		{ value: 'entry', label: 'Values' },
		{ value: 'ratelimit', label: 'Rate limits' },
		{ value: 'generation', label: 'Generations' }
	]
};

/** A size in bytes, the way a person reads it. */
export function formatBytes(bytes: number): string {
	if (bytes < 1024) return `${bytes} B`;
	if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
	return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

/** How long a key has left: "never" for one that does not expire. */
export function formatTTL(seconds: number): string {
	if (seconds < 0) return 'never';
	if (seconds < 60) return `${seconds}s`;
	if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
	return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
}
