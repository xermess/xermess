<script lang="ts">
	import { Menu } from '@ark-ui/svelte/menu';
	import { Portal } from '@ark-ui/svelte/portal';
	import { RiCheckLine, RiFontFamily } from 'svelte-remixicon';
	import { FONTS, font, isFont } from '$lib/state/font.svelte';
	import type { Size } from './control';
	import Icon from './Icon.svelte';

	type Props = { size?: Size; variant?: 'ghost' | 'subtle' };

	let { size = 'md', variant = 'subtle' }: Props = $props();

	/** The glyph is a share of the button, the same way IconButton scales. */
	const glyph: Record<Size, string> = { sm: '1rem', md: '1.125rem', lg: '1.25rem' };
</script>

<!--
	Which typeface the panel is set in, while one is being chosen. Each row is
	written in the face it names, so the menu is its own preview; picking one
	changes the whole panel at once, and the choice is kept in a cookie the
	server reads, the way the theme is.

	No tooltip on the trigger: a tooltip and a menu on one button fight over
	the same handlers, and the menu names itself as soon as it opens.
-->
<Menu.Root positioning={{ placement: 'bottom-end', gutter: 6 }}>
	<Menu.Trigger
		class="control"
		data-icon="true"
		data-size={size}
		data-variant={variant}
		data-palette="neutral"
		aria-label="Choose the panel's typeface"
	>
		<Icon icon={RiFontFamily} size={glyph[size]} />
	</Menu.Trigger>

	<Portal>
		<Menu.Positioner>
			<Menu.Content class="font-menu">
				<Menu.RadioItemGroup
					value={font.current}
					onValueChange={(details) => {
						if (isFont(details.value)) font.set(details.value);
					}}
				>
					<Menu.ItemGroupLabel>Typeface</Menu.ItemGroupLabel>

					{#each FONTS as face (face.id)}
						<Menu.RadioItem value={face.id}>
							<span class="face" data-face={face.id}>
								<Menu.ItemText class="name">{face.label}</Menu.ItemText>
								<span class="detail">{face.detail}</span>
							</span>
							<Menu.ItemIndicator class="check">
								<Icon icon={RiCheckLine} />
							</Menu.ItemIndicator>
						</Menu.RadioItem>
					{/each}
				</Menu.RadioItemGroup>
			</Menu.Content>
		</Menu.Positioner>
	</Portal>
</Menu.Root>

<style>
	:global(.font-menu) {
		min-width: 15rem;
	}

	.face {
		display: flex;
		flex: 1;
		flex-direction: column;
		min-width: 0;
		line-height: 1.3;
	}

	/* Each row is set in the face it offers, whatever the panel is set in. */
	.face[data-face='roboto'] {
		font-family: 'Roboto', sans-serif;
	}

	.face[data-face='product-sans'] {
		font-family: 'Product Sans', sans-serif;
	}

	.face[data-face='plex'] {
		font-family: 'IBM Plex Sans', sans-serif;
	}

	.face :global(.name) {
		color: var(--color-text);
	}

	.detail {
		color: var(--color-text-hint);
		font-size: var(--text-xs);
	}

	:global(.font-menu .check) {
		display: inline-flex;
		flex: none;
		width: 16px;
		color: var(--color-text);
	}

	/* The column stays reserved on every row, so the names line up whichever
	   one is ticked. */
	:global(.font-menu .check[data-state='unchecked']) {
		visibility: hidden;
	}
</style>
