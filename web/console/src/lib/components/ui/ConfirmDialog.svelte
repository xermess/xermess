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

<!--
	A confirmation before something hard to undo. It is an alert dialog (announced on open),
	stacked above full-window dialogs, and styled like the popovers.
-->
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
						<Icon icon={icon ?? icons[tone]} size="18" />
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
	/* The card itself — centred, bordered, how it arrives — is every
	dialog's, in styles/ark.css. A question is narrower, padded, and sits
	above the dialog it is usually asked from. */
	:global(.confirm-backdrop[data-scope='dialog'][data-part='backdrop']) {
		z-index: 60;
	}

	:global(.confirm-positioner[data-scope='dialog'][data-part='positioner']) {
		z-index: 65;
	}

	:global(.confirm-content[data-scope='dialog'][data-part='content']) {
		--dialog-width: 26rem;

		padding: var(--space-4);
	}

	.head {
		display: flex;
		align-items: flex-start;
		gap: var(--space-3);
	}

	/* The tone, said once: a circle in the palette the yes button is in,
	   the way a menu's danger row is tinted. */
	.badge {
		display: inline-flex;
		flex: none;
		align-items: center;
		justify-content: center;
		width: 34px;
		height: 34px;
		border: 1px solid var(--palette-border);
		border-radius: var(--radius-control);
		background: var(--palette-subtle);
		color: var(--palette-fg);
	}

	.text {
		flex: 1;
		min-width: 0;
		padding-top: 6px;
	}

	.text :global([data-scope='dialog'][data-part='title']) {
		font-size: var(--text-base);
		line-height: 1.4;
		overflow-wrap: anywhere;
	}

	.text :global([data-scope='dialog'][data-part='description']) {
		margin-top: 2px;
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
		margin-top: var(--space-4);
	}

	/* On a phone the card uses more of the narrow window, and the two
	   answers take the width between them, so neither is a small target at
	   the edge of the screen. */
	@media (max-width: 34rem) {
		:global(.confirm-positioner[data-scope='dialog'][data-part='positioner']) {
			padding: var(--space-2);
		}

		.actions > :global(*) {
			flex: 1;
		}
	}
</style>
