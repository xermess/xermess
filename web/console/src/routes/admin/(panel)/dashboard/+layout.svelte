<script lang="ts">
	import type { Snippet } from 'svelte';
	import Sidebar from '$lib/components/layout/Sidebar.svelte';

	let { children }: { children: Snippet } = $props();
</script>

<div class="dashboard">
	<Sidebar />

	<div class="content">
		{@render children()}
	</div>
</div>

<style>
	/* The column is as wide as the panel layout says, the same value the
	   header's logo block uses, so the two line up while it folds. */
	.dashboard {
		display: grid;
		grid-template-columns: var(--sidebar-width) 1fr;
		align-items: start;
	}

	/* The frame every page is drawn in: inset by the gutter on every side,
	   and centred once the window is wider than the frame, so a wide monitor
	   gets even margins instead of stretched fields. Pages add no padding of
	   their own — a table, a form and a heading all share this left edge.

	   A page is a stack — heading, toolbar, table — and the frame spaces it,
	   so every page has the same rhythm without saying so. A drawer or a
	   selection bar is portalled or fixed, and takes no part in it. */
	.content {
		grid-column: 2;
		box-sizing: border-box;
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
		width: 100%;
		min-width: 0;
		max-width: calc(var(--content-max-width) + var(--page-gutter) * 2);
		margin-inline: auto;
		padding: var(--space-5) var(--page-gutter) var(--space-6);
	}

	@media (max-width: 55rem) {
		.dashboard {
			grid-template-columns: 1fr;
		}

		.content {
			grid-column: auto;
		}
	}
</style>
