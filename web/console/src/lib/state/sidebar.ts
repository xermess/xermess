import { COOKIES } from '$lib/brand';

/**
 * Sidebar width, kept in a cookie so the server renders it already folded. Read by
 * (panel)/+layout.server.ts.
 */
export type SidebarState = 'wide' | 'mini';

const ONE_YEAR = 60 * 60 * 24 * 365;

export function isSidebarState(value: unknown): value is SidebarState {
	return value === 'wide' || value === 'mini';
}

/** Writes the choice down, so the next page arrives the same way. */
export function rememberSidebar(state: SidebarState) {
	document.cookie = `${COOKIES.sidebar}=${state}; path=/; max-age=${ONE_YEAR}; samesite=lax`;
}

/**
 * The sidebar groups the reader folded. Closed ones are stored, so new sections arrive open. A
 * cookie, so the first frame is right.
 */
export function parseClosedGroups(raw: string | undefined): string[] {
	if (!raw) return [];

	return raw
		.split(',')
		.map((id) => id.trim())
		.filter((id) => /^[a-z-]+$/.test(id));
}

export function rememberClosedGroups(closed: string[]) {
	const value = closed.join(',');

	document.cookie = `${COOKIES.branches}=${encodeURIComponent(value)}; path=/; max-age=${ONE_YEAR}; samesite=lax`;
}
