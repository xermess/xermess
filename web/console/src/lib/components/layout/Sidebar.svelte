<script lang="ts">
	import { page } from '$app/state';
	import { MediaQuery } from 'svelte/reactivity';
	import type { Admin } from '$lib/api';
	import { useShell } from '$lib/state/shell.svelte';
	import SidebarGroup from './sidebar/SidebarGroup.svelte';
	import { isCurrentSection, visibleGroups, type Section } from './sidebar/sections';

	const shell = useShell();

	/** Below this there is no room for a column beside the page, so the
	    sidebar becomes a panel over it — the same rows, slid in from the edge
	    and dismissed when one is chosen. The width is the styles' too; it is
	    here as well because the choice is which component to draw, not only
	    how to draw it. */
	const narrow = new MediaQuery('max-width: 55rem');

	/** Folded to icons: only the column does that. A panel has the width for
	    names, so it always shows them. */
	const folded = $derived(shell.collapsed && !narrow.current);

	const admin = $derived(page.data.admin as Admin | undefined);
	const groups = $derived(visibleGroups(admin));

	// The page being read, by its route id rather than its path: that is the
	// shape the section list is written in, so a row is marked without either
	// side having to resolve anything.
	const isCurrent = (route: Section) => isCurrentSection(route, page.route.id);

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
	<nav aria-label="Sections">
		{#each groups as group (group.id)}
			<SidebarGroup
				{group}
				{isCurrent}
				collapsed={folded}
				open={shell.isOpen(group.id)}
				onToggle={() => shell.toggleGroup(group.id)}
			/>
		{/each}
	</nav>
</aside>

<style>
	aside {
		/* Its colours are the nav-* tokens in theme.ts. */
		position: fixed;
		top: var(--header-height);
		bottom: 0;
		left: 0;
		z-index: 9;
		display: flex;
		flex-direction: column;
		width: var(--sidebar-width);
		border-right: 1px solid var(--color-border);
		background: var(--nav-surface);
		overflow: hidden;
	}

	/* Groups a little apart, rows close together: the space says which rows
	   belong with which heading, with no rules or boxes needed. The list
	   scrolls on a window too short for it, with a thin bar that shows only
	   while the pointer is over the column. */
	nav {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: var(--space-2);
		padding: var(--space-2) var(--space-2) var(--space-3);
		overflow-x: hidden;
		overflow-y: auto;
		overscroll-behavior: contain;
		scrollbar-width: thin;
		scrollbar-color: transparent transparent;
	}

	nav:hover {
		scrollbar-color: var(--scrollbar-thumb) transparent;
	}

	/* Folded, the rows are icons and the column is narrow: the same inset
	   on both sides keeps them centred under the header's mark, and the
	   groups are held apart by their rules rather than by space. */
	.collapsed nav {
		gap: 0;
		padding-inline: 8px;
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
				transform var(--speed-slow) cubic-bezier(0.4, 0, 0.2, 1),
				visibility var(--speed-slow);
		}

		aside.open {
			transform: translateX(0);
			visibility: visible;
		}

		/* A panel has room for names, so it never draws itself as a rail. */
		.collapsed nav {
			gap: var(--space-2);
			padding-inline: var(--space-2);
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
