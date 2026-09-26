<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Menu } from '@ark-ui/svelte/menu';
	import { Portal } from '@ark-ui/svelte/portal';

	type Props = {
		/** What the button opens, for anyone not looking at it — and its tooltip. */
		label: string;
		/** What the button shows: an icon, an avatar. */
		trigger: Snippet;
		/** The rows: MenuGroup, MenuItem, MenuInfo and MenuSeparator. */
		children: Snippet;
		/** `icon` is a square button the height of a small control; `avatar`
		    is a round one for a picture of somebody. */
		shape?: 'icon' | 'avatar';
		/** How wide the menu opens, so it keeps one shape whatever is in it. */
		width?: string;
	};

	let { label, trigger, children, shape = 'icon', width = '17rem' }: Props = $props();
</script>

<!-- A button in the header that opens a card of rows under it, against the
     right edge. The card is portalled, so nothing it opens over can clip it;
     how a row looks is MenuItem's, and what a card looks like is ark.css's. -->
<Menu.Root positioning={{ placement: 'bottom-end', gutter: 6 }}>
	<Menu.Trigger
		class="control dropdown-trigger"
		data-size="sm"
		data-variant="ghost"
		data-palette="neutral"
		data-shape={shape}
		aria-label={label}
		title={label}
	>
		{@render trigger()}
	</Menu.Trigger>

	<Portal>
		<Menu.Positioner>
			<Menu.Content class="dropdown-menu" style="--menu-width: {width}">
				{@render children()}
			</Menu.Content>
		</Menu.Positioner>
	</Portal>
</Menu.Root>

<style>
	:global(.dropdown-trigger.control) {
		width: var(--control-height-sm);
		padding: 0;
		color: var(--color-text-hint);
	}

	:global(.dropdown-trigger.control:hover),
	:global(.dropdown-trigger.control[data-state='open']) {
		color: var(--color-text);
	}

	:global(.dropdown-trigger.control[data-state='open']) {
		background: var(--palette-subtle);
	}

	/* An avatar fills its button, and says it is open with a ring rather
	   than a fill behind it, which a round picture would hide. */
	:global(.dropdown-trigger.control[data-shape='avatar']) {
		border-radius: var(--radius-pill);
	}

	:global(.dropdown-trigger.control[data-shape='avatar']:hover),
	:global(.dropdown-trigger.control[data-shape='avatar'][data-state='open']) {
		background: transparent;
		box-shadow: 0 0 0 2px var(--color-border);
	}

	:global(.dropdown-menu) {
		width: min(var(--menu-width), calc(100vw - var(--space-4)));
	}
</style>
