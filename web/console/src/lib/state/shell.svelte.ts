import { getContext, setContext } from 'svelte';
import { SvelteSet } from 'svelte/reactivity';
import { rememberClosedBranches, rememberSidebar, type SidebarState } from './sidebar';

/**
 * How the panel's own frame is left: how wide the sidebar is, and which of its
 * sections are folded away.
 *
 * The header's logo block and the dashboard's sidebar are one column of the
 * screen, so they have to agree on how wide it is and fold together. The
 * panel layout makes this once per page view and hands it down as context —
 * never as module state, which on the server would be shared by every
 * request.
 */
export type Shell = {
	readonly collapsed: boolean;
	toggle: () => void;

	/** Whether a branch of the sidebar is showing its pages. */
	isOpen: (id: string) => boolean;
	toggleBranch: (id: string) => void;

	/** Whether the sidebar is open as a panel over the page, which is what it
	    is on a screen too narrow for a column beside one. It is not
	    remembered: a panel somebody left open is not a preference, and a page
	    that loads with one over it is a page nobody asked for. */
	readonly menuOpen: boolean;
	setMenu: (open: boolean) => void;
};

const key = Symbol('shell');

/** Makes the shared state, starting as the cookies the server read say. */
export function provideShell(initial: SidebarState, closedBranches: string[]): Shell {
	let collapsed = $state(initial === 'mini');

	// The closed ones are held, not the open ones, so a section a release adds
	// arrives open rather than hidden from whoever had folded the others.
	const closed = new SvelteSet(closedBranches);

	let menuOpen = $state(false);

	const shell: Shell = {
		get collapsed() {
			return collapsed;
		},

		toggle() {
			collapsed = !collapsed;
			rememberSidebar(collapsed ? 'mini' : 'wide');
		},

		isOpen(id: string) {
			return !closed.has(id);
		},

		toggleBranch(id: string) {
			if (closed.has(id)) {
				closed.delete(id);
			} else {
				closed.add(id);
			}

			rememberClosedBranches([...closed]);
		},

		get menuOpen() {
			return menuOpen;
		},

		setMenu(open: boolean) {
			menuOpen = open;
		}
	};

	return setContext(key, shell);
}

/** The shared state, from a component inside the panel layout. */
export function useShell(): Shell {
	const shell = getContext<Shell | undefined>(key);
	if (!shell) throw new Error('useShell() needs the panel layout above it');

	return shell;
}
