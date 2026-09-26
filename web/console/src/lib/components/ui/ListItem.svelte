<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = {
		/** Something at the start of the row: usually a Thumb. */
		lead?: Snippet;
		/** The row's title. Left out, the children are the whole content. */
		title?: string;
		/** A line under the title. */
		description?: string;
		/** Things at the end of the row: tags, buttons. */
		end?: Snippet;
		/** Whatever else the row holds, under the title and description. */
		children?: Snippet;
		/** Makes the whole row a link: to the list of what it counts, say.
		    Already resolved; the row only follows it. */
		href?: string;
	};

	let { lead, title, description, end, children, href }: Props = $props();
</script>

{#snippet row()}
	{#if lead}<span class="lead">{@render lead()}</span>{/if}

	<div class="content">
		{#if title}<span class="title">{title}</span>{/if}
		{#if description}<span class="description">{description}</span>{/if}
		{@render children?.()}
	</div>

	{#if end}<span class="end">{@render end()}</span>{/if}
{/snippet}

{#if href}
	<!-- A row that leads somewhere is one link, not a row with a link in it:
	     the whole of it takes the click, and the wash says so. -->
	<li class="linked">
		<!-- The caller resolved the address. -->
		<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
		<a class="item" {href}>{@render row()}</a>
	</li>
{:else}
	<li class="item">{@render row()}</li>
{/if}

<style>
	/* PocketBase's .list-item: at least 54px tall, ruled off above by a shadow
	   rather than a border, so the rule never doubles a box's own. */
	.item {
		display: flex;
		align-items: center;
		gap: 10px;
		min-height: 54px;
		padding: 10px var(--list-inset, var(--space-4));
		overflow-wrap: anywhere;
	}

	.item:not(:first-child),
	.linked:not(:first-child) {
		box-shadow: 0 -1px 0 0 var(--color-secondary);
	}

	a.item {
		color: inherit;
		text-decoration: none;
		transition: background-color var(--speed-fast);
	}

	a.item:hover {
		background: var(--row-hover);
	}

	a.item:focus-visible {
		outline: 2px solid var(--color-info);
		outline-offset: -2px;
	}

	.lead,
	.end {
		display: inline-flex;
		flex-shrink: 0;
		align-items: center;
		gap: 5px;
	}

	.content {
		display: flex;
		flex: 1;
		flex-direction: column;
		min-width: 0;
		line-height: 1.5;
	}

	.title {
		font-size: var(--text-base);
	}

	.description {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	@media (max-width: 34rem) {
		.item {
			flex-wrap: wrap;
		}

		.end {
			margin-left: auto;
		}
	}
</style>
