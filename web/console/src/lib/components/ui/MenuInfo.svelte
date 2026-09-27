<script lang="ts">
	import type { ComponentType, Snippet } from 'svelte';
	import Icon from './Icon.svelte';

	type Props = {
		/** The icon before the words, unless `lead` draws something else. */
		icon?: ComponentType;
		/** Drawn in place of the icon: an avatar, say. */
		lead?: Snippet;
		label: string;
		description?: string;
		/** Anything at the far end: a key, a tag. */
		end?: Snippet;
	};

	let { icon, lead, label, description, end }: Props = $props();
</script>

<!-- A row that tells rather than does — who is signed in, a shortcut — laid
     out as the rows around it are, so the menu reads as one list, but with no
     hover and no place in the arrow keys' order. -->
<div class="info">
	{#if lead}{@render lead()}{:else if icon}<Icon {icon} />{/if}
	<span class="text">
		<span class="label" title={label}>{label}</span>
		{#if description}<small>{description}</small>{/if}
	</span>
	{#if end}<span class="end">{@render end()}</span>{/if}
</div>

<style>
	.info {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		min-height: 36px;
		padding: 8px 10px;
		color: var(--color-text);
		font-size: var(--text-base);
		line-height: 1.3;
	}

	.info > :global(svg) {
		flex: none;
		color: var(--color-text-hint);
	}

	.text {
		display: flex;
		flex: 1;
		flex-direction: column;
		min-width: 0;
	}

	.label {
		overflow: hidden;
		font-weight: 500;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	small {
		overflow: hidden;
		color: var(--color-text-hint);
		text-overflow: ellipsis;
		white-space: nowrap;
		font-size: var(--text-xs);
	}

	.end {
		display: flex;
		flex: none;
		align-items: center;
		gap: 4px;
		margin-left: auto;
	}
</style>
