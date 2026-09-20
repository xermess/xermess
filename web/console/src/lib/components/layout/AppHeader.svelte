<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { RiBookOpenLine, RiGithubFill, RiShieldKeyholeLine } from 'svelte-remixicon';
	import type { Admin } from '$lib/api';
	import { Icon, IconLink, ThemeToggle } from '$lib/components/ui';
	import { DOCS_URL, GITHUB_URL } from '$lib/constants';
	import { useTranslator } from '$lib/i18n';
	import { useShell } from '$lib/state/shell.svelte';
	import AccountMenu from './AccountMenu.svelte';
	import CommandPalette from './CommandPalette.svelte';

	type Props = { admin: Admin };

	let { admin }: Props = $props();

	const t = useTranslator();

	/** The logo block is the top of the sidebar's column, so it folds with it. */
	const shell = useShell();

	/** The dashboard is where the sidebar is, and so where the logo block is
	    ruled off as the top of its column. Elsewhere, such as the profile, it
	    keeps the width without a line leading nowhere. */
	const besideSidebar = $derived(page.route.id?.startsWith('/admin/(panel)/dashboard') ?? false);
</script>

<header class:mini={shell.collapsed}>
	<div class="brand-column" class:ruled={besideSidebar}>
		<a class="brand" href={resolve('/admin/dashboard')} aria-label="xermess">
			<span class="mark">
				<Icon icon={RiShieldKeyholeLine} size="1.125rem" />
			</span>
			<strong aria-hidden={shell.collapsed}>xermess</strong>
		</a>
	</div>

	<!-- The one thing in the bar worth the width: everywhere the panel can go,
	     two or three letters away. -->
	<div class="search">
		<CommandPalette />
	</div>

	<!-- Somewhere to read rather than something to do, so they are quiet: an
	     icon each, named by a tooltip, grouped away from the page's own
	     buttons. The account menu keeps both as named rows, which is what a
	     narrow screen is left with. -->
	<div class="utilities">
		<IconLink
			href={DOCS_URL}
			icon={RiBookOpenLine}
			label={t('shell.documentation')}
			size="sm"
			target="_blank"
			rel="noreferrer noopener"
		/>

		<IconLink
			href={GITHUB_URL}
			icon={RiGithubFill}
			label={t('shell.github')}
			size="sm"
			target="_blank"
			rel="noreferrer noopener"
		/>

		<span class="rule" aria-hidden="true"></span>

		<ThemeToggle size="sm" variant="ghost" />
	</div>

	<div class="account">
		<AccountMenu {admin} size="sm" />
	</div>
</header>

<style>
	/* The bar stays put while the page scrolls, so the account menu is always
	   one click away on a long table. Where to go is the sidebar's job. */
	header {
		position: sticky;
		top: 0;
		z-index: 10;
		display: flex;
		align-items: center;
		height: var(--header-height);
		padding-right: var(--space-4);
		border-bottom: 1px solid var(--color-border);
		background: var(--color-surface);
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

	/* Beside the sidebar, the block also runs over the header's bottom line,
	   so the logo and the sections under it are one column with no seam. */
	.brand-column.ruled {
		position: relative;
		align-self: flex-start;
		height: calc(100% + 1px);
		border-right-color: var(--color-border);
		background: var(--color-surface);
	}

	/* The mark sits over the sidebar's icons, which are 20px in from the
	   edge: it is 28px wide to their 16px, so it starts 6px earlier and the
	   two are centred on one line, folded or not. Nothing moves sideways
	   while the column folds; the name just runs out of room and fades. */
	.brand {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		height: 100%;
		padding-left: 14px;
		border-radius: var(--radius-sm);
		font-size: var(--text-lg);
		text-decoration: none;
		white-space: nowrap;
		color: var(--color-text);
	}

	.brand:focus-visible {
		outline: 2px solid var(--color-accent);
		outline-offset: -4px;
	}

	.mark {
		display: grid;
		flex-shrink: 0;
		place-items: center;
		width: 28px;
		height: 28px;
		border-radius: var(--radius-sm);
		background: var(--color-accent);
		color: var(--color-accent-text);
	}

	.brand strong {
		transition: opacity var(--speed);
	}

	.mini .brand strong {
		opacity: 0;
	}

	/* The start of the content area, just past the sidebar's column, so the
	   field lines up with the page's own heading rather than floating in the
	   gap. It takes the width it is given and stops: the bar is not a form. */
	.search {
		flex: 1;
		min-width: 0;
		max-width: 18rem;
		padding-left: var(--space-4);
	}

	/* Everything that is not the page and not the account, kept together at
	   the end so the bar reads left to right: where you are, where you can
	   go, what is around it, who you are. */
	.utilities {
		margin-left: auto;
		display: flex;
		align-items: center;
		gap: var(--space-1);
	}

	/* A hairline, not a gap: it says the account is a different kind of thing
	   from the links beside it without spending the width a gap would. */
	.rule {
		width: 1px;
		height: 18px;
		margin: 0 var(--space-1);
		background: var(--color-border);
	}

	.account {
		display: flex;
		align-items: center;
		padding-left: var(--space-1);
	}

	/* Narrow screens have no sidebar column, only a row of sections under the
	   bar, so the logo block is only as wide as the mark. */
	@media (max-width: 55rem) {
		header {
			padding-right: var(--space-2);
		}

		.brand-column,
		.brand-column.ruled {
			align-self: stretch;
			width: auto;
			height: auto;
			border-right-color: transparent;
		}

		.brand {
			padding: 0 var(--space-2);
		}

		.brand strong {
			display: none;
		}

		/* No room for them beside the sections; the account menu still has
		   both as named rows, which is where they are looked for anyway.
		   The palette stays — folded to its magnifier — because it is the
		   only way to reach a page quickly without the sidebar. */
		.utilities :global(a.control) {
			display: none;
		}

		.rule {
			display: none;
		}

		.search {
			flex: 0 0 auto;
			padding-left: var(--space-2);
		}
	}
</style>
