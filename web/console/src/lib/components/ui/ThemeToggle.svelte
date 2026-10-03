<script lang="ts">
	import { RiMoonLine, RiSunLine } from 'svelte-remixicon';
	import { theme } from '$lib/state/theme.svelte';
	import type { Size } from './control';
	import Icon from './Icon.svelte';
	import Tooltip from './Tooltip.svelte';

	type Props = { size?: Size; variant?: 'ghost' | 'subtle' };

	let { size = 'md', variant = 'subtle' }: Props = $props();

	/** The glyph is a share of the button, the same way IconButton scales. */
	const glyph: Record<Size, string> = { sm: '1rem', md: '1.125rem', lg: '1.25rem' };
</script>

<!--
	Both icons render and CSS shows the one matching the server-set data-theme, so the first
	frame is right with no flicker.
-->
<Tooltip label="Switch theme">
	{#snippet children(trigger)}
		<button
			{...trigger()}
			type="button"
			class="control"
			data-icon="true"
			data-size={size}
			data-variant={variant}
			data-palette="neutral"
			onclick={() => theme.toggle()}
			aria-label="Switch between the light and dark theme"
		>
			<span class="icons" style="--glyph: {glyph[size]}">
				<span class="moon"><Icon icon={RiMoonLine} size={glyph[size]} /></span>
				<span class="sun"><Icon icon={RiSunLine} size={glyph[size]} /></span>
			</span>
		</button>
	{/snippet}
</Tooltip>

<style>
	/* The shape and the press come from the shared control; what belongs to
	   this button is the two icons turning into each other. */

	.icons {
		position: relative;
		display: grid;
		place-items: center;
		width: var(--glyph);
		height: var(--glyph);
	}

	.moon,
	.sun {
		position: absolute;
		display: grid;
		place-items: center;
		transition:
			opacity 200ms ease,
			transform 300ms cubic-bezier(0.4, 0, 0.2, 1);
	}

	/* The icon shows the theme you would switch to: a moon while light, a sun
	   while dark. The hidden one waits rotated a quarter turn away, so the
	   pair turn into each other when the theme changes. */
	.sun {
		opacity: 0;
		transform: rotate(-90deg) scale(0.5);
	}

	:global([data-theme='dark']) .sun {
		opacity: 1;
		transform: rotate(0) scale(1);
	}

	:global([data-theme='dark']) .moon {
		opacity: 0;
		transform: rotate(90deg) scale(0.5);
	}

	/* Nobody has chosen yet: follow the system, the same way the palette does. */
	@media (prefers-color-scheme: dark) {
		:global(:root:not([data-theme='light'])) .sun {
			opacity: 1;
			transform: rotate(0) scale(1);
		}

		:global(:root:not([data-theme='light'])) .moon {
			opacity: 0;
			transform: rotate(90deg) scale(0.5);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		button,
		.moon,
		.sun {
			transition: none;
		}
	}
</style>
