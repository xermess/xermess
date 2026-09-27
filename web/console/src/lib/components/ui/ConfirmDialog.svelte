<script lang="ts">
	import type { ComponentType, Snippet } from 'svelte';
	import { Dialog } from '@ark-ui/svelte/dialog';
	import { Portal } from '@ark-ui/svelte/portal';
	import {
		RiAlertLine,
		RiErrorWarningLine,
		RiInformationLine,
		RiQuestionLine
	} from 'svelte-remixicon';
	import Button from './Button.svelte';
	import Icon from './Icon.svelte';

	type Tone = 'danger' | 'warning' | 'info' | 'neutral';

	type Props = {
		open?: boolean;
		/** The question, as a question: "Delete 3 users?". */
		title: string;
		/** What saying yes does, and what cannot be undone. */
		description?: string;
		/** `danger` for what cannot be taken back, `warning` for what can
		    but hurts, `info` or `neutral` for a plain question. It colours the
		    icon and the button that says yes. */
		tone?: Tone;
		/** Drawn in the badge; each tone has one of its own. */
		icon?: ComponentType;
		/** What the button that says yes is called: the action, never "OK". */
		confirmLabel: string;
		cancelLabel?: string;
		/** True while what was confirmed is under way: the yes button spins,
		    and the dialog cannot be dismissed until it is done. */
		busy?: boolean;
		onConfirm: () => void;
		/** What does not fit in a sentence: the records it touches, say. */
		children?: Snippet;
	};

	let {
		open = $bindable(false),
		title,
		description,
		tone = 'danger',
		icon,
		confirmLabel,
		cancelLabel = 'Cancel',
		busy = false,
		onConfirm,
		children
	}: Props = $props();

	const icons: Record<Tone, ComponentType> = {
		danger: RiErrorWarningLine,
		warning: RiAlertLine,
		info: RiInformationLine,
		neutral: RiQuestionLine
	};

	/** Focus starts on the way out, not the way in: Enter on a dialog that
	    has only just appeared should not do the thing it asks about. */
	let actions = $state<HTMLElement | null>(null);
</script>

<!-- A question in the middle of the window, asked before something that is
     hard to take back. It is an alert dialog, so a screen reader reads it out
     as it opens, and it sits above a drawer, since that is where most of these
     questions are asked from.

     The card is the popovers' sheet — a select's list, a menu — grown to hold
     a sentence: the same hairline border, the same shadow, on the same
     surface, so a question reads as part of the panel rather than the
     browser's. -->
<Dialog.Root
	{open}
	role="alertdialog"
	onOpenChange={(details) => {
		if (!busy || details.open) open = details.open;
	}}
	closeOnInteractOutside={!busy}
	closeOnEscape={!busy}
	lazyMount
	unmountOnExit
	initialFocusEl={() => actions?.querySelector('button') ?? null}
>
	<Portal>
		<Dialog.Backdrop class="confirm-backdrop" />
		<Dialog.Positioner class="confirm-positioner">
			<Dialog.Content class="confirm-content" data-palette={tone}>
				<div class="head">
					<span class="badge" aria-hidden="true">
						<Icon icon={icon ?? icons[tone]} size="20" />
					</span>

					<div class="text">
						<Dialog.Title>{title}</Dialog.Title>
						{#if description}<Dialog.Description>{description}</Dialog.Description>{/if}
						{#if children}<div class="body">{@render children()}</div>{/if}
					</div>
				</div>

				<div class="actions" bind:this={actions}>
					<Button variant="subtle" disabled={busy} onclick={() => (open = false)}>
						{cancelLabel}
					</Button>
					<Button colorPalette={tone} loading={busy} disabled={busy} onclick={onConfirm}>
						{confirmLabel}
					</Button>
				</div>
			</Dialog.Content>
		</Dialog.Positioner>
	</Portal>
</Dialog.Root>

<style>
	/* The dialog's anatomy is styled once in styles/ark.css, for the drawer.
	   These are one selector more specific: above a drawer, and centred. */
	:global(.confirm-backdrop[data-scope='dialog'][data-part='backdrop']) {
		z-index: 60;
	}

	:global(.confirm-positioner[data-scope='dialog'][data-part='positioner']) {
		z-index: 65;
		justify-content: center;
		align-items: center;
		padding: var(--space-4);
		overflow-y: auto;
	}

	:global(.confirm-content[data-scope='dialog'][data-part='content']) {
		width: min(28rem, 100%);
		height: auto;
		/* In the middle while it fits, and scrolled from its top when it
		   does not, rather than cut off above the window. */
		margin: auto;
		padding: var(--space-5);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		background: var(--color-surface);
		box-shadow: var(--shadow-md);
	}

	/* It arrives a shade small and settles, the way a select's list opens.
	   Two names, because Ark only waits for a leaving animation whose name
	   changes (the note in ark.css). */
	:global(.confirm-content[data-scope='dialog'][data-part='content'][data-state='open']) {
		animation: confirm-in var(--speed) ease-out;
	}

	:global(.confirm-content[data-scope='dialog'][data-part='content'][data-state='closed']) {
		animation: confirm-out var(--speed) ease-in forwards;
	}

	@keyframes -global-confirm-in {
		from {
			transform: scale(0.98) translateY(-4px);
		}
	}

	@keyframes -global-confirm-out {
		to {
			transform: scale(0.98) translateY(-4px);
		}
	}

	.head {
		display: flex;
		align-items: flex-start;
		gap: var(--space-4);
	}

	/* The tone, said once: a circle in the palette the yes button is in,
	   the way a menu's danger row is tinted. */
	.badge {
		display: inline-flex;
		flex: none;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		border: 1px solid var(--palette-border);
		border-radius: 50%;
		background: var(--palette-subtle);
		color: var(--palette-fg);
	}

	.text {
		flex: 1;
		min-width: 0;
		padding-top: 2px;
	}

	.text :global([data-scope='dialog'][data-part='title']) {
		font-size: var(--text-lg);
		line-height: 1.35;
		overflow-wrap: anywhere;
	}

	.text :global([data-scope='dialog'][data-part='description']) {
		margin-top: var(--space-1);
		font-size: var(--text-sm);
		line-height: 1.5;
	}

	.body {
		margin-top: var(--space-3);
		font-size: var(--text-sm);
	}

	.actions {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-2);
		margin-top: var(--space-5);
	}

	/* On a phone the two answers take the width between them, so neither is
	   a small target at the edge of the screen. */
	@media (max-width: 34rem) {
		.actions > :global(*) {
			flex: 1;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		:global(.confirm-content[data-scope='dialog'][data-part='content'][data-state='open']),
		:global(.confirm-content[data-scope='dialog'][data-part='content'][data-state='closed']) {
			animation-duration: 1ms;
		}
	}
</style>
