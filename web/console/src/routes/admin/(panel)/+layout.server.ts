import type { Admin } from '$lib/api';
import { ADMIN_DEPENDENCY, COOKIES } from '$lib/constants';
import { apiGet } from '$lib/server/api';
import { isSidebarState, parseClosedBranches, type SidebarState } from '$lib/state/sidebar';
import type { LayoutServerLoad } from './$types';

/**
 * Everything in this group needs a session. Doing the check on the server
 * means a page is rendered already signed in, rather than appearing and then
 * being replaced once the browser has asked.
 *
 * How the reader left the sidebar is read here too, not in the dashboard: the
 * header's logo block is the top of the same column, so every page in the
 * group needs to know its width — and which of its sections are folded away —
 * and the first frame is already right.
 */
export const load: LayoutServerLoad = async ({ cookies, depends, fetch }) => {
	depends(ADMIN_DEPENDENCY);

	const { admin } = await apiGet<{ admin: Admin }>('/admin/me', fetch);

	const saved = cookies.get(COOKIES.sidebar);
	const sidebar: SidebarState = isSidebarState(saved) ? saved : 'wide';
	const closedBranches = parseClosedBranches(cookies.get(COOKIES.branches));

	return { admin, sidebar, closedBranches };
};
