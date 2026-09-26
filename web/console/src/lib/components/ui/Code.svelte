<script lang="ts">
	import type { Snippet } from 'svelte';

	type Props = {
		/** An identifier, a key, a code: something read character by character. */
		children: Snippet;
		/** `quiet` sets it in the hint colour, for a value beside the name it
		    belongs to rather than the value a column is about. */
		tone?: 'default' | 'quiet';
		/** Cut a long value to its column with an ellipsis; the title still
		    has it in full. */
		truncate?: boolean;
		title?: string;
	};

	let { children, tone = 'default', truncate = false, title }: Props = $props();
</script>

<!-- A machine value set in the monospace face on a quiet tint, so an id, a
     client_id or a language code reads as a value and not as prose. -->
<code class={tone} class:truncate {title}>{@render children()}</code>

<style>
	code {
		display: inline-flex;
		align-items: center;
		max-width: 100%;
		min-height: 24px;
		padding: 0 7px;
		border-radius: var(--radius-sm);
		background: var(--color-secondary-alt);
		color: var(--color-text);
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		vertical-align: middle;
		white-space: nowrap;
	}

	.quiet {
		color: var(--color-text-hint);
	}

	.truncate {
		display: inline-block;
		overflow: hidden;
		line-height: 24px;
		text-overflow: ellipsis;
	}
</style>
