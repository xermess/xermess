<script lang="ts" generics="Value extends string">
	type Props = {
		/** What the group is called, for anyone not looking at it. */
		label: string;
		options: readonly { value: Value; label: string }[];
		/** The option that is on. */
		value: Value;
		onChange: (value: Value) => void;
	};

	let { label, options, value, onChange }: Props = $props();
</script>

<!-- A handful of mutually exclusive filters — "All · Verified · Unverified"
     — as one control at a field's height, so it sits level with the search
     box beside it. The chosen one is raised out of the track. -->
<div class="segmented" role="group" aria-label={label}>
	{#each options as option (option.value)}
		<button
			type="button"
			class:selected={option.value === value}
			aria-pressed={option.value === value}
			onclick={() => onChange(option.value)}
		>
			{option.label}
		</button>
	{/each}
</div>

<style>
	.segmented {
		display: flex;
		flex: none;
		gap: 2px;
		height: var(--control-height);
		padding: 3px;
		border-radius: var(--radius-md);
		background: var(--color-secondary);
	}

	button {
		padding: 0 var(--space-3);
		border: none;
		border-radius: calc(var(--radius-md) - 3px);
		background: transparent;
		color: var(--color-text-hint);
		font: inherit;
		font-size: var(--text-sm);
		font-weight: 500;
		white-space: nowrap;
		cursor: pointer;
		transition:
			background-color var(--speed-fast),
			color var(--speed-fast);
	}

	button:hover {
		color: var(--color-text);
	}

	button.selected {
		background: var(--color-surface);
		box-shadow: var(--shadow-sm);
		color: var(--color-text);
		font-weight: 600;
	}

	@media (max-width: 40rem) {
		.segmented {
			overflow-x: auto;
		}
	}
</style>
