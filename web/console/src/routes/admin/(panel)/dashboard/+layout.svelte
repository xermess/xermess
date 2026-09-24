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

	/* No side padding: the tables inside reach the sidebar and the window
	   edge, and everything else is inset by the gutter instead. */
	.content {
		grid-column: 2;
		min-width: 0;
		width: min(100%, var(--content-max-width));
		margin-inline: auto;
		padding: var(--space-4) 0;
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
