<!--
  @component
  Draws the panel's toasts (see ./toast.ts), centred at the bottom of the
  screen, newest nearest the edge. Mounted once, in the root layout.

  A toast is a pill in the theme turned over — black on the light page, white
  on the dark — so it is noticed without shouting: an icon in the brand's colour, the
  message, and a round close button. The icon is one colour whatever the
  toast says; its shape tells a success from a failure. A second line of detail runs on after
  the title rather than beneath it, which keeps the pill one line tall
  whenever it fits. It pauses while the pointer is over it, and Ark announces
  each to assistive technology as it arrives.
-->
<script lang="ts">
	import { Toast, Toaster } from '@ark-ui/svelte/toast';
	import { RiCheckboxCircleFill, RiCloseLine, RiErrorWarningFill } from 'svelte-remixicon';
	import Icon from './Icon.svelte';
	import { toaster } from './toast';
</script>

<Toaster {toaster}>
	{#snippet children(toast)}
		<Toast.Root class="toast">
			<span class="glyph"
				><Icon
					icon={toast().type === 'success' ? RiCheckboxCircleFill : RiErrorWarningFill}
					size="1.125rem"
				/></span
			>
			<div class="message">
				<Toast.Title class="title">{toast().title}</Toast.Title>
				{#if toast().description}
					<Toast.Description class="description">{toast().description}</Toast.Description>
				{/if}
			</div>
			<Toast.CloseTrigger class="close" aria-label="Dismiss">
				<Icon icon={RiCloseLine} size="0.875rem" />
			</Toast.CloseTrigger>
		</Toast.Root>
	{/snippet}
</Toaster>

<style>
	/* Positioned and animated by Ark, through these variables; the rest is
	   the pill. */
	:global(.toast) {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		width: max-content;
		max-width: min(30rem, calc(100vw - 2 * var(--page-gutter)));
		min-height: 44px;
		padding: 6px 6px 6px 12px;
		border-radius: var(--radius-pill);
		background: var(--toast-surface);
		color: var(--toast-text);
		box-shadow:
			0 12px 32px -8px rgb(0 0 0 / 45%),
			0 2px 6px rgb(0 0 0 / 20%);

		translate: var(--x) var(--y);
		scale: var(--scale);
		z-index: var(--z-index);
		height: var(--height);
		opacity: var(--opacity);
		will-change: translate, opacity, scale;
		transition:
			translate 400ms,
			scale 400ms,
			opacity 400ms,
			height 400ms;
		transition-timing-function: cubic-bezier(0.21, 1.02, 0.73, 1);
	}

	:global(.toast[data-state='closed']) {
		transition:
			translate 200ms,
			scale 200ms,
			opacity 200ms;
		transition-timing-function: cubic-bezier(0.06, 0.71, 0.55, 1);
	}

	.glyph {
		display: inline-flex;
		flex: none;
		color: var(--toast-icon);
	}

	/* The title, and the detail running on after it. */
	.message {
		flex: 1;
		min-width: 0;
		font-size: var(--text-base);
		line-height: 1.4;
		overflow-wrap: anywhere;
	}

	:global(.toast .title),
	:global(.toast .description) {
		display: inline;
	}

	:global(.toast .title) {
		font-weight: 600;
	}

	:global(.toast .description) {
		margin-left: 6px;
		color: var(--toast-hint);
		font-size: var(--text-sm);
	}

	:global(.toast .close) {
		display: inline-flex;
		flex: none;
		align-items: center;
		justify-content: center;
		width: 28px;
		height: 28px;
		border: none;
		border-radius: var(--radius-pill);
		background: color-mix(in srgb, var(--toast-text), transparent 90%);
		color: var(--toast-text);
		cursor: pointer;
		transition: background-color var(--speed-fast);
	}

	:global(.toast .close:hover) {
		background: color-mix(in srgb, var(--toast-text), transparent 80%);
	}

	:global(.toast .close:focus-visible) {
		outline: 2px solid var(--toast-icon);
		outline-offset: 1px;
	}

	@media (prefers-reduced-motion: reduce) {
		:global(.toast),
		:global(.toast[data-state='closed']) {
			transition: opacity 150ms;
		}
	}
</style>
