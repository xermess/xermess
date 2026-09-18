<script lang="ts">
	import {
		RiAddLine,
		RiArrowDownLine,
		RiArrowUpLine,
		RiCloseLine,
		RiLock2Line
	} from 'svelte-remixicon';
	import type { LoginStep, LoginStepSpec } from '$lib/api';
	import { Icon, IconButton, Tag } from '$lib/components/ui';
	import { markFor } from './steps';

	type Props = {
		/** The steps, in the order they happen. Edited in place. */
		steps: LoginStep[];
		/** Every step there is, which is what can be added. */
		kinds: LoginStepSpec[];
	};

	let { steps = $bindable(), kinds }: Props = $props();

	/** The steps not in the flow yet, in catalog order, which is the order
	    they are offered in. */
	const available = $derived(kinds.filter((kind) => !steps.includes(kind.step)));

	function spec(step: LoginStep): LoginStepSpec | undefined {
		return kinds.find((kind) => kind.step === step);
	}

	function add(step: LoginStep) {
		steps = [...steps, step];
	}

	function remove(at: number) {
		steps = steps.filter((_, index) => index !== at);
	}

	/** Moving a step by one, which is as much reordering as a list of five
	    needs — and works with a keyboard, which dragging does not. */
	function move(at: number, by: -1 | 1) {
		const to = at + by;
		if (to < 1 || to >= steps.length) return;

		const moved = [...steps];
		[moved[at], moved[to]] = [moved[to], moved[at]];
		steps = moved;
	}
</script>

<div class="builder">
	<ol class="steps">
		{#each steps as step, index (step)}
			{@const kind = spec(step)}
			<li class="step" class:planned={kind?.implemented === false}>
				<span class="order">{index + 1}</span>

				<span class="mark"><Icon icon={markFor(step)} size="1rem" /></span>

				<span class="about">
					<span class="title">
						{kind?.label ?? step}
						{#if kind?.fixed}
							<Tag small><Icon icon={RiLock2Line} size="0.7rem" /> always first</Tag>
						{:else if kind?.implemented === false}
							<Tag tone="warning" small>not run yet</Tag>
						{/if}
					</span>
					<span class="description">{kind?.description ?? ''}</span>
				</span>

				{#if !kind?.fixed}
					<span class="controls">
						<IconButton
							icon={RiArrowUpLine}
							label="Move {kind?.label ?? step} earlier"
							size="sm"
							variant="ghost"
							disabled={index <= 1}
							onclick={() => move(index, -1)}
						/>
						<IconButton
							icon={RiArrowDownLine}
							label="Move {kind?.label ?? step} later"
							size="sm"
							variant="ghost"
							disabled={index === steps.length - 1}
							onclick={() => move(index, 1)}
						/>
						<IconButton
							icon={RiCloseLine}
							label="Remove {kind?.label ?? step}"
							size="sm"
							variant="ghost"
							colorPalette="danger"
							onclick={() => remove(index)}
						/>
					</span>
				{/if}
			</li>
		{/each}
	</ol>

	{#if available.length > 0}
		<div class="add">
			{#each available as kind (kind.step)}
				<button type="button" class="offer" onclick={() => add(kind.step)}>
					<Icon icon={RiAddLine} size="0.85rem" />
					{kind.label}
					{#if !kind.implemented}<span class="soon">not run yet</span>{/if}
				</button>
			{/each}
		</div>
	{/if}
</div>

<style>
	.builder {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}

	.steps {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.step {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		padding: var(--space-2);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-sm);
		background: var(--color-surface);
	}

	/* A step the server does not run yet is drawn as the plan it is. */
	.step.planned {
		border-style: dashed;
	}

	/* Where the step comes in the order, which is the whole point of a list
	   that can be rearranged. */
	.order {
		display: grid;
		place-items: center;
		flex-shrink: 0;
		width: 22px;
		height: 22px;
		border-radius: var(--radius-pill);
		background: var(--color-secondary-alt);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		font-variant-numeric: tabular-nums;
	}

	.mark {
		display: grid;
		place-items: center;
		flex-shrink: 0;
		width: 30px;
		height: 30px;
		border: 1px solid var(--color-secondary-alt);
		border-radius: var(--radius-sm);
		background: var(--color-surface-alt);
		color: var(--color-text-hint);
	}

	.about {
		display: flex;
		flex-direction: column;
		min-width: 0;
		line-height: 1.35;
	}

	.title {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	.description {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.controls {
		display: flex;
		align-items: center;
		gap: 2px;
		margin-left: auto;
	}

	.add {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-1);
	}

	.offer {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		padding: 4px 9px;
		border: 1px dashed var(--color-border);
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--color-text-hint);
		font: inherit;
		font-size: var(--text-sm);
		cursor: pointer;
		transition:
			border-color var(--speed-fast),
			color var(--speed-fast);
	}

	.offer:hover {
		border-color: var(--color-text-hint);
		color: var(--color-text);
	}

	.soon {
		color: var(--color-text-disabled);
	}
</style>
