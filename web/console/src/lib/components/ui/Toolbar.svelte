<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = {
		/** The search box and the filters, from the left. */
		children: Snippet;
		/** Anything that belongs against the right edge instead. */
		end?: Snippet;
	};

	let { children, end }: Props = $props();
</script>

<!-- The row between a list's heading and its table: search first, filters
     beside it, and anything else pushed to the far end. The search box is
     held to a readable width, so a wide window does not make it a line two
     thousand pixels long; on a narrow one everything wraps. -->
<div class="toolbar">
	<div class="start">{@render children()}</div>
	{#if end}<div class="end">{@render end()}</div>{/if}
</div>

<style>
	.toolbar {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-2);
	}

	.start {
		display: flex;
		flex: 1 1 auto;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-2);
		min-width: 0;
	}

	.start > :global(.search) {
		flex: 1 1 18rem;
		max-width: 28rem;
	}

	.end {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		margin-left: auto;
	}

	@media (max-width: 40rem) {
		.start > :global(.search) {
			max-width: none;
		}
	}
</style>
