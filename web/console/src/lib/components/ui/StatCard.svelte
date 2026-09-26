<script lang="ts">
	import type { ComponentType } from 'svelte';
	import { RiArrowRightSLine } from 'svelte-remixicon';
	import Icon from './Icon.svelte';
	import type { ColorPalette } from './control';

	type Props = {
		label: string;
		value: number;
		icon: ComponentType;
		/** What gives the number its context, on one line beside it. A metric
		    is read in a row, so the note never wraps: four cards of one height
		    with their numbers on one line is the whole point. */
		note?: string;
		/** The palette the note is in, for the one that is the exception —
		    a note that says an administrator is locked out, say. */
		tone?: ColorPalette;
		/** Where the number leads, when the administrator may go there. */
		href?: string;
	};

	let { label, value, icon, note, tone = 'neutral', href }: Props = $props();
</script>

<!-- A metric rather than a card: the number is the point, so the label and
     the note are one line each beside it and the whole thing is two rows. -->
<svelte:element this={href ? 'a' : 'div'} class="stat" class:link={href} {href} data-tone={tone}>
	<span class="head">
		<Icon {icon} size="0.9375rem" />
		<span class="label">{label}</span>
		{#if href}
			<span class="go" aria-hidden="true"><Icon icon={RiArrowRightSLine} size="0.8125rem" /></span>
		{/if}
	</span>

	<span class="figure">
		<strong>{value.toLocaleString()}</strong>
		{#if note}<span class="note">{note}</span>{/if}
	</span>
</svelte:element>

<style>
	.stat {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
		padding: var(--space-2) var(--space-3);
		border: 1px solid var(--color-secondary-alt);
		border-radius: var(--radius-sm);
		background: var(--color-surface);
		color: var(--color-text);
		text-decoration: none;
		transition:
			background-color var(--speed),
			border-color var(--speed);
	}

	.link:hover,
	.link:focus-visible {
		outline: 0;
		border-color: var(--color-border);
		background: var(--color-surface-alt);
	}

	/* The label and the icon share a line so the number below is the first
	   thing the card is about; the arrow goes to the far end of it, which is
	   where a link's other end is expected. */
	.head {
		display: flex;
		align-items: center;
		gap: 5px;
		min-width: 0;
		color: var(--color-text-hint);
	}

	.label {
		overflow: hidden;
		font-size: var(--text-sm);
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.go {
		margin-left: auto;
		color: var(--color-text-disabled);
		transition: color var(--speed);
	}

	.link:hover .go {
		color: var(--color-text);
	}

	/* Baseline-aligned, so the numbers of a row sit on one line however many
	   digits each has, and the note takes what is left rather than pushing
	   the number about. */
	.figure {
		display: flex;
		align-items: baseline;
		gap: var(--space-2);
		min-width: 0;
	}

	strong {
		flex: none;
		font-size: var(--text-xl);
		font-weight: 600;
		line-height: 1.25;
		font-variant-numeric: tabular-nums;
	}

	.note {
		flex: 1;
		min-width: 0;
		overflow: hidden;
		color: var(--color-text-hint);
		font-size: var(--text-xs);
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.stat[data-tone='success'] .note {
		color: var(--color-success);
	}

	.stat[data-tone='warning'] .note {
		color: var(--color-warning);
	}

	.stat[data-tone='danger'] .note {
		color: var(--color-danger);
	}

	.stat[data-tone='info'] .note {
		color: var(--color-info);
	}

	@media (max-width: 34rem) {
		.stat {
			padding: var(--space-2);
		}
	}
</style>
