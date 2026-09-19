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
			class="link"
			class:current
			class:collapsed
			aria-current={current ? 'page' : undefined}
		>
			<span class="icon"><Icon {icon} /></span>
			<span class="label">{label}</span>
		</a>
	{/snippet}
</Tooltip>

<style>
	/* The icon keeps its place whether the column is wide or folded — 12px in
	   from the link, 20px from the screen's edge, under the header's mark — so
	   folding never moves it; the words after it just run out of room. The
	   link clips them, and nothing beside the icon can squeeze it. */
	.link {
		position: relative;
		display: flex;
		align-items: center;
		gap: 10px;
		height: 32px;
		padding: 0 12px;
		overflow: hidden;
		border-radius: var(--radius-sm);
		color: var(--color-text);
		font-size: var(--text-base);
		white-space: nowrap;
		text-decoration: none;
		transition:
			background-color var(--speed-fast),
			color var(--speed-fast);
	}

	.link:hover {
		background: var(--color-secondary);
	}

	.link:focus-visible {
		outline: 2px solid var(--color-accent);
		outline-offset: -2px;
	}

	.link.current {
		background: var(--color-secondary-alt);
		font-weight: 600;
	}

	.icon {
		display: inline-flex;
		flex: none;
		width: 16px;
		height: 16px;
		align-items: center;
		justify-content: center;
		color: var(--color-text-hint);
		transition: color var(--speed-fast);
	}

	.link:hover .icon,
	.link.current .icon {
		color: var(--color-text);
	}

	.label {
		flex: 1 1 0;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		transition: opacity var(--speed);
	}

	/* Folded: the name fades out, so the icon has the row to itself. */
	.collapsed .label {
		opacity: 0;
	}
</style>
