<script lang="ts">
	import type { Snippet } from 'svelte';
	import { RiDeleteBinLine } from 'svelte-remixicon';
	import Button from './Button.svelte';
	import Icon from './Icon.svelte';
	import Note from './Note.svelte';

	type Props = {
		/** What the button does, as a heading: "Remove this language". */
		title: string;
		/** What is lost by doing it. */
		description: string;
		/** Asks before anything is done (a ConfirmDialog). */
		onclick?: () => void;
		/** The button's word. */
		label?: string;
		/** In place of the button, for a zone that asks its question inline. */
		children?: Snippet;
	};

	let { title, description, onclick, label = 'Remove', children }: Props = $props();
</script>

<!-- The one thing on a form that cannot be taken back, kept apart from the
     button that saves: at the foot of the form, ruled in the danger colour,
     saying what it costs beside the button that does it. -->
<section class="danger-zone">
	<div class="text">
		<strong>{title}</strong>
		<Note>{description}</Note>
	</div>

	<div class="action">
		{#if children}
			{@render children()}
		{:else}
			<Button colorPalette="danger" variant="subtle" size="sm" {onclick}>
				<Icon icon={RiDeleteBinLine} />
				{label}
			</Button>
		{/if}
	</div>
</section>

<style>
	.danger-zone {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: var(--space-3);
		margin-top: var(--space-6);
		padding: var(--space-3) var(--space-4);
		border: 1px solid color-mix(in srgb, var(--color-danger) 40%, transparent);
		border-radius: var(--radius-surface);
	}

	.text {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: 2px;
		min-width: 12rem;
	}

	.action {
		display: flex;
		flex-shrink: 0;
		gap: var(--space-2);
	}
</style>
