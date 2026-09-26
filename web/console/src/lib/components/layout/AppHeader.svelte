<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { MediaQuery } from 'svelte/reactivity';
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
	import { breakpoints } from '$lib/theme';
	import { initials } from '$lib/utils/format';
	import AccountMenu from './AccountMenu.svelte';
	import CommandPalette from './CommandPalette.svelte';
	import HelpMenu from './HelpMenu.svelte';

	type Props = { admin: Admin; organization: OrganizationBrand };

	let { admin, organization }: Props = $props();

	/** The logo block is the top of the sidebar's column, so it folds with it. */
	const shell = useShell();

	/** The dashboard is where the sidebar is, and so where the logo block is
	    ruled off as the top of its column. Elsewhere, such as the profile, it
	    keeps the width without a line leading nowhere. */
	const besideSidebar = $derived(page.route.id?.startsWith('/admin/(panel)/dashboard') ?? false);

	/** The same width the sidebar reads: below it the column is a panel, so
	    the control beside the logo opens and closes that panel rather than
	    folding a column that is not there. */
	const narrow = new MediaQuery(`max-width: ${breakpoints.sidebar}`);

	const nav = $derived(
		narrow.current
			? {
					icon: shell.menuOpen ? RiCloseLine : RiMenuLine,
					label: shell.menuOpen ? 'Close the menu' : 'Open the menu',
					press: () => shell.setMenu(!shell.menuOpen)
				}
			: {
					icon: shell.collapsed ? RiSidebarUnfoldLine : RiSidebarFoldLine,
					label: shell.collapsed ? 'Expand the sidebar' : 'Collapse the sidebar',
					press: shell.toggle
				}
	);

	/** The organisation's name, or the product's on an installation nobody
	    has named yet. */
	const name = $derived(organization.name.trim() || BRAND.name);
</script>

<!-- Read left to right: whose panel this is, the sidebar's control, and at
     the far end the tools — search, help, the theme — then who is signed in. -->
<header class:mini={shell.collapsed}>
	<div class="brand-column" class:ruled={besideSidebar}>
		<a class="brand" href={resolve('/admin/dashboard')} aria-label="{name} Console">
			<Thumb src={organization.logo_url} text={initials(name)} size="xs" tone="accent" bare />

			<span class="names" aria-hidden={shell.collapsed}>
				<strong title={name}>{name}</strong>
				<small>Console</small>
			</span>
		</a>
	</div>

	<div class="bar">
		<div class="start">
			<!-- The control for the navigation: it folds the column where there
			     is one, and opens the panel where the column has become one.
			     Only on the dashboard, the one page with navigation beside it. -->
			{#if besideSidebar}
				<IconButton icon={nav.icon} label={nav.label} size="sm" onclick={nav.press} />
			{/if}
		</div>

		<!-- Search, help and the theme are quiet icons, grouped; a hairline,
		     then who is signed in, which is a different kind of thing. -->
		<div class="end">
			<CommandPalette />
			<HelpMenu />
			<ThemeToggle size="sm" variant="ghost" />
			<span class="rule" aria-hidden="true"></span>
			<AccountMenu {admin} />
		</div>
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
		height: var(--header-height);
		border-bottom: 1px solid var(--color-border);
		background: var(--color-surface);
		overscroll-behavior: none;
	}

	/* The top of the sidebar's column: exactly as wide, and on the dashboard
	   ruled off on the same line as the sidebar's edge, so the two read as one
	   block. The width comes from the panel layout, which animates it, so
	   both fold on the same frames. */
	.brand-column {
		display: flex;
		flex-shrink: 0;
		align-items: center;
		align-self: stretch;
		width: var(--sidebar-width);
		overflow: hidden;
		border-right: 1px solid transparent;
		transition: border-color var(--speed);
	}

	.brand-column.ruled {
		border-right-color: var(--color-border);
	}

	/* The mark sits over the sidebar's icons, centred on the same line folded
	   or not. Nothing moves sideways while the column folds; the names just
	   run out of room and fade. */
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
		outline: 2px solid var(--color-accent);
		outline-offset: -4px;
	}

	/* The organisation, and underneath it, quieter, what this panel is. */
	.names {
		display: flex;
		flex-direction: column;
		min-width: 0;
		line-height: 1.25;
		transition: opacity var(--speed);
	}

	.names strong {
		overflow: hidden;
		font-size: var(--text-base);
		font-weight: 600;
		text-overflow: ellipsis;
	}

	.names small {
		color: var(--color-text-hint);
		font-size: var(--text-xs);
	}

	.mini .names {
		opacity: 0;
	}

	/* Everything past the logo block: the sidebar's control at the start,
	   and the tools pushed to the end. */
	.bar {
		display: flex;
		flex: 1;
		align-items: center;
		gap: var(--space-2);
		min-width: 0;
		height: 100%;
		padding: 0 var(--page-gutter) 0 var(--space-2);
	}

	.start {
		display: flex;
		align-items: center;
	}

	.end {
		display: flex;
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

	/* Narrow screens have no sidebar column — the sections are a panel the
	   control beside the logo opens — so the logo block is only as wide as
	   the mark, and the search takes the room in the middle. */
	@media (max-width: 55rem) {
		.brand-column,
		.brand-column.ruled {
			width: auto;
			border-right-color: transparent;
		}

		.brand {
			padding: 0 var(--space-1) 0 var(--space-3);
		}

		.names {
			display: none;
		}

		.bar {
			padding-left: 0;
		}
	}

	@media (max-width: 40rem) {
		.rule {
			margin: 0 var(--space-1);
		}
	}
</style>
