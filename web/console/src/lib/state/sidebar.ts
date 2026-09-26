import { COOKIES } from '$lib/brand';

/** How wide the dashboard sidebar is: the full column, or icons only.
 *
 *  The choice lives in a cookie rather than localStorage so the server can
 *  read it and render the panel already folded, the same reason the theme is
 *  kept in one. The panel layout, (panel)/+layout.server.ts, reads this name. */
export type SidebarState = 'wide' | 'mini';

const ONE_YEAR = 60 * 60 * 24 * 365;

export function isSidebarState(value: unknown): value is SidebarState {
	return value === 'wide' || value === 'mini';
}

/** Writes the choice down, so the next page arrives the same way. */
export function rememberSidebar(state: SidebarState) {
	document.cookie = `${COOKIES.sidebar}=${state}; path=/; max-age=${ONE_YEAR}; samesite=lax`;
}

/** Which branches of the sidebar the reader has folded away.
 *
 *  The closed ones are written down rather than the open ones, so a section
 *  added in a later release arrives open: a reader who has never touched the
 *  column sees everything, and one who folded three sections still has them
 *  folded. It is a cookie for the same reason the width is — the server reads
 *  it and the first frame is already right.
 */
export function parseClosedBranches(raw: string | undefined): string[] {
	if (!raw) return [];

	return raw
		.split(',')
		.map((id) => id.trim())
		.filter((id) => /^[a-z-]+$/.test(id));
}

export function rememberClosedBranches(closed: string[]) {
	const value = closed.join(',');

	document.cookie = `${COOKIES.branches}=${encodeURIComponent(value)}; path=/; max-age=${ONE_YEAR}; samesite=lax`;
}
