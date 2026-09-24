<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { MediaQuery } from 'svelte/reactivity';
	import type { Admin } from '$lib/api';
	import { useShell } from '$lib/state/shell.svelte';
	import SidebarBranch from './sidebar/SidebarBranch.svelte';
	import SidebarLink from './sidebar/SidebarLink.svelte';
	import { visibleBranches, visibleOverview, type Section } from './sidebar/sections';

	const shell = useShell();

	/** Below this there is no room for a column beside the page, so the
	    sidebar becomes a panel over it — the same rows, slid in from the edge
	    and dismissed when one is chosen. The width is the styles' too; it is
	    here as well because the choice is which component to draw, not only
	    how to draw it. */
	const narrow = new MediaQuery('max-width: 55rem');

	/** Folded to icons: only the column does that. A panel has the width for
	    names, so its branches open underneath rather than beside. */
	const folded = $derived(shell.collapsed && !narrow.current);

	const admin = $derived(page.data.admin as Admin | undefined);
	const overview = $derived(visibleOverview(admin));
	const branches = $derived(visibleBranches(admin));

	/** Activity is the dashboard's own page, so it only matches exactly; the
	    others also match anything below them. */
	function isCurrent(route: Section): boolean {
		const href = resolve(route);
		const path = page.url.pathname;

		return route === '/admin/(panel)/dashboard'
			? path === href
			: path === href || path.startsWith(`${href}/`);
	}

	// Choosing a page is finishing with the panel, so it closes itself rather
	// than staying over what it was asked to show.
	$effect(() => {
		void page.url.pathname;
		shell.setMenu(false);
	});

	// A panel is also finished with when the window grows enough to hold a
	// column: left open, it would sit over a page that already has one.
	$effect(() => {
		if (!narrow.current) shell.setMenu(false);
	});

	// Escape closes it — listened for on the way down rather than on the way
	// up, because the control that opens the panel carries a tooltip, and a
	// tooltip takes Escape for itself before it reaches the window.
	$effect(() => {
		function dismiss(event: KeyboardEvent) {
			if (event.key === 'Escape' && shell.menuOpen) shell.setMenu(false);
		}

		window.addEventListener('keydown', dismiss, true);

		return () => window.removeEventListener('keydown', dismiss, true);
	});
</script>

<!-- Only ever over the page, and only while it is open: at column widths
     there is nothing to dim. -->
{#if narrow.current && shell.menuOpen}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div class="backdrop" onclick={() => shell.setMenu(false)}></div>
{/if}

<aside class:collapsed={folded} class:open={shell.menuOpen}>
	<!-- The frame holds the border and the background and never scrolls; the
	     list inside it does. That is what keeps the scrollbar off the column's
	     edge, and what lets the gutter be reserved once rather than appearing
	     under whichever branch was opened last. -->
	<div class="scroll">
		<nav aria-label="Sections">
			{#if overview.length > 0}
				<ul class="top">
					{#each overview as item (item.route)}
						<li>
							<SidebarLink
								route={item.route}
								label={item.label}
								icon={item.icon}
								current={isCurrent(item.route)}
								collapsed={folded}
							/>
						</li>
					{/each}
				</ul>
			{/if}

			{#each branches as branch (branch.id)}
				<div class="branch">
					<SidebarBranch
						{branch}
						{isCurrent}
						collapsed={folded}
						open={shell.isOpen(branch.id)}
						onToggle={() => shell.toggleBranch(branch.id)}
					/>
				</div>
			{/each}
		</nav>
	</div>
</aside>

<style>
	aside {
		/* What every row in the column is drawn with. They are named here, on
		   the one element that owns the navigation, so a row does not have to
		   know which of the palette's greys means "the page you are on". */
		--nav-row-height: 36px;
		--nav-hover: color-mix(in srgb, var(--color-secondary) 60%, transparent);
		--nav-current: var(--color-secondary-alt);
		--nav-group-active: color-mix(in srgb, var(--color-secondary-alt) 42%, transparent);
		--nav-mark: var(--color-info);
		/* How far a branch's pages are inset from its rule. */
		--nav-branch-inset: 10px;
		/* Content ends before the scrollbar lane, with a visible breathing
		   room for overlay bars that do not take layout space. */
		--nav-scrollbar-inset: calc(var(--scrollbar-size) + var(--space-2));

		position: fixed;
		top: var(--header-height);
		bottom: 0;
		left: 0;
		z-index: 9;
		display: flex;
		flex-direction: column;
		width: var(--sidebar-width);
		border-right: 1px solid var(--color-border);
		background: var(--color-surface);
		overflow: hidden;
	}

	/* The one thing that scrolls: the frame around it keeps the border and the
	   background, so the bar is never against the column's edge and opening a
	   branch cannot make the whole column jump.

	   The gutter is reserved whether or not there is anything to scroll, so
	   growing the list never shifts the rows sideways. The right padding also
	   leaves a visible gap after the scrollbar lane, which matters on systems
	   where the bar floats over the content instead of taking its own space. */
	.scroll {
		flex: 1;
		padding: var(--space-2) var(--nav-scrollbar-inset) var(--space-3) var(--space-2);
		overflow-x: hidden;
		overflow-y: auto;
		overscroll-behavior: none;
		scrollbar-gutter: stable;
	}

	nav {
		display: flex;
		flex-direction: column;
	}

	ul {
		display: flex;
		flex-direction: column;
		gap: 1px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	/* The pages that stand on their own are ruled off from the sections under
	   them: one line, where the old column needed a heading over every group
	   to say the same thing. */
	.top {
		padding-bottom: var(--space-2);
		margin-bottom: var(--space-2);
		border-bottom: 1px solid var(--color-border);
	}

	.branch + .branch {
		margin-top: var(--space-2);
		padding-top: var(--space-1);
		border-top: 1px solid color-mix(in srgb, var(--color-border) 72%, transparent);
	}

	/* Folded, the rows are icons: the gutter would be most of the column, so
	   the list keeps its inset and loses the reservation. Nothing under the
	   fold is long enough to scroll anyway. */
	.collapsed .scroll {
		padding-inline: 8px;
		scrollbar-gutter: auto;
	}

	/* The rail is a compact set of icons, not a stack of section cards. */
	.collapsed .branch + .branch {
		margin-top: 1px;
		padding-top: 0;
		border-top: 0;
	}

	/* ---- As a panel, on a screen too narrow for a column ------------------
	   The same rows, slid in from the edge over a dimmed page, rather than a
	   row of icons above it: a list that reads top to bottom is a list
	   somebody can use with a thumb.

	   It is a media query and not only the class the script adds, so a narrow
	   screen is served a panel that is already off the edge. Waiting for the
	   script would show the column first and take it away. */
	@media (max-width: 55rem) {
		aside {
			width: min(17rem, 82vw);
			box-shadow: var(--shadow-md);
			transform: translateX(-100%);
			visibility: hidden;
			transition:
				transform var(--speed-drawer) cubic-bezier(0.4, 0, 0.2, 1),
				visibility var(--speed-drawer);
		}

		aside.open {
			transform: translateX(0);
			visibility: visible;
		}

		/* A panel has room for names, so it never draws itself as a rail. */
		.collapsed .scroll {
			padding: var(--space-2) var(--nav-scrollbar-inset) var(--space-3) var(--space-2);
			scrollbar-gutter: stable;
		}
	}

	.backdrop {
		position: fixed;
		top: var(--header-height);
		right: 0;
		bottom: 0;
		left: 0;
		z-index: 8;
		background: var(--color-overlay);
		animation: fade var(--speed) ease-out;
	}

	@keyframes fade {
		from {
			opacity: 0;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		aside {
			transition: none;
		}

		.backdrop {
			animation: none;
		}
	}
</style>
