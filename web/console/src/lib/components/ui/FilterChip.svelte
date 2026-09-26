<script lang="ts">
	import type { Snippet } from 'svelte';
	import { RiCloseLine } from 'svelte-remixicon';
	import Icon from './Icon.svelte';

	type Props = {
		/** What clearing it does, as a tooltip: "Clear the role filter". */
		title: string;
		onclear: () => void;
		/** What the list is narrowed to. A `<strong>` in it is set in the
		    monospace face, for the name of a thing. */
		children: Snippet;
	};

	let { title, onclear, children }: Props = $props();
</script>

<!-- A filter that arrived with the address — "role: admin", "only
     ada@example.com" — shown in the toolbar so it is not forgotten, and the
     way back to everything is to click it. -->
<button type="button" class="chip" {title} onclick={onclear}>
	<span>{@render children()}</span>
	<Icon icon={RiCloseLine} />
</button>

<style>
	.chip {
		display: inline-flex;
		flex: none;
		align-items: center;
		gap: var(--space-1);
		height: var(--control-height);
		padding: 0 var(--space-3);
		border: none;
		border-radius: var(--radius-md);
		background: var(--surface-info);
		color: var(--color-text);
		font: inherit;
		font-size: var(--text-sm);
		white-space: nowrap;
		cursor: pointer;
		transition: filter var(--speed-fast);
	}

	.chip:hover {
		filter: brightness(0.96);
	}

	.chip :global(strong) {
		font-family: var(--font-mono);
		font-size: var(--text-sm);
	}
</style>
