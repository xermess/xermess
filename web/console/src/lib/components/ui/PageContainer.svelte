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

<!-- A column for content read top to bottom — a form, a page of settings —
     inside the dashboard's frame. It is held to a readable width but not
     centred: it starts at the frame's left edge, under the page's title, so
     the title and the fields line up and nothing moves between pages. -->
<div class="container {size}">
	{@render children()}
</div>

<style>
	.container {
		--container-width: var(--form-max-width);

		display: flex;
		flex-direction: column;
		gap: var(--space-4);
		width: 100%;
		max-width: var(--container-width);
		min-width: 0;
	}

	.sm {
		--container-width: 430px;
	}

	.lg {
		--container-width: 1080px;
	}
</style>
