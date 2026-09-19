<script lang="ts">
	import type { Snippet } from 'svelte';

	/** A step of the trail: its name, and where it leads when it is a page of
	    its own — "APIs" above one API — so the trail is the way back. */
	type Crumb = string | { label: string; href: string };

	type Props = {
		/** Where the page sits, outermost first; the last is the page itself. */
		crumbs: Crumb[];
		/** Small controls right beside the title, such as a refresh button. */
		secondary?: Snippet;
		/** The page's main actions, at the far end. */
		actions?: Snippet;
	};

	let { crumbs, secondary, actions }: Props = $props();

	const label = (crumb: Crumb) => (typeof crumb === 'string' ? crumb : crumb.label);
</script>

<!-- PocketBase's page header: breadcrumbs, the secondary buttons right beside
     them, and the primary ones pushed to the far side. The last crumb is the
     page's title, so it is the heading. -->
<header class="page-header">
	<nav class="breadcrumbs" aria-label="Breadcrumb">
		{#each crumbs.slice(0, -1) as crumb, index (index)}
			{#if typeof crumb === 'string'}
				<span class="crumb">{crumb}</span>
			{:else}
				<!-- The page resolved the address: it knows the route. -->
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
				<a class="crumb link" href={crumb.href}>{crumb.label}</a>
			{/if}
		{/each}
		<h1 class="crumb current" aria-current="page">{label(crumbs.at(-1) ?? '')}</h1>
	</nav>

	{#if secondary}<div class="secondary">{@render secondary()}</div>{/if}
	{#if actions}<div class="actions">{@render actions()}</div>{/if}
</header>

<style>
	.page-header {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 10px var(--space-4);
		min-height: var(--control-height);
	}

	.breadcrumbs {
		display: inline-flex;
		align-items: center;
		gap: 30px;
		min-width: 0;
		color: var(--color-text-hint);
		font-size: 1.286rem;
	}

	.crumb {
		position: relative;
		margin: 0;
		font-size: inherit;
		font-weight: normal;
		white-space: nowrap;
	}

	.crumb:not(.current)::after {
		position: absolute;
		top: 0;
		right: -18px;
		height: 100%;
		align-content: center;
		color: var(--color-text-disabled);
		font-size: 0.85em;
		content: '/';
		pointer-events: none;
	}

	.link {
		color: inherit;
		text-decoration: none;
		transition: color var(--speed-fast);
	}

	.link:hover {
		color: var(--color-text);
	}

	.current {
		overflow: hidden;
		color: var(--color-text);
		text-overflow: ellipsis;
	}

	.secondary {
		display: inline-flex;
		align-items: center;
		gap: 10px;
	}

	/* The count a list page puts beside its name — "42 total" — is quieter
	   than the name. The page writes the span; how it looks is decided here,
	   so every list says it the same way. */
	.secondary :global(.total) {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.actions {
		display: inline-flex;
		align-items: center;
		gap: 10px;
		margin-left: auto;
	}

	/* A narrow screen has room for the page's own name only. */
	@media (max-width: 34rem) {
		.crumb:not(.current) {
			display: none;
		}
	}
</style>
