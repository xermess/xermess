<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Dialog } from '@ark-ui/svelte/dialog';
	import { Portal } from '@ark-ui/svelte/portal';
	import { RiCloseLine } from 'svelte-remixicon';
	import IconButton from './IconButton.svelte';

	type Props = {
		open?: boolean;
		title: string;
		description?: string;
		/** `sm` for a question, `md` for a short form, `lg` for a page of
		    settings. */
		size?: 'sm' | 'md' | 'lg';
		/** Whether the reader may dismiss it — false while something it
		    started is still under way. */
		closable?: boolean;
		/** The body. A question needs none: its title and description say it. */
		children?: Snippet;
		/** The buttons, against the bottom right. */
		footer?: Snippet;
	};

	let {
		open = $bindable(false),
		title,
		description,
		size = 'md',
		closable = true,
		children,
		footer
	}: Props = $props();

	/** Where focus lands when it opens: the heading, which a screen reader
	    reads first, rather than the close button, whose tooltip would pop up
	    over the title before anything was done. */
	let heading = $state<HTMLElement | null>(null);
</script>

<!-- The other kind of dialog from a drawer: a card in the middle of the
     window, for something that is not a record — a question to answer, the
     account's own settings. It is built when it opens and taken down when it
     has left, as a drawer is, so it always opens fresh. -->
<Dialog.Root
	{open}
	onOpenChange={(details) => {
		if (closable || details.open) open = details.open;
	}}
	lazyMount
	unmountOnExit
	initialFocusEl={() => heading}
>
	<Portal>
		<Dialog.Backdrop />
		<Dialog.Positioner class="modal-positioner">
			<Dialog.Content class="modal-content" data-size={size}>
				<header bind:this={heading} tabindex="-1">
					<div>
						<Dialog.Title>{title}</Dialog.Title>
						{#if description}<Dialog.Description>{description}</Dialog.Description>{/if}
					</div>

					{#if closable}
						<IconButton
							icon={RiCloseLine}
							label="Close"
							size="sm"
							placement="left"
							onclick={() => (open = false)}
						/>
					{/if}
				</header>

				{#if children}<div class="body">{@render children()}</div>{/if}

				{#if footer}<footer>{@render footer()}</footer>{/if}
			</Dialog.Content>
		</Dialog.Positioner>
	</Portal>
</Dialog.Root>

<style>
	/* The dialog's anatomy is styled once in styles/ark.css, for the drawer.
	   These are one selector more specific, and centre it instead — in both
	   directions, as ConfirmDialog is, so every dialog that is not a drawer
	   opens in the same place. */
	:global(.modal-positioner[data-scope='dialog'][data-part='positioner']) {
		justify-content: center;
		align-items: center;
		padding: var(--space-4);
		overflow-y: auto;
	}

	:global(.modal-content[data-scope='dialog'][data-part='content']) {
		--modal-width: 40rem;

		width: min(var(--modal-width), 100%);
		height: auto;
		/* In the middle while it fits, and scrolled from its top when it does
		   not, rather than cut off above the window. */
		margin: auto;
		max-height: calc(100dvh - 2 * var(--space-4));
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-md);
		overflow: hidden;
	}

	:global(.modal-content[data-scope='dialog'][data-part='content'][data-size='sm']) {
		--modal-width: 26rem;
	}

	/* A page of settings keeps one height whatever tab is open, so the card
	   does not jump as its contents change; the body scrolls inside it. */
	:global(.modal-content[data-scope='dialog'][data-part='content'][data-size='lg']) {
		--modal-width: 48rem;

		height: min(42rem, calc(100dvh - 2 * var(--space-4)));
	}

	/* It arrives a shade small and settles, rather than sliding in from the
	   side the way a drawer does. Two names, because Ark only waits for a
	   leaving animation whose name changes (the note in ark.css). */
	:global(.modal-content[data-scope='dialog'][data-part='content'][data-state='open']) {
		animation: modal-in var(--speed) ease-out;
	}

	:global(.modal-content[data-scope='dialog'][data-part='content'][data-state='closed']) {
		animation: modal-out var(--speed) ease-out forwards;
	}

	@keyframes -global-modal-in {
		from {
			opacity: 0;
			transform: scale(0.98) translateY(-4px);
		}
	}

	@keyframes -global-modal-out {
		to {
			opacity: 0;
			transform: scale(0.98) translateY(-4px);
		}
	}

	/* Without a body the header is the whole card above the buttons, so it
	   keeps its padding at the bottom too. */
	header:last-child,
	header:has(+ footer) {
		padding-bottom: var(--space-5);
	}

	header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-3);
		padding: var(--space-4) var(--space-4) var(--space-3) var(--space-5);
	}

	header:focus {
		outline: none;
	}

	header > div {
		padding-top: 4px;
	}

	/* The body scrolls on its own, so the title and the buttons stay put
	   around a long form. */
	.body {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		padding: 0 var(--space-5) var(--space-5);
		overscroll-behavior: contain;
	}

	footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-2);
		padding: var(--space-3) var(--space-5);
		border-top: 1px solid var(--color-border);
		background: var(--color-surface-alt);
	}

	@media (prefers-reduced-motion: reduce) {
		:global(.modal-content[data-scope='dialog'][data-part='content'][data-state='open']),
		:global(.modal-content[data-scope='dialog'][data-part='content'][data-state='closed']) {
			animation-duration: 1ms;
		}
	}
</style>
