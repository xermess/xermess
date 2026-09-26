<script lang="ts">
	import type { Snippet } from 'svelte';

	/** A step of the trail: its name, and where it leads when it is a page of
	    its own — "APIs" above one API — so the trail is the way back. */
	type Crumb = string | { label: string; href: string };

	type Props = {
		/** Where the page sits, outermost first; the last is the page itself. */
		crumbs: Crumb[];
		/** A sentence under the title saying what the page is for. */
		description?: string;
		/** How many records the page lists, shown beside the title. */
		count?: number;
		/** Tags beside the title that describe the record itself — "default",
		    "unsaved", an algorithm — rather than anything that can be done. */
		meta?: Snippet;
		/** Small controls — refresh, settings — grouped before the actions. */
		secondary?: Snippet;
		/** The page's main actions, at the far end of the title row. */
		actions?: Snippet;
	};

	let { crumbs, description, count, meta, secondary, actions }: Props = $props();

	const label = (crumb: Crumb) => (typeof crumb === 'string' ? crumb : crumb.label);

	const trail = $derived(crumbs.slice(0, -1));
</script>

<!-- Every page's heading, in the order people read one: where they are, what
     this is, what it is for. The trail is small and above, the title is the
     page's h1, and everything that can be done sits on the title's row at the
     far end — the icon buttons first, then the one action that matters. -->
<header class="page-header">
	{#if trail.length > 0}
		<nav class="trail" aria-label="Breadcrumb">
			{#each trail as crumb, index (index)}
				{#if index > 0}<span class="separator" aria-hidden="true">/</span>{/if}
				{#if typeof crumb === 'string'}
					<span class="crumb">{crumb}</span>
				{:else}
					<!-- The page resolved the address: it knows the route. -->
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
					<a class="crumb link" href={crumb.href}>{crumb.label}</a>
				{/if}
			{/each}
		</nav>
	{/if}

	<div class="title-row">
		<div class="title">
			<h1>{label(crumbs.at(-1) ?? '')}</h1>
			{#if count !== undefined}
				<span class="count" title="{count} in total">{count.toLocaleString()}</span>
			{/if}
			{#if meta}<div class="meta">{@render meta()}</div>{/if}
		</div>

		{#if secondary || actions}
			<div class="tools">
				{#if secondary}<div class="secondary">{@render secondary()}</div>{/if}
				{#if actions}<div class="actions">{@render actions()}</div>{/if}
			</div>
		{/if}
	</div>

	{#if description}
		<p class="description">{description}</p>
	{/if}
</header>

<style>
	.page-header {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
		min-width: 0;
	}

	.trail {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 6px;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.separator {
		color: var(--color-text-disabled);
	}

	.link {
		color: inherit;
		text-decoration: none;
		transition: color var(--speed-fast);
	}

	.link:hover {
		color: var(--color-text);
	}

	/* The title and the tools share a row, centred on each other, and the
	   tools wrap under the title rather than squeezing it. */
	.title-row {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-2) var(--space-4);
		min-height: var(--control-height);
	}

	.title {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		min-width: 0;
	}

	h1 {
		overflow: hidden;
		margin: 0;
		color: var(--color-text);
		font-size: var(--text-2xl);
		font-weight: 600;
		letter-spacing: -0.01em;
		line-height: 1.2;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.count {
		padding: 2px 8px;
		border-radius: var(--radius-pill);
		background: var(--color-secondary);
		color: var(--color-text-hint);
		font-size: var(--text-xs);
		font-variant-numeric: tabular-nums;
		font-weight: 600;
		line-height: 1.5;
	}

	.meta {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-1);
	}

	.tools,
	.secondary,
	.actions {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	/* A quiet rule between the icon buttons and the actions, so the two kinds
	   read as two groups. */
	.secondary + .actions {
		padding-left: var(--space-2);
		border-left: 1px solid var(--color-border);
	}

	/* The count a page used to put among its buttons — "42 total" — still
	   reads quietly if a page puts one there. */
	.secondary :global(.total) {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		white-space: nowrap;
	}

	.description {
		max-width: 75ch;
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-base);
		line-height: 1.55;
	}

	@media (max-width: 34rem) {
		h1 {
			font-size: var(--text-xl);
		}

		.tools {
			width: 100%;
		}

		.actions {
			margin-left: auto;
		}
	}
</style>
