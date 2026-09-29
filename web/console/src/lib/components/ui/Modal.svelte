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

<!-- The other kind of dialog from a FullscreenDialog: a card in the middle
     of the window, for something that is not a record — a question to
     answer, a short choice. It is built when it opens and taken down when it
     has left, as a full-window one is, so it always opens fresh. -->
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
		<Dialog.Positioner>
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
	/* The card itself — centred, bordered, how it arrives — is every
	dialog's, in styles/ark.css. A modal says only how wide it is. */
	:global(.modal-content[data-scope='dialog'][data-part='content']) {
		--dialog-width: 40rem;
	}

	:global(.modal-content[data-scope='dialog'][data-part='content'][data-size='sm']) {
		--dialog-width: 26rem;
	}

	/* A page of settings keeps one height whatever tab is open, so the card
	   does not jump as its contents change; the body scrolls inside it. */
	:global(.modal-content[data-scope='dialog'][data-part='content'][data-size='lg']) {
		--dialog-width: 48rem;

		height: min(42rem, calc(100dvh - 2 * var(--space-4)));
	}

	/* Without a body the header is the whole card above the buttons, so it
	   keeps its padding at the bottom too. */
	header:last-child,
	header:has(+ footer) {
		padding-bottom: var(--space-4);
	}

	header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-3);
		padding: var(--space-3) var(--space-3) var(--space-2) var(--space-4);
	}

	header:focus {
		outline: none;
	}

	header > div {
		min-width: 0;
		padding-top: 6px;
	}

	header :global([data-scope='dialog'][data-part='title']) {
		font-size: var(--text-lg);
		line-height: 1.35;
		overflow-wrap: anywhere;
	}

	/* The body scrolls on its own, so the title and the buttons stay put
	   around a long form. */
	.body {
		flex: 1;
		min-height: 0;
		overflow-y: auto;
		padding: 0 var(--space-4) var(--space-4);
		overscroll-behavior: contain;
	}

	footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-2);
		padding: var(--space-2) var(--space-4);
		border-top: 1px solid var(--color-border);
		background: var(--color-surface-alt);
	}

	/* On a phone the card takes nearly the whole width, and a large one the
	   whole height, so a page of settings is not a letterbox. */
	@media (max-width: 34rem) {
		:global(.modal-content[data-scope='dialog'][data-part='content'][data-size='lg']) {
			height: calc(100dvh - 2 * var(--space-2));
		}

		footer > :global(*) {
			flex: 1;
		}
	}
</style>
