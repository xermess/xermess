<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = {
		/** A box of its own: bordered and rounded, with rows inset to match.
		    Leave it off for a list inside a flush Panel, whose box it already
		    is. */
		bordered?: boolean;
		/** What the list is, for a screen reader. */
		label?: string;
		children: Snippet;
	};

	let { bordered = false, label, children }: Props = $props();
</script>

<!-- PocketBase's .list: rows one under another, each ruled off from the one
     above. Fill it with ListItem. -->
<ul class="list" class:bordered aria-label={label}>
	{@render children()}
</ul>

<style>
	.list {
		--list-inset: var(--space-4);

		display: block;
		width: 100%;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.bordered {
		--list-inset: 13px;

		overflow: hidden;
		border: 1px solid var(--color-secondary-alt);
		border-radius: var(--radius-surface);
		background: var(--color-surface);
	}

	@media (max-width: 34rem) {
		.list:not(.bordered) {
			--list-inset: var(--space-3);
		}
	}
</style>
