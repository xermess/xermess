<script lang="ts">
	import { tick, untrack } from 'svelte';
	import { navigating } from '$app/state';
	import { NavigationProgress } from '$lib/state/progress.svelte';

	/** How long the finished bar takes to fade out. */
	const FADE = 300;

	let bar: HTMLDivElement;

	// The jump back to the start is put on the page and read back — the read
	// makes the browser apply it — before the transition returns, so the bar
	// starts from nothing rather than sliding back. No animation frame is
	// waited for: a tab in the background never gets one.
	const progress = new NavigationProgress(async () => {
		await tick();
		void bar.offsetWidth;
	});

	// Only whether a navigation is under way is watched; what start and
	// finish read of the bar's own state is not, or showing it would run
	// this again.
	$effect(() => {
		const going = navigating.to !== null;

		untrack(() => (going ? progress.start() : progress.finish()));
	});

	$effect(() => () => progress.dispose());
</script>

<!-- The bar across the top of the window while a page is on its way, the
     way NProgress draws one. It is moved, never resized — a transform, so the
     browser slides it without laying anything out — and it is only a
     picture: the page itself says when it has changed. The timing is in
     lib/state/progress.svelte.ts. -->
<div
	bind:this={bar}
	class="progress"
	class:visible={progress.visible}
	class:instant={progress.instant}
	style:--progress={progress.value}
	style:--fade="{FADE}ms"
	aria-hidden="true"
></div>

<style>
	.progress {
		position: fixed;
		top: 0;
		left: 0;
		z-index: 100;
		width: 100%;
		height: 3px;
		background: var(--color-brand);
		opacity: 0;
		pointer-events: none;
		transform: translateX(calc((var(--progress) - 1) * 100%));
		transition:
			transform 250ms ease-out,
			opacity var(--fade) ease;
	}

	.visible {
		opacity: 1;
		/* Only while it is out does it get a layer of its own. */
		will-change: transform;
		transition: transform 250ms ease-out;
	}

	.instant {
		transition: none;
	}

	/* NProgress's glow at the leading edge: a soft light where the bar is
	   heading, which is what makes it read as moving rather than as a line. */
	.progress::after {
		content: '';
		position: absolute;
		top: 0;
		right: 0;
		width: 96px;
		height: 100%;
		box-shadow:
			0 0 10px var(--color-brand),
			0 0 5px var(--color-brand);
		opacity: 0.9;
		transform: rotate(2deg) translateY(-3px);
	}

	@media (prefers-reduced-motion: reduce) {
		.progress,
		.visible {
			transition: opacity var(--fade) ease;
		}
	}
</style>
