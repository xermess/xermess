<script lang="ts">
	import type { ComponentType, Snippet } from 'svelte';
	import { Icon } from '$lib/components/ui';

	type Props = {
		title: string;
		/** What the fields are for, and where they show. */
		description: string;
		icon: ComponentType;
		/** Labels under the description: where the values are published. */
		meta?: Snippet;
		children: Snippet;
	};

	let { title, description, icon, meta, children }: Props = $props();
</script>

<!-- One group of the organisation's settings: what it is on the left, the
     fields on the right, so the page reads as a list of subjects rather than
     a column of boxes. Narrow, the words go above the fields. The layout is
     measured on the section's own width, as FieldGrid's is. -->
<section class="frame">
	<div class="section">
		<header>
			<span class="icon" aria-hidden="true"><Icon {icon} size="1.1rem" /></span>
			<div class="text">
				<h2>{title}</h2>
				<p>{description}</p>
				{#if meta}<div class="meta">{@render meta()}</div>{/if}
			</div>
		</header>

		<div class="card">
			{@render children()}
		</div>
	</div>
</section>

<style>
	.frame {
		container-type: inline-size;
	}

	.section {
		display: grid;
		grid-template-columns: minmax(0, 17rem) minmax(0, 1fr);
		align-items: start;
		gap: var(--space-5);
	}

	header {
		display: flex;
		align-items: flex-start;
		gap: var(--space-3);
		padding-top: var(--space-1);
	}

	/* The same mark the panels' headers carry, boxed so the column of
	   sections has a line of icons down its left edge. */
	.icon {
		display: inline-flex;
		flex: none;
		align-items: center;
		justify-content: center;
		width: 34px;
		height: 34px;
		border: 1px solid var(--color-secondary-alt);
		border-radius: var(--radius-sm);
		background: var(--color-surface-alt);
		color: var(--color-text-hint);
	}

	.text {
		min-width: 0;
		padding-top: 1px;
	}

	h2 {
		margin: 0;
		font-size: var(--text-base);
		font-weight: 600;
		line-height: 1.4;
	}

	p {
		margin: 2px 0 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		line-height: 1.5;
	}

	.meta {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-1);
		margin-top: var(--space-2);
	}

	.card {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		min-width: 0;
		padding: var(--space-4);
		border: 1px solid var(--color-secondary-alt);
		border-radius: var(--radius-sm);
		background: var(--color-surface);
		box-shadow: var(--shadow-panel);
	}

	@container (max-width: 44rem) {
		.section {
			grid-template-columns: minmax(0, 1fr);
			gap: var(--space-3);
		}

		.card {
			padding: var(--space-3);
		}
	}
</style>
