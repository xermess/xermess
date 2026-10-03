<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Menu } from '@ark-ui/svelte/menu';
	import { Portal } from '@ark-ui/svelte/portal';
	import { Tooltip } from '@ark-ui/svelte/tooltip';
	import { mergeProps } from '@ark-ui/svelte/utils';

	type Props = {
		/** What the button opens, for anyone not looking at it — and its tooltip. */
		label: string;
		/** What the button shows: an icon, an avatar. */
		trigger: Snippet;
		/** The rows: MenuGroup, MenuItem, MenuInfo and MenuSeparator. */
		children: Snippet;
		/**
		 * `avatar` for a picture of somebody; `labelled` for a button with an avatar and a name,
		 * which needs no tooltip.
		 */
		shape?: 'icon' | 'avatar' | 'labelled';
		/** How wide the menu opens, so it keeps one shape whatever is in it. */
		width?: string;
	};

	let { label, trigger, children, shape = 'icon', width = '17rem' }: Props = $props();

	/** The menu and the tooltip both find their trigger by its id, to place
	    themselves against it; one button is both, so they share one. */
	const uid = $props.id();
	const id = `${uid}-trigger`;
</script>

<!--
	A header icon button that opens a portalled card of rows at the right edge. Row look is
	MenuItem's, card look is ark.css's.
-->
<Menu.Root positioning={{ placement: 'bottom-end', gutter: 6 }} ids={{ trigger: id }}>
	<Tooltip.Root
		disabled={shape === 'labelled'}
		openDelay={250}
		closeDelay={80}
		positioning={{ placement: 'bottom', gutter: 6, strategy: 'fixed' }}
		ids={{ trigger: id }}
	>
		<Tooltip.Trigger>
			{#snippet asChild(tooltip)}
				<Menu.Trigger>
					{#snippet asChild(menu)}
						<!-- The menu's props last, so its open state is the one the
						     button says it is in. -->
						<button
							{...mergeProps(tooltip(), menu())}
							type="button"
							class="control dropdown-trigger"
							data-icon={shape === 'labelled' ? undefined : 'true'}
							data-size="sm"
							data-variant="ghost"
							data-palette="neutral"
							data-shape={shape}
							aria-label={label}
						>
							{@render trigger()}
						</button>
					{/snippet}
				</Menu.Trigger>
			{/snippet}
		</Tooltip.Trigger>

		<Portal>
			<Tooltip.Positioner>
				<Tooltip.Content>{label}</Tooltip.Content>
			</Tooltip.Positioner>
		</Portal>
	</Tooltip.Root>

	<Portal>
		<Menu.Positioner>
			<Menu.Content class="dropdown-menu" style="--menu-width: {width}">
				{@render children()}
			</Menu.Content>
		</Menu.Positioner>
	</Portal>
</Menu.Root>

<style>
	/* Open, it keeps the fill an icon button has under the pointer, so it is
	   plain which button the card belongs to. */
	:global(.dropdown-trigger.control[data-state='open']) {
		background: var(--palette-subtle);
	}

	/* Round like the icon buttons beside it, and only as wide as its words:
	   the picture flush to the left, the name, a chevron. */
	:global(.dropdown-trigger.control[data-shape='labelled']) {
		gap: var(--space-2);
		max-width: 16rem;
		padding: 0 var(--space-2) 0 3px;
		border-radius: var(--radius-pill);
		color: var(--color-text);
		font-weight: 500;
	}

	/* Where the bar is too narrow for a name, the menu hides its words (see
	   AccountMenu), and what is left is a circle the size of the icon buttons
	   beside it — not a stretched pill around one picture. */
	@media (max-width: 64rem) {
		:global(.dropdown-trigger.control[data-shape='labelled']) {
			width: var(--control-height-sm);
			padding: 0;
			gap: 0;
		}
	}

	:global(.dropdown-menu) {
		width: min(var(--menu-width), calc(100vw - var(--space-4)));
	}
</style>
