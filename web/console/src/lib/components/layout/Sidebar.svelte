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
</aside>

<style>
	aside {
		/* What every row in the column is drawn with. They are named here, on
		   the one element that owns the navigation, so a row does not have to
		   know which of the palette's greys means "the page you are on". Every
		   one is a solid colour: a wash mixed with transparency reads as a
		   different grey over every surface it lands on. */
		--nav-row-height: 36px;
		/* Every row is in the full text colour: on white a grey name reads as
		   disabled, not as a place to go. Sections are told from pages by
		   weight, not by fading them. */
		--nav-text: var(--color-text);
		--nav-hover: var(--color-secondary);
		/* The page being read is the one row in the brand colour, the way
		   Telegram marks the open chat: nothing else in the column is blue,
		   so it is found without looking for it. */
		--nav-current: var(--color-brand);
		--nav-current-text: var(--color-brand-text);
		--nav-mark: var(--color-brand);
		/* How far a branch's pages are inset from its rule. */
		--nav-branch-inset: 8px;

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

	/* The rows are sized to fit the column without a scrollbar. On a window
	   too short even for that the list still moves under a wheel or a
	   thumb, so nothing becomes unreachable, but it draws no bar and reserves
	   no gutter for one. */
	nav {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: var(--space-1);
		padding: var(--space-2) var(--space-2) var(--space-3);
		overflow-x: hidden;
		overflow-y: auto;
		overscroll-behavior: none;
		scrollbar-width: none;
	}

	nav::-webkit-scrollbar {
		display: none;
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
		padding-bottom: var(--space-1);
		border-bottom: 1px solid var(--color-border);
	}

	/* Folded, the rows are icons and the column is narrow: the same inset
	   on both sides keeps them centred under the header's mark. */
	.collapsed nav {
		gap: 1px;
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
				transform var(--speed-drawer) cubic-bezier(0.4, 0, 0.2, 1),
				visibility var(--speed-drawer);
		}

		aside.open {
			transform: translateX(0);
			visibility: visible;
		}

		/* A panel has room for names, so it never draws itself as a rail. */
		.collapsed nav {
			gap: var(--space-1);
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
