<script lang="ts">
	import Button from './Button.svelte';

	type Props = {
		/** Whether the save is under way. */
		saving?: boolean;
		/** Whether the form can be saved as it stands — false while something
		    it needs is missing. */
		ready?: boolean;
		/** Puts the form back to what is stored. */
		ondiscard: () => void;
	};

	let { saving = false, ready = true, ondiscard }: Props = $props();
</script>

<!-- The buttons of a settings page that saves as a whole: shown only once
     something has been changed, and then stuck to the foot of the window as a
     floating card, so they are within reach of whichever field is being
     edited. It submits the form it is in. -->
<div class="save-bar" role="region" aria-label="Unsaved changes">
	<span class="label"><span class="dot" aria-hidden="true"></span>Unsaved changes</span>

	<Button variant="subtle" size="sm" onclick={ondiscard} disabled={saving}>Discard</Button>
	<Button type="submit" size="sm" loading={saving} disabled={saving || !ready}>
		{saving ? 'Saving…' : 'Save changes'}
	</Button>
</div>

<style>
	.save-bar {
		position: sticky;
		bottom: var(--space-3);
		z-index: 5;
		display: flex;
		align-items: center;
		gap: var(--space-2);
		padding: var(--space-2) var(--space-2) var(--space-2) var(--space-4);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-surface);
		background: var(--color-surface);
		box-shadow: var(--shadow-md);
		animation: save-bar-in var(--speed) ease-out;
	}

	.label {
		display: inline-flex;
		flex: 1;
		align-items: center;
		gap: var(--space-2);
		min-width: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.dot {
		flex: none;
		width: 8px;
		height: 8px;
		border-radius: var(--radius-pill);
		background: var(--color-warning);
	}

	@keyframes save-bar-in {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.save-bar {
			animation: none;
		}
	}
</style>
