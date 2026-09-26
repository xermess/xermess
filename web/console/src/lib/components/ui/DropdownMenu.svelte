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
		/** `avatar` for a picture of somebody, which fills more of the
		    button than an icon does. The button is the same either way. */
		/** `labelled` is a pill with words in it — an avatar and a name — for
		    the one menu that says whose it is. It needs no tooltip, having
		    its words on it. */
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

<!-- A button in the header that opens a card of rows under it, against the
     right edge. The card is portalled, so nothing it opens over can clip it;
     how a row looks is MenuItem's, and what a card looks like is ark.css's. -->
<!-- A button in the header that opens a card of rows under it, against the
     right edge. The button is an icon button — the same round control, the
     same size and hover, the same tooltip — so it sits in a row of them as
     one of the set. The card is portalled, so nothing it opens over can clip
     it; how a row looks is MenuItem's, and what a card looks like is
     ark.css's. -->
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

	:global(.dropdown-menu) {
		width: min(var(--menu-width), calc(100vw - var(--space-4)));
	}
</style>
