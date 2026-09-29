<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = {
		children: Snippet;
		/** How many fields to a row where there is room for them. */
		columns?: 2 | 3;
		/** `compact` for a form in a dialog; `comfortable` puts more air
		    between the columns of a page of settings. */
		spacing?: 'compact' | 'comfortable';
	};

	let { children, columns = 2, spacing = 'compact' }: Props = $props();
</script>

<!-- Fields side by side, each as wide as the others. A field that should be
     read in full — an address, a URL — spans the row: give its wrapper
     class="full". The grid goes to one column when it has no room for two,
     measured on itself rather than the window, so it behaves the same in a
     dialog, a card and a page. -->
<div class="frame">
	<div class="field-grid {spacing}" style:--columns={columns}>
		{@render children()}
	</div>
</div>

<style>
	.frame {
		container-type: inline-size;
	}

	.field-grid {
		display: grid;
		grid-template-columns: repeat(var(--columns), minmax(0, 1fr));
		align-items: start;
		gap: var(--space-3);
	}

	.comfortable {
		column-gap: var(--space-4);
	}

	.field-grid > :global(.full) {
		grid-column: 1 / -1;
	}

	@container (max-width: 30rem) {
		.field-grid {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
