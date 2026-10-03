<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import {
		RiCloseLine,
		RiMenuLine,
		RiSidebarFoldLine,
		RiSidebarUnfoldLine
	} from 'svelte-remixicon';
	import type { Admin, OrganizationBrand } from '$lib/api';
	import { BRAND } from '$lib/brand';
	import { IconButton, ThemeToggle, Thumb } from '$lib/components/ui';
	import { useShell } from '$lib/state/shell.svelte';
	import { initials } from '$lib/utils/format';
	import AccountMenu from './AccountMenu.svelte';
	import CommandPalette from './CommandPalette.svelte';
	import HelpMenu from './HelpMenu.svelte';

	type Props = { admin: Admin; organization: OrganizationBrand };

	let { admin, organization }: Props = $props();

	/** The logo block is the top of the sidebar's column, so it folds with it. */
	const shell = useShell();

	/** The dashboard is where the sidebar is, and so where the header carries
	    its controls and the logo block is ruled off as the top of its column.
	    Elsewhere, such as the profile, it keeps the width without a line
	    leading nowhere. */
	const besideSidebar = $derived(page.route.id?.startsWith('/admin/(panel)/dashboard') ?? false);

	/** The organisation's name, or the product's on an installation nobody
	    has named yet. */
	const name = $derived(organization.name.trim() || BRAND.name);
</script>

<!--
	The header row: logo and sidebar control, then search, help, theme and the account menu. On
	wide screens the logo block tops the sidebar column; on narrow ones the menu button comes
	first. CSS picks which control shows, so the server-rendered page is right at any width.
-->
<header class:mini={shell.collapsed} class:dashboard={besideSidebar}>
	{#if besideSidebar}
		<div class="menu-control">
			<IconButton
				icon={shell.menuOpen ? RiCloseLine : RiMenuLine}
				label={shell.menuOpen ? 'Close the menu' : 'Open the menu'}
				size="sm"
				placement="right"
				onclick={() => shell.setMenu(!shell.menuOpen)}
			/>
		</div>
	{/if}

	<div class="brand-column">
		<a class="brand" href={resolve('/admin/dashboard')} aria-label="{name} Console">
			<Thumb
				src={organization.logo_url}
				text={initials(name)}
				size="xs"
				shape="circle"
				tone="accent"
				bare
			/>

			<span class="wordmark">Console</span>
		</a>
	</div>

	{#if besideSidebar}
		<div class="fold-control">
			<IconButton
				icon={shell.collapsed ? RiSidebarUnfoldLine : RiSidebarFoldLine}
				label={shell.collapsed ? 'Expand the sidebar' : 'Collapse the sidebar'}
				size="sm"
				onclick={shell.toggle}
			/>
		</div>
	{/if}

	<!-- The tools are quiet icons, grouped at the far end; a hairline, then
	     who is signed in, which is a different kind of thing. -->
	<div class="end">
		<CommandPalette />
		<HelpMenu />
		<ThemeToggle size="sm" variant="ghost" />
		<span class="rule" aria-hidden="true"></span>
		<AccountMenu {admin} />
	</div>
</header>

<style>
	/* The bar stays put while the page scrolls, so the account menu is always
	   one click away on a long table. Where to go is the sidebar's job. */
	header {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		z-index: 10;
		display: flex;
		align-items: center;
		gap: var(--space-2);
		height: var(--header-height);
		padding-right: var(--page-gutter);
		border-bottom: 1px solid var(--color-border);
		background: var(--color-surface);
		overscroll-behavior: none;
	}

	/* The top of the sidebar's column: exactly as wide, and on the dashboard
	   ruled off on the same line as the sidebar's edge and wearing its tint,
	   so the two read as one block. The width comes from the panel layout,
	   which animates it, so both fold on the same frames. */
	.brand-column {
		display: flex;
		flex-shrink: 0;
		align-items: center;
		align-self: stretch;
		width: var(--sidebar-width);
		min-width: 0;
		overflow: hidden;
		border-right: 1px solid transparent;
		transition: border-color var(--speed);
	}

	.dashboard .brand-column {
		border-right-color: var(--color-border);
		background: var(--nav-surface);
	}

	/* The mark sits over the sidebar's icons, centred on the same line folded
	   or not. Nothing moves sideways while the column folds; the word just
	   runs out of room and fades. */
	.brand {
		display: flex;
		align-items: center;
		gap: 10px;
		min-width: 0;
		height: 100%;
		padding: 0 var(--space-3) 0 14px;
		color: var(--color-text);
		text-decoration: none;
		white-space: nowrap;
	}

	.brand:focus-visible {
		outline: 2px solid var(--color-info);
		outline-offset: -4px;
	}

	/* What this is, in the display face so it reads as a mark rather than a
	   label: Poppins' round, geometric letters, set tight. A line height of
	   one keeps the box to the letters, so centring it beside the mark
	   centres the word. */
	.wordmark {
		font-family: var(--font-display);
		font-size: var(--text-xl);
		font-weight: 600;
		line-height: 1;
		letter-spacing: -0.03em;
		transition: opacity var(--speed);
	}

	.end {
		display: flex;
		flex-shrink: 0;
		align-items: center;
		gap: 2px;
		margin-left: auto;
	}

	/* A hairline, not a gap: the account is a different kind of thing from
	   the icons beside it. */
	.rule {
		width: 1px;
		height: 20px;
		margin: 0 var(--space-2);
		background: var(--color-border);
	}

	/* ---- A window wide enough for the sidebar column ------------------- */
	@media (min-width: 55.0625rem) {
		.menu-control {
			display: none;
		}

		/* Folded, the column is the mark alone. */
		.mini .wordmark {
			opacity: 0;
		}
	}

	/* ---- A window too narrow for it (breakpoints.sidebar) -------------- */
	@media (max-width: 55rem) {
		header {
			gap: var(--space-1);
			padding-left: var(--space-2);
		}

		.fold-control {
			display: none;
		}

		/* No column to top: the logo block is only as wide as the mark and
		   the word, on the header's own white. */
		.brand-column,
		.dashboard .brand-column {
			width: auto;
			flex-shrink: 1;
			border-right-color: transparent;
			background: none;
		}

		.brand {
			gap: var(--space-2);
			padding: 0 var(--space-1);
		}

		.rule {
			margin: 0 var(--space-1);
		}
	}
</style>
