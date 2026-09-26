<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = {
		/** The key, as it is printed on the keyboard: "/", "↵", "esc". */
		children: Snippet;
	};

	let { children }: Props = $props();
</script>

<!-- A key on the keyboard, drawn as a small cap, wherever a shortcut is
     written down: the search's footer, the help menu. -->
<kbd>{@render children()}</kbd>

<style>
	kbd {
		display: inline-grid;
		place-items: center;
		min-width: 20px;
		height: 20px;
		padding: 0 5px;
		border: 1px solid var(--color-border);
		border-radius: 4px;
		background: var(--color-surface-alt);
		color: var(--color-text-hint);
		font-family: var(--font-mono);
		font-size: var(--text-xs);
		line-height: 1;
	}
</style>
