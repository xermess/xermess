<script lang="ts">
	import type { ComponentType, Snippet } from 'svelte';
	import Thumb from './Thumb.svelte';

	type Props = {
		title: string;
		/** A line under the title: what the section is for. */
		description?: string;
		/** A mark beside the title, for a page of settings whose sections are
		    subjects of their own — the organisation's identity, its support. */
		icon?: ComponentType;
		/** Tags under the description: where the values show, or their state. */
		meta?: Snippet;
		/** Controls on the right of the title, such as an "Add" button. */
		action?: Snippet;
		children: Snippet;
	};

	let { title, description, icon, meta, action, children }: Props = $props();
</script>

<!--
	One titled group of a form, laid out by its own width: title beside a card of fields where
	there is room, title above fields where there is not.
-->
<section class="form-section">
	<div class="layout">
		<header>
			<div class="heading">
				{#if icon}
					<span class="icon" aria-hidden="true"><Thumb {icon} size="sm" /></span>
				{/if}

				<div class="text">
					<h3>{title}</h3>
					{#if description}
						<p>{description}</p>
					{/if}
					{#if meta}
						<div class="meta">{@render meta()}</div>
					{/if}
				</div>
			</div>

			{#if action}
				<div class="action">{@render action()}</div>
			{/if}
		</header>

		<div class="content">
			{@render children()}
		</div>
	</div>
</section>

<style>
	.form-section {
		container-type: inline-size;
	}

	.form-section + :global(.form-section) {
		margin-top: var(--space-5);
	}

	/* The rule is on the layout rather than the section, since a section is
	   its own container and can only restyle what is inside it: ruled off
	   while stacked, and not once the fields are in a card of their own. */
	:global(.form-section) + .form-section > .layout {
		padding-top: var(--space-5);
		border-top: 1px solid var(--color-border);
	}

	.layout {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}

	header {
		/* Thumb's `sm` size, for what lines up beside it. */
		--icon-width: 34px;

		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-3);
	}

	.heading {
		display: flex;
		align-items: flex-start;
		gap: var(--space-3);
		min-width: 0;
	}

	/* The title sits on the icon's middle line. */
	.icon {
		display: flex;
		flex: none;
	}

	.icon + .text {
		padding-top: 7px;
	}

	.text {
		min-width: 0;
	}

	.meta {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-1);
		margin-top: var(--space-2);
	}

	h3 {
		margin: 0;
		color: var(--color-text);
		font-size: var(--text-base);
		font-weight: 600;
	}

	p {
		margin: 2px 0 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		line-height: 1.45;
	}

	.action {
		display: flex;
		flex-shrink: 0;
		flex-wrap: wrap;
		gap: var(--space-2);
	}

	.content {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		min-width: 0;
	}

	/* A switch brings a row's padding with it, for the rule between two of
	   them. At the edges of the section that padding is only a gap, so it
	   goes — directly inside, or inside the one group that holds them. */
	.content > :global(.switch-field:first-child),
	.content > :global(:first-child:not(.switch-field) > .switch-field:first-child) {
		padding-top: 0;
	}

	.content > :global(.switch-field:last-child),
	.content > :global(:last-child:not(.switch-field) > .switch-field:last-child) {
		padding-bottom: 0;
	}

	@container (min-width: 44rem) {
		.layout {
			display: grid;
			grid-template-columns: 14rem minmax(0, 1fr);
			/* The card is as tall as its fields, not as its title's column. */
			align-items: start;
			gap: var(--space-5);
		}

		:global(.form-section) + .form-section > .layout {
			padding-top: 0;
			border-top: none;
		}

		/* Beside the card, the title reads down and its action sits under
		   the description rather than across from it. */
		header {
			flex-direction: column;
			justify-content: flex-start;
			padding-top: var(--space-2);
		}

		/* Under an icon's title, the action lines up with the words, not
		   with the icon. */
		.heading:has(.icon) + .action {
			padding-left: calc(var(--icon-width) + var(--space-3));
		}

		.content {
			padding: var(--space-4);
			border: 1px solid var(--color-border);
			border-radius: var(--radius-surface);
			background: var(--color-surface);
		}
	}

	/* On a settings page there is room to spare, and a title with an icon
	   beside it needs more of it than a title alone. */
	@container (min-width: 60rem) {
		.layout {
			grid-template-columns: 17rem minmax(0, 1fr);
		}
	}
</style>
