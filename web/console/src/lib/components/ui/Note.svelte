<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = {
		/** A sentence or two, quieter than the content around it. */
		children: Snippet;
	};

	let { children }: Props = $props();
</script>

<!-- The quiet line under a list, a form or a heading: what an empty section
     means, what a setting does, what happens next. It brings no margin — the
     stack it sits in spaces it. -->
<p class="note">{@render children()}</p>

<style>
	.note {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		line-height: 1.5;
	}

	/* A link in a note is part of the sentence: the text's colour, marked by
	   its underline rather than the brand. */
	.note :global(a) {
		color: var(--color-text);
	}
</style>
