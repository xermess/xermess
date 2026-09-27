<script lang="ts">
	import type { ComponentType } from 'svelte';
	import Icon from './Icon.svelte';

	type Props = {
		/** A picture: a logo, a photo. While it loads, and if it never does,
		    the icon or the letters are shown instead. */
		src?: string;
		/** An icon to show. */
		icon?: ComponentType;
		/** Or a few letters, such as someone's initials. */
		text?: string;
		/** 28px for a bar, 34px for a row, the height of a control for a card. */
		size?: 'xs' | 'sm' | 'md';
		/** `circle` is for a person; `square` for everything else. */
		shape?: 'square' | 'circle';
		/** `accent` fills it with the brand, for the one mark that stands for
		    the account or the product; `neutral` is the quiet tint. */
		tone?: 'neutral' | 'accent';
		/** Show a picture as it is, with no frame or tint behind it: for a
		    logo that has its own shape and should be seen whole. */
		bare?: boolean;
	};

	let {
		src,
		icon,
		text,
		size = 'sm',
		shape = 'square',
		tone = 'neutral',
		bare = false
	}: Props = $props();

	/** The address that did not load, rather than a flag: a new one gets its
	    own chance without anything having to reset this. */
	let failed = $state('');

	const picture = $derived(src && src !== failed ? src : '');

	const glyph = { xs: '0.875rem', sm: '1rem', md: '1.15rem' };
</script>

<!-- PocketBase's .thumb: a square on the faint tint, in the hint colour, for
     what leads a row, a card or a bar — a logo, an avatar, an icon. -->
<span
	class="thumb {size} {shape} {tone}"
	class:picture
	class:bare={bare && picture}
	aria-hidden="true"
>
	{#if picture}
		<img src={picture} alt="" onerror={() => (failed = picture)} />
	{:else if icon}
		<Icon {icon} size={glyph[size]} />
	{:else if text}
		{text}
	{/if}
</span>

<style>
	.thumb {
		display: inline-flex;
		flex-shrink: 0;
		align-items: center;
		justify-content: center;
		overflow: hidden;
		aspect-ratio: 1;
		border: 1px solid var(--color-secondary-alt);
		border-radius: var(--radius-sm);
		background: var(--color-surface-alt);
		color: var(--color-text-hint);
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.02em;
	}

	.xs {
		width: 28px;
		font-size: 10px;
	}

	.sm {
		width: 34px;
	}

	.md {
		width: var(--control-height);
		font-size: var(--text-sm);
	}

	.circle {
		border-radius: var(--radius-pill);
	}

	.accent {
		border-color: transparent;
		background: var(--color-accent);
		color: var(--color-accent-text);
		font-weight: 700;
	}

	/* A picture brings its own colours; the frame only holds it. */
	.picture {
		border-color: var(--color-border);
		background: var(--color-surface-alt);
	}

	/* Bare, the picture is the whole mark: no border, no tint, no padding,
	   and its own corners — unless it was asked to be a circle, which is a
	   shape to cut it to rather than a frame around it. */
	.picture.bare {
		border: none;
		border-radius: 0;
		background: transparent;
	}

	.picture.bare.circle {
		border-radius: var(--radius-pill);
	}

	img {
		width: 100%;
		height: 100%;
		object-fit: contain;
	}
</style>
