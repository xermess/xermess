<script lang="ts">
	import { resolve } from '$app/paths';
	import type { ComponentType } from 'svelte';
	import { Icon, Tooltip } from '$lib/components/ui';
	import type { Section } from './sections';

	type Props = {
		route: Section;
		label: string;
		icon: ComponentType;
		current: boolean;
		/** Folded to its icon: the name moves into a tooltip. */
		collapsed: boolean;
	};

	let { route, label, icon, current, collapsed }: Props = $props();
</script>

<!-- Unfolded, the tooltip would only repeat the name, so it is off. -->
<Tooltip {label} placement="right" disabled={!collapsed}>
	{#snippet children(trigger)}
		<a
			{...trigger()}
			href={resolve(route)}
			class="row"
			class:current
			class:collapsed
			aria-current={current ? 'page' : undefined}
		>
			<span class="icon"><Icon {icon} size="1.125rem" /></span>
			<span class="label">{label}</span>
		</a>
	{/snippet}
</Tooltip>

<style>
	/* The icon keeps its place whether the column is wide or folded — the
	   same inset from the row's left, under the header's mark — so folding
	   never moves it; the words after it just run out of room. */
	.row {
		position: relative;
		display: flex;
		align-items: center;
		gap: 10px;
		height: var(--nav-item-height);
		padding: 0 10px;
		overflow: hidden;
		border-radius: var(--radius-control);
		color: var(--nav-text);
		font-size: var(--text-base);
		font-weight: 500;
		white-space: nowrap;
		text-decoration: none;
		transition:
			background-color var(--speed-fast),
			color var(--speed-fast);
	}

	/* The icons are a step quieter than the names, so the column reads as
	   words with marks beside them rather than a wall of glyphs. */
	.icon {
		display: inline-flex;
		flex: none;
		align-items: center;
		justify-content: center;
		width: 20px;
		height: 20px;
		color: var(--nav-icon);
		transition: color var(--speed-fast);
	}

	.label {
		flex: 1 1 0;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		transition: opacity var(--speed);
	}

	.row:hover {
		background: var(--nav-hover);
	}

	.row:hover .icon {
		color: var(--nav-text);
	}

	.row:focus-visible {
		outline: 2px solid var(--color-info);
		outline-offset: -2px;
	}

	/* The page being read: filled solid in the brand colour, which no hover
	   comes near, so it needs no other mark. These come after the hover's
	   rules at the same weight, so the pointer never swaps the fill. */
	.row.current {
		background: var(--nav-current);
		color: var(--nav-current-text);
		font-weight: 600;
	}

	.row.current .icon {
		color: var(--nav-current-text);
	}

	/* Folded: the name fades out, and the row closes to a circle as wide as
	   it is tall, centred in the rail under the header's mark. Left at the
	   rail's width the pill would be a squashed oval, 40 by 34. */
	.row.collapsed {
		gap: 0;
		width: var(--nav-item-height);
		margin-inline: auto;
		padding: 0;
		justify-content: center;
	}

	.collapsed .label {
		flex: 0 0 0;
		opacity: 0;
	}

	@media (prefers-reduced-motion: reduce) {
		.row,
		.icon,
		.label {
			transition: none;
		}
	}
</style>
