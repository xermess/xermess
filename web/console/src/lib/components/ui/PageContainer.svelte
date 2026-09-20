<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = {
		/** How wide the content may grow: `sm` for a single form, `md` for a
		    page of settings, `lg` for a page with columns. */
		size?: 'sm' | 'md' | 'lg';
		children: Snippet;
	};

	let { size = 'md', children }: Props = $props();
</script>

<!-- A centred column for a page that reads top to bottom rather than filling
     the width: PocketBase's .wrapper, at its three widths. The gutter stays
     on narrow screens, where the column is the whole width. -->
<div class="container {size}">
	{@render children()}
</div>

<style>
	.container {
		--container-width: 760px;

		box-sizing: border-box;
		width: min(100%, var(--container-width));
		max-width: calc(100vw - (var(--page-gutter) * 2));
		margin-inline: auto;
		padding-inline: var(--page-gutter);
		padding-block: var(--space-2) var(--space-6);
	}

	.sm {
		--container-width: 430px;
	}

	.lg {
		--container-width: 1080px;
	}
</style>
