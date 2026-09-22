<script lang="ts">
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { RiSidebarFoldLine, RiSidebarUnfoldLine } from 'svelte-remixicon';
	import type { Admin } from '$lib/api';
	import { Icon, Tooltip } from '$lib/components/ui';
	import SidebarLink from './sidebar/SidebarLink.svelte';
	import { visibleSections, type Section } from './sidebar/sections';
	import { useTranslator } from '$lib/i18n';

	type Props = {
		/** Folded to icons only. The width itself is set by the panel layout. */
		collapsed: boolean;
		onToggle: () => void;
	};

	let { collapsed, onToggle }: Props = $props();

	const groups = $derived(visibleSections(page.data.admin as Admin | undefined));

	/** Activity is the dashboard's own page, so it only matches exactly; the
	    others also match anything below them. */
	function isCurrent(route: Section): boolean {
		const href = resolve(route);
		const path = page.url.pathname;

		return route === '/admin/(panel)/dashboard'
			? path === href
			: path === href || path.startsWith(`${href}/`);
	}

	const t = useTranslator();
</script>

<aside class:collapsed>
	<nav aria-label={t('shell.sections')}>
		{#each groups as group (group.key ?? 'top')}
			<div class="group" role="group" aria-label={group.key ? t(group.key) : undefined}>
				{#if group.key}
					<h2>{t(group.key)}</h2>
				{/if}

				<ul>
					{#each group.items as item (item.route)}
						<li>
							<SidebarLink
								route={item.route}
								label={t(item.key)}
								icon={item.icon}
								current={isCurrent(item.route)}
								{collapsed}
							/>
						</li>
					{/each}
				</ul>
			</div>
		{/each}
	</nav>

	<div class="foot">
		<Tooltip label={t('shell.expand')} placement="right" disabled={!collapsed}>
			{#snippet children(trigger)}
				<button
					{...trigger()}
					type="button"
					class="fold"
					onclick={onToggle}
					aria-expanded={!collapsed}
					aria-label={collapsed ? t('shell.expand') : t('shell.collapse')}
				>
					<span class="icon">
						<Icon icon={collapsed ? RiSidebarUnfoldLine : RiSidebarFoldLine} />
					</span>
					<span class="label">{t('shell.collapse')}</span>
				</button>
			{/snippet}
		</Tooltip>
	</div>
</aside>

<style>
	aside {
		position: fixed;
		top: var(--header-height);
		bottom: 0;
		left: 0;
		z-index: 9;
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		gap: var(--space-3);
		width: var(--sidebar-width);
		padding: 8px;
		border-right: 1px solid var(--color-border);
		background: var(--color-surface);
		overflow-x: hidden;
		overflow-y: auto;
		overscroll-behavior: none;
		scrollbar-width: thin;
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

	/* A group is its heading and its links; the rule between groups only
	   shows once the headings have folded away, so the column still reads as
	   groups without their names. */
	.group + .group {
		margin-top: 6px;
		border-top: 1px solid transparent;
		transition:
			border-color var(--speed),
			padding var(--speed);
	}

	h2 {
		height: 26px;
		margin: 0;
		padding: 0 12px;
		overflow: hidden;
		color: var(--color-text-hint);
		font-size: 11px;
		font-weight: 600;
		line-height: 28px;
		letter-spacing: 0.06em;
		text-transform: uppercase;
		white-space: nowrap;
		transition:
			height var(--speed),
			opacity var(--speed);
	}

	/* Folded: the headings close up and a rule takes their place. */
	.collapsed .group + .group {
		padding-top: 6px;
		border-top-color: var(--color-border);
	}

	.collapsed h2 {
		height: 0;
		opacity: 0;
	}

	.foot {
		padding-top: 8px;
		border-top: 1px solid var(--color-border);
	}

	/* The fold button is laid out like a link, so it sits in the column as
	   one more row rather than a control bolted underneath. */
	.fold {
		display: flex;
		align-items: center;
		gap: 10px;
		width: 100%;
		height: var(--nav-item-height);
		padding: 0 12px;
		overflow: hidden;
		border: none;
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--color-text-hint);
		font: inherit;
		font-size: var(--text-base);
		white-space: nowrap;
		cursor: pointer;
		transition:
			background-color var(--speed-fast),
			color var(--speed-fast);
	}

	.fold:hover {
		background: var(--color-secondary);
		color: var(--color-text);
	}

	.fold:focus-visible {
		outline: 2px solid var(--color-accent);
		outline-offset: -2px;
	}

	.fold .icon {
		display: inline-flex;
		flex: none;
		width: 16px;
		justify-content: center;
	}

	.fold .label {
		transition: opacity var(--speed);
	}

	.collapsed .fold .label {
		opacity: 0;
	}

	/* Narrow screens have no room for a column, so the sections become one
	   scrollable row above the content. */
	@media (max-width: 55rem) {
		aside {
			position: sticky;
			top: var(--header-height);
			z-index: 9;
			width: auto;
			height: auto;
			padding: 6px var(--space-2);
			border-right: none;
			border-bottom: 1px solid var(--color-border);
			overflow-x: auto;
			overscroll-behavior-x: none;
		}

		nav,
		ul {
			flex-direction: row;
			gap: 2px;
		}

		.group + .group,
		.collapsed .group + .group {
			margin: 0 0 0 2px;
			padding: 0 0 0 4px;
			border-top: none;
			border-left: 1px solid var(--color-border);
		}

		h2,
		.foot {
			display: none;
		}

		/* Folding only means something beside the content, so here every
		   section keeps its name, and the marks are left to the pages. */
		aside :global(.label) {
			opacity: 1 !important;
		}

		aside :global(.soon),
		aside :global(.dot) {
			display: none;
		}
	}
</style>
