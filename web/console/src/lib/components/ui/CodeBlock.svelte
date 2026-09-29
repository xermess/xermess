<script lang="ts">
	import CopyButton from './CopyButton.svelte';
	import { highlighted } from './highlight.svelte';

	type Props = {
		/** The JSON, as it is to be read and copied. */
		code: string;
		/** A name over the code — "ID token", "Example access token". Without
		    one the copy button sits in the block's corner. */
		title?: string;
		/** What is copied, for the button's name. */
		copy?: string;
	};

	let { code, title, copy = title ?? 'code' }: Props = $props();

	/** Until the colours arrive — on the server, and while the highlighter
	    loads — the block is the same text uncoloured, in the same box. */
	const highlight = highlighted(() => code);
</script>

<!-- A block of code in the code face, coloured by the theme's syntax
     tokens (see ./highlight.svelte.ts), with a copy button. -->
<div class="code-block" class:titled={title}>
	{#if title}
		<div class="head">
			<span>{title}</span>
			<CopyButton value={code} label={copy} />
		</div>
	{:else}
		<span class="corner"><CopyButton value={code} label={copy} /></span>
	{/if}

	<div class="body">
		{#if highlight.html}
			<!-- Shiki's own output, from code it escaped itself. -->
			<!-- eslint-disable-next-line svelte/no-at-html-tags -->
			{@html highlight.html}
		{:else}
			<pre class="shiki"><code>{code}</code></pre>
		{/if}
	</div>
</div>

<style>
	.code-block {
		position: relative;
		min-width: 0;
		overflow: hidden;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-surface);
		background: var(--color-surface-alt);
	}

	.head {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-2);
		padding: 4px 4px 4px var(--space-3);
		border-bottom: 1px solid var(--color-border);
		background: var(--color-secondary);
		font-size: var(--text-sm);
		font-weight: 600;
	}

	.corner {
		position: absolute;
		top: 6px;
		right: 6px;
		z-index: 1;
	}

	.body {
		overflow: auto;
	}

	.body :global(pre.shiki) {
		padding: var(--space-3) var(--space-4);
		line-height: 1.6;
	}

	/* The copy button floats over the first line; the text keeps clear of it. */
	.code-block:not(.titled) .body :global(pre.shiki) {
		padding-right: calc(var(--control-height) + var(--space-2));
	}
</style>
