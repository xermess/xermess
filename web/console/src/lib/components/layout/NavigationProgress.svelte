<script lang="ts">
	import { tick, untrack } from 'svelte';
	import { afterNavigate, beforeNavigate } from '$app/navigation';
	import { navigating } from '$app/state';
	import { NavigationProgress } from '$lib/state/progress.svelte';

	/** How long the finished bar takes to fade out. */
	const FADE = 150;

	let bar: HTMLDivElement;

	// The jump back to the start is put on the page and read back — the read
	// makes the browser apply it — before the transition returns, so the bar
	// starts from nothing rather than sliding back. No animation frame is
	// waited for: a tab in the background never gets one.
	const progress = new NavigationProgress(async () => {
		await tick();
		void bar.offsetWidth;
	});

	// Each navigation is caught as it begins. Finishing is shared by all
	// successful navigations because SvelteKit does not run another
	// beforeNavigate callback while one is already in progress; an aborted
	// navigation must therefore leave the bar alone when another one has
	// already replaced it. The navigating watcher below catches a failed
	// replacement, while afterNavigate also covers a very quick preload.
	// Leaving the app is the browser's to show, not ours.
	const finish = () => void progress.finish();

	beforeNavigate((navigation) => {
		if (navigation.willUnload) return;

		progress.start();
		void navigation.complete.catch(() => {
			queueMicrotask(() => {
				if (navigating.to === null) finish();
			});
		});
	});

	afterNavigate(finish);

	$effect(() => {
		if (navigating.to !== null) return;
		untrack(finish);
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
		height: 2px;
		background: var(--color-brand);
		opacity: 0;
		pointer-events: none;
		transform: translateX(calc((var(--progress) - 1) * 100%));
		transition:
			transform 120ms linear,
			opacity var(--fade) linear;
	}

	.visible {
		opacity: 1;
		/* Only while it is out does it get a layer of its own. */
		will-change: transform;
		transition: transform 120ms linear;
	}

	.instant {
		transition: none;
	}

	@media (prefers-reduced-motion: reduce) {
		.progress,
		.visible {
			transition: opacity var(--fade) linear;
		}
	}
</style>
