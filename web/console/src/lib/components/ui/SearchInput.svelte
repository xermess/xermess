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
		id?: string;
		name?: string;
		oninput?: (event: Event) => void;
		onkeydown?: (event: KeyboardEvent) => void;
	};

	let {
		value = $bindable(''),
		label,
		placeholder,
		size = 'md',
		onsubmit,
		id,
		name,
		oninput,
		onkeydown
	}: Props = $props();

	/** Search boxes are fields too, even when they sit outside a form. */
	const uid = $props.id();
</script>

{#snippet field()}
	<Icon icon={RiSearchLine} />
	<input
		id={id ?? uid}
		{name}
		type="search"
		{placeholder}
		bind:value
		{oninput}
		{onkeydown}
		aria-label={label}
	/>
{/snippet}

<!-- The search box every list has: the same outlined box as a field — it
     wears .field-box, so its border, hover and focus are the field's
     (styles/fields.css) — at a control's height, with the magnifier inside. -->
{#if onsubmit}
	<form
		class="search field-box"
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
	<label class="search field-box" data-size={size}>
		{@render field()}
	</label>
{/if}

<style>
	/* It takes the width a toolbar row leaves it, and holds its height in a
	   column: a basis of auto, so growing never squeezes it, and a max
	   height, so a tall column cannot stretch it either. */
	.search {
		--search-height: var(--control-height);

		flex: 1 1 auto;
		display: flex;
		align-items: center;
		gap: var(--space-2);
		min-width: 0;
		height: var(--search-height);
		max-height: var(--search-height);
		padding: 0 12px;
		color: var(--color-text-hint);
	}

	.search[data-size='sm'] {
		--search-height: var(--control-height-sm);

		padding: 0 10px;
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
