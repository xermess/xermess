<script lang="ts">
	import { RiSearchLine } from 'svelte-remixicon';
	import type { Size } from './control';
	import Icon from './Icon.svelte';

	type Props = {
		value: string;
		/** What is being searched, read out in place of a visible label. */
		label: string;
		placeholder?: string;
		/** `md` in a page's toolbar, `sm` inside a panel. */
		size?: Exclude<Size, 'lg'>;
		/** Given, Enter submits the search as a form; left out, the box is a
		    filter that reacts as it is typed in. */
		onsubmit?: () => void;
		oninput?: (event: Event) => void;
		onkeydown?: (event: KeyboardEvent) => void;
	};

	let {
		value = $bindable(''),
		label,
		placeholder,
		size = 'md',
		onsubmit,
		oninput,
		onkeydown
	}: Props = $props();
</script>

{#snippet field()}
	<Icon icon={RiSearchLine} />
	<input type="search" {placeholder} bind:value {oninput} {onkeydown} aria-label={label} />
{/snippet}

<!-- The search box every list has: the same outlined box as a field
     (styles/fields.css), at a control's height, with the magnifier inside. -->
{#if onsubmit}
	<form
		class="search"
		data-size={size}
		role="search"
		onsubmit={(event) => {
			event.preventDefault();
			onsubmit();
		}}
	>
		{@render field()}
	</form>
{:else}
	<label class="search" data-size={size}>
		{@render field()}
	</label>
{/if}

<style>
	/* It takes the width a toolbar row leaves it, and keeps its height in a
	   column: a basis of auto rather than 0, so growing never squeezes it. */
	.search {
		flex: 1 1 auto;
		display: flex;
		align-items: center;
		gap: var(--space-2);
		min-width: 0;
		height: var(--control-height);
		padding: 0 12px;
		border: 1px solid var(--color-input-border);
		border-radius: var(--radius-sm);
		background: var(--color-input);
		color: var(--color-text-hint);
		transition:
			border-color var(--speed),
			box-shadow var(--speed);
	}

	.search[data-size='sm'] {
		height: var(--control-height-sm);
		padding: 0 10px;
	}

	.search:hover {
		border-color: var(--color-input-border-hover);
	}

	.search:focus-within {
		border-color: var(--color-brand);
		box-shadow: inset 0 0 0 1px var(--color-brand);
	}

	input {
		flex: 1;
		min-width: 0;
		border: none;
		background: transparent;
		color: var(--color-text);
		font: inherit;
		font-size: var(--text-base);
		font-weight: 500;
	}

	.search[data-size='sm'] input {
		font-size: var(--text-sm);
	}

	input:focus {
		outline: none;
	}
</style>
