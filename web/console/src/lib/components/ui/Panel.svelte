<script lang="ts">
	import type { ComponentType, Snippet } from 'svelte';
	import Icon from './Icon.svelte';

	type Props = {
		title: string;
		icon?: ComponentType;
		/** What sits at the right of the header: labels, or a link to more. */
		meta?: Snippet;
		/** Drops the content's padding, for a list whose rows run edge to edge. */
		flush?: boolean;
		children: Snippet;
	};

	let { title, icon, meta, flush = false, children }: Props = $props();
</script>

<!-- A titled block of a page: PocketBase's open accordion, as its settings
     pages use it — a header strip on the faint tint, then the content on the
     surface, in one bordered box with a soft drop beneath.

     <Panel title="Sessions" icon={RiComputerLine} flush>
       {#snippet meta()}<Tag>2 active</Tag>{/snippet}
       <List>…</List>
     </Panel>
-->
<section class="panel">
	<header>
		{#if icon}<span class="icon"><Icon {icon} size="1.15rem" /></span>{/if}
		<h2>{title}</h2>
		{#if meta}<div class="meta">{@render meta()}</div>{/if}
	</header>

	<div class="content" class:flush>
		{@render children()}
	</div>
</section>

<style>
	.panel {
		display: flex;
		flex-direction: column;
		min-width: 0;
		border: 1px solid var(--color-secondary-alt);
		border-radius: var(--radius-surface);
		background: var(--color-surface-alt);
		box-shadow: var(--shadow-panel);
	}

	header {
		display: flex;
		align-items: center;
		gap: 10px;
		min-height: var(--control-height-lg);
		padding: 10px var(--space-4);
		border-radius: var(--radius-surface) var(--radius-surface) 0 0;
		background: var(--color-surface-alt);
	}

	.icon {
		display: inline-flex;
		color: var(--color-text-hint);
	}

	h2 {
		flex: 1;
		min-width: 0;
		margin: 0;
		overflow: hidden;
		font-size: var(--text-base);
		font-weight: normal;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.meta {
		display: inline-flex;
		flex-shrink: 0;
		align-items: center;
		gap: var(--space-1);
	}

	.content {
		flex: 1;
		padding: var(--space-4);
		border-top: 1px solid var(--color-secondary-alt);
		border-radius: 0 0 var(--radius-surface) var(--radius-surface);
		background: var(--color-surface);
	}

	.content.flush {
		padding: 0;
	}

	@media (max-width: 34rem) {
		header {
			padding: 10px var(--space-3);
		}

		.content:not(.flush) {
			padding: var(--space-3);
		}
	}
</style>
