<script lang="ts">
	import { RiArrowDownSLine } from 'svelte-remixicon';
	import { Icon } from '$lib/components/ui';
	import SidebarLink from './SidebarLink.svelte';
	import type { Section, SidebarGroup } from './sections';

	type Props = {
		group: SidebarGroup;
		/** Showing its pages. Ignored while the column is folded to icons,
		    where every page is one icon and there is nothing to fold. */
		open: boolean;
		onToggle: () => void;
		/** Which page is being read. */
		isCurrent: (route: Section) => boolean;
		/** The column is folded to its icons. */
		collapsed: boolean;
	};

	let { group, open, onToggle, isCurrent, collapsed }: Props = $props();

	/** A folded group still says it holds the page being read, so hiding a
	    group never hides where you are. */
	const holdsCurrent = $derived(group.items.some((item) => isCurrent(item.route)));

	const shown = $derived(open || collapsed);
</script>

<!-- A heading and its pages. The heading is quiet — small, and in the hint
     colour — so the column reads as its pages, sorted, rather than as a
     tree. It folds its group; folded to icons, it is a rule between groups
     instead, and every page is an icon with its name in a tooltip. -->
<section
	class="group"
	class:collapsed
	aria-labelledby={collapsed ? undefined : `group-${group.id}`}
>
	{#if collapsed}
		<hr aria-hidden="true" />
	{:else}
		<h2>
			<button
				type="button"
				id="group-{group.id}"
				class="heading"
				aria-expanded={open}
				aria-controls="pages-{group.id}"
				onclick={onToggle}
			>
				<span class="name">{group.label}</span>
				{#if !open && holdsCurrent}
					<span class="dot" aria-label="Holds the page you are on"></span>
				{/if}
				<span class="chevron" class:closed={!open} aria-hidden="true">
					<Icon icon={RiArrowDownSLine} size="0.875rem" />
				</span>
			</button>
		</h2>
	{/if}

	<!-- Kept in the markup while folded, so unfolding is a height change
	     rather than a mount; hidden from the keyboard and assistive
	     technology while it is. -->
	<div id="pages-{group.id}" class="pages" class:open={shown} inert={!shown}>
		<ul>
			{#each group.items as item (item.route)}
				<li>
					<SidebarLink
						route={item.route}
						label={item.label}
						icon={item.icon}
						current={isCurrent(item.route)}
						{collapsed}
					/>
				</li>
			{/each}
		</ul>
	</div>
</section>

<style>
	.group {
		display: flex;
		flex-direction: column;
	}

	h2 {
		margin: 0;
	}

	.heading {
		display: flex;
		align-items: center;
		gap: 6px;
		width: 100%;
		height: 26px;
		padding: 0 8px 0 12px;
		border: none;
		border-radius: var(--radius-control);
		background: transparent;
		color: var(--color-text-hint);
		font: inherit;
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		text-align: left;
		cursor: pointer;
		transition: color var(--speed-fast);
	}

	.heading:hover {
		color: var(--color-text);
	}

	.heading:focus-visible {
		outline: 2px solid var(--color-info);
		outline-offset: -2px;
	}

	.name {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	/* A folded group that holds the page being read. */
	.dot {
		flex: none;
		width: 6px;
		height: 6px;
		border-radius: var(--radius-pill);
		background: var(--nav-current);
	}

	/* The chevron is quiet until the heading is under the pointer, and always
	   there on a folded group, which is the one that needs finding. */
	.chevron {
		display: inline-flex;
		flex: none;
		opacity: 0;
		transition:
			opacity var(--speed-fast),
			transform var(--speed);
	}

	.heading:hover .chevron,
	.heading:focus-visible .chevron,
	.chevron.closed {
		opacity: 1;
	}

	.chevron.closed {
		transform: rotate(-90deg);
	}

	/* Unfolding is a height, not a jump. The list is the grid item that can
	   shrink to nothing. */
	.pages {
		display: grid;
		grid-template-rows: 0fr;
		visibility: hidden;
		transition:
			grid-template-rows var(--speed) ease,
			visibility 0s linear var(--speed);
	}

	.pages.open {
		grid-template-rows: 1fr;
		visibility: visible;
		transition: grid-template-rows var(--speed) ease;
	}

	ul {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-height: 0;
		margin: 0;
		padding: 0;
		overflow: hidden;
		list-style: none;
	}

	/* Folded to icons: a short rule between groups, centred under the icons,
	   and none above the first. */
	hr {
		width: 20px;
		margin: 6px auto;
		border: none;
		border-top: 1px solid var(--color-border);
	}

	.group:first-child hr {
		display: none;
	}

	@media (prefers-reduced-motion: reduce) {
		.pages,
		.pages.open,
		.chevron {
			transition: none;
		}
	}
</style>
