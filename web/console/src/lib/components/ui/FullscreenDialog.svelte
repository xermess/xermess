<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Dialog } from '@ark-ui/svelte/dialog';
	import { Portal } from '@ark-ui/svelte/portal';
	import { RiCloseLine } from 'svelte-remixicon';
	import Code from './Code.svelte';
	import IconButton from './IconButton.svelte';

	type Props = {
		open: boolean;
		title: string;
		/** A line under the title: what the dialog is for. */
		description?: string;
		/** A tag beside the title, for a record's id and the like. */
		meta?: string;
		/** Given when the body is a form, so the submit button among the
		    actions belongs to it and Enter saves. Without it the dialog is a
		    plain container. */
		onsubmit?: (event: SubmitEvent) => void;
		children: Snippet;
		/** The buttons, at the right of the bar: a cancel and the one that
		    saves. A dialog that only shows something passes none — the bar's
		    close button is how it is left. */
		actions?: Snippet;
		/** A few words before the buttons about why they are as they are —
		    what is unsaved, what is still missing. */
		status?: Snippet;
	};

	let {
		open = $bindable(false),
		title,
		description,
		meta,
		onsubmit,
		children,
		actions,
		status
	}: Props = $props();

	/** Where focus goes when it opens: the sheet itself, so the first thing a
	    keyboard user tabs to is the close button and then the first field, and
	    nothing is ringed before anyone has pressed a key. */
	let sheet = $state<HTMLElement | null>(null);
</script>

<!--
	lazyMount with unmountOnExit rebuilds the sheet on every opening, so it always starts fresh.
	Ark only honours unmountOnExit with a leaving animation to wait for (see styles/ark.css).
-->
<Dialog.Root bind:open lazyMount unmountOnExit initialFocusEl={() => sheet}>
	<Portal>
		<Dialog.Backdrop />
		<Dialog.Positioner class="fullscreen">
			<Dialog.Content class="fullscreen">
				<svelte:element
					this={onsubmit ? 'form' : 'div'}
					class="sheet"
					{onsubmit}
					bind:this={sheet}
					tabindex="-1"
				>
					<header class="bar">
						<IconButton
							icon={RiCloseLine}
							label="Close"
							size="sm"
							placement="bottom"
							onclick={() => (open = false)}
						/>

						<span class="rule" aria-hidden="true"></span>

						<div class="heading">
							<Dialog.Title>{title}</Dialog.Title>
							{#if meta}<Code tone="quiet" truncate title={meta}>{meta}</Code>{/if}
						</div>
					</header>

					<div class="body">
						<div class="column">
							{#if description}
								<Dialog.Description>{description}</Dialog.Description>
							{/if}

							{@render children()}
						</div>
					</div>

					<!-- After the body in the markup, so the keyboard reaches the
					     fields before the button that saves them; the grid draws
					     it at the right of the bar, or along the bottom on a
					     phone. -->
					{#if actions || status}
						<div class="actions">
							{#if status}<span class="status">{@render status()}</span>{/if}
							{@render actions?.()}
						</div>
					{/if}
				</svelte:element>
			</Dialog.Content>
		</Dialog.Positioner>
	</Portal>
</Dialog.Root>

<style>
	/* The bar stays put and only the body scrolls, so the title and the
	   button that saves are in view however long the form is. The bar and the
	   actions share its first row; the body is the rest. */
	.sheet {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		grid-template-rows: auto minmax(0, 1fr);
		min-height: 0;
		height: 100%;
	}

	.sheet:focus {
		outline: none;
	}

	.bar,
	.actions {
		grid-row: 1;
		display: flex;
		align-items: center;
		gap: var(--space-2);
		min-height: var(--header-height);
		border-bottom: 1px solid var(--color-border);
	}

	.bar {
		grid-column: 1 / -1;
		min-width: 0;
		padding: var(--space-1) var(--space-3);
	}

	/* A sheet with no actions lets the title run the whole width; one with
	   them keeps the title clear of the buttons. */
	.sheet:has(> .actions) .bar {
		grid-column: 1;
	}

	.actions {
		grid-column: 2;
		padding: var(--space-1) var(--space-3) var(--space-1) 0;
	}

	.rule {
		flex-shrink: 0;
		width: 1px;
		height: 20px;
		background: var(--color-border);
	}

	.heading {
		display: flex;
		flex: 1;
		align-items: center;
		gap: var(--space-2);
		min-width: 0;
	}

	.heading :global([data-scope='dialog'][data-part='title']) {
		flex-shrink: 0;
		max-width: 100%;
		overflow: hidden;
		font-size: var(--text-lg);
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	/* The record's id or address is the first thing to give way when the bar
	   is short of room. */
	.heading > :global(code) {
		min-width: 0;
	}

	.status {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.body {
		grid-row: 2;
		grid-column: 1 / -1;
		overflow-y: auto;
		/* The room for the scrollbar is kept whether or not there is one, so
		   a form does not shift sideways as it grows past the window. */
		scrollbar-gutter: stable;
		overscroll-behavior: contain;
	}

	/* One readable column, centred: at this width a section's title sits
	   beside its fields (FormSection), and the form is read down the middle
	   of the window rather than along its whole width. */
	.column {
		width: min(var(--dialog-column-width), 100%);
		margin: 0 auto;
		padding: var(--space-5) var(--space-4) var(--space-6);
	}

	.column > :global([data-scope='dialog'][data-part='description']) {
		max-width: 64ch;
		margin: 0 0 var(--space-4);
		font-size: var(--text-base);
		line-height: 1.5;
	}

	/* On a phone the bar keeps the close button and the title, and the
	   actions move to the bottom edge, under the thumb, each an equal share
	   of the width. */
	@media (max-width: 40rem) {
		.sheet {
			grid-template-columns: minmax(0, 1fr);
			grid-template-rows: auto minmax(0, 1fr) auto;
		}

		.actions {
			grid-row: 3;
			grid-column: 1;
			flex-wrap: wrap;
			min-height: 0;
			padding: var(--space-2) var(--space-3);
			border-top: 1px solid var(--color-border);
			border-bottom: none;
			background: var(--color-surface-alt);
		}

		.actions > :global(.control) {
			flex: 1;
		}

		/* A row of its own over the buttons, which share the width below. */
		.status {
			flex-basis: 100%;
			text-align: center;
		}

		.column {
			padding: var(--space-4) var(--space-3) var(--space-5);
		}
	}
</style>
