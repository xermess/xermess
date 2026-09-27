<script lang="ts">
	import Button from './Button.svelte';

	type Props = {
		onclick: () => void;
		/** True while the next page is on its way. */
		loading?: boolean;
		/** How many rows are shown, and how many there are in all when the
		    list knows its total: "50 of 600", or "50 shown" without one. */
		shown?: number;
		total?: number;
	};

	let { onclick, loading = false, shown, total }: Props = $props();
</script>

<div class="more">
	{#if shown !== undefined}
		<span class="count">
			{total === undefined
				? `${shown.toLocaleString()} shown`
				: `${shown.toLocaleString()} of ${total.toLocaleString()}`}
		</span>
	{/if}
	<Button variant="subtle" {loading} {onclick}>Show more</Button>
</div>

<style>
	.more {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-1);
		padding-top: var(--space-1);
	}

	.count {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}
</style>
