<script lang="ts" generics="Value extends string">
	type Props = {
		/** What the group is called, for anyone not looking at it. */
		label: string;
		options: readonly { value: Value; label: string; reset?: boolean }[];
		/** The option that is on. */
		value: Value;
		onChange: (value: Value) => void;
	};

	let { label, options, value, onChange }: Props = $props();
</script>

<!--
	Mutually exclusive filters as one control at field height. The active option is filled with
	the brand colour (shadows vanish in dark mode); "All" is ruled off from the values.
-->
<div class="segmented" role="group" aria-label={label}>
	{#each options as option, index (option.value)}
		<button
			type="button"
			class:selected={option.value === value}
			aria-pressed={option.value === value}
			onclick={() => onChange(option.value)}
		>
			{option.label}
		</button>

		<!-- A line rather than a wider gap: an eye reads a rule as a boundary
		     and a gap as a rounding error. -->
		{#if option.reset && index < options.length - 1}
			<span class="rule" aria-hidden="true"></span>
		{/if}
	{/each}
</div>

<style>
	.segmented {
		display: flex;
		flex: none;
		align-items: center;
		gap: 2px;
		height: var(--control-height);
		padding: 3px;
		border: 1px solid var(--color-input-border);
		border-radius: var(--radius-control);
		background: var(--color-secondary);
	}

	button {
		display: inline-flex;
		align-items: center;
		height: 100%;
		padding: 0 var(--space-3);
		border: none;
		border-radius: var(--radius-control);
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

	/* Under the pointer a segment fills a step further along than the track's
	   own next grey: darker on a light page, lighter on a dark one, so the one
	   rule answers in both. Its label goes to full text colour at the same
	   time, which is the stronger half of that answer. */
	button:hover {
		background: color-mix(in srgb, var(--color-text) 14%, var(--color-secondary));
		color: var(--color-text);
	}

	/* The one that is on, and the pointer on it. Said together and then apart
	   because the hover above would otherwise take the fill away. */
	button.selected,
	button.selected:hover {
		background: var(--color-brand);
		color: var(--color-brand-text);
		font-weight: 600;
	}

	button.selected:hover {
		background: var(--color-brand-hover);
	}

	/* The theme's mid grey, not its border colour: a border is fainter than the
	   track it is drawn on in both themes, and a rule nobody can see rules
	   nothing off. */
	.rule {
		align-self: stretch;
		width: 1px;
		margin: 7px var(--space-1);
		background: var(--color-text-disabled);
	}

	/* The theme's own text, not the brand's: the ring has to be seen both
	   against the track it sits in and against the fill of the segment it may
	   be on, and one of those is the brand. */
	button:focus-visible {
		outline: 2px solid var(--color-text);
		outline-offset: -2px;
	}

	/* A narrow window scrolls the segments sideways, without a scrollbar, as
	   the tabs do: the control is a fixed height, and a scrollbar taking room
	   from it pushes the segments over its bottom edge — a second, upright
	   scrollbar inside it. */
	@media (max-width: 40rem) {
		.segmented {
			overflow-x: auto;
			scrollbar-width: none;
		}

		.segmented::-webkit-scrollbar {
			display: none;
		}
	}
</style>
