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
		/** A page under a branch rather than one of the column's own: it sits
		    against the tree's rule, and marks that rule when it is the page
		    being read. */
		nested?: boolean;
	};

	let { route, label, icon, current, collapsed, nested = false }: Props = $props();
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
			class:nested
			aria-current={current ? 'page' : undefined}
		>
			<span class="icon"><Icon {icon} /></span>
			<span class="label">{label}</span>
		</a>
	{/snippet}
</Tooltip>

<style>
	/* The icon keeps its place whether the column is wide or folded — 12px in
	   from the row, under the header's mark — so folding never moves it; the
	   words after it just run out of room. The row clips them, and nothing
	   beside the icon can squeeze it. */
	.row {
		position: relative;
		display: flex;
		align-items: center;
		gap: var(--space-2);
		height: var(--nav-row-height);
		padding: 0 12px;
		overflow: hidden;
		border-radius: var(--radius-md);
		color: var(--color-text-hint);
		font-size: var(--text-base);
		white-space: nowrap;
		text-decoration: none;
		transition:
			background-color var(--speed-fast),
			color var(--speed-fast);
	}

	.row:hover {
		background: var(--nav-hover);
		color: var(--color-text);
	}

	.row:focus-visible {
		outline: 2px solid var(--color-accent);
		outline-offset: -2px;
	}

	/* The page being read: filled, named in full, and marked down its left
	   edge. The fill alone was easy to miss beside a hover; the mark is what
	   the eye finds first when it comes back to the column. */
	.row.current {
		background: var(--nav-current);
		color: var(--color-text);
		font-weight: 600;
	}

	.row.current::before {
		content: '';
		position: absolute;
		top: 50%;
		left: 0;
		width: 3px;
		height: 18px;
		border-radius: 0 var(--radius-pill) var(--radius-pill) 0;
		background: var(--nav-mark);
		transform: translateY(-50%);
	}

	.icon {
		display: inline-flex;
		flex: none;
		width: 16px;
		height: 16px;
		align-items: center;
		justify-content: center;
		color: inherit;
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

	/* A page under a branch. The rule to its left already says where it
	   belongs, so the row carries no mark of its own — it colours that rule
	   instead, which is what makes the tree read as a tree.

	   Its mark sits outside the row, on the rule, so the row may not clip:
	   the label clips itself, and a nested row is never folded, which is the
	   only thing the clip was for. */
	.row.nested {
		height: calc(var(--nav-row-height) - 2px);
		padding-left: 10px;
		overflow: visible;
	}

	.row.nested.current::before {
		left: calc(var(--nav-branch-inset) * -1);
		height: calc(100% - 8px);
		border-radius: var(--radius-pill);
	}

	.nested .icon {
		width: 15px;
		height: 15px;
	}
</style>
