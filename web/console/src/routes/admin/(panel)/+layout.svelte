<script lang="ts">
	import { untrack, type Snippet } from 'svelte';
	import AppHeader from '$lib/components/layout/AppHeader.svelte';
	import { provideShell } from '$lib/state/shell.svelte';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: Snippet } = $props();

	/** Starts as the server rendered it and is the reader's from then on, so
	    the value is read once on purpose: a later load returns the same
	    cookie, and re-reading it would undo a toggle. */
	const shell = provideShell(untrack(() => data.sidebar));
</script>

<div class="shell" class:mini={shell.collapsed}>
	<AppHeader admin={data.admin} />

	<main>
		{@render children()}
	</main>
</div>

<style>
	/* The column's width is one registered property, animated here, so the
	   header's logo block and the sidebar under it are the same number on
	   every frame of the fold rather than two animations kept in step. */
	@property --sidebar-width {
		syntax: '<length>';
		inherits: true;
		initial-value: 240px;
	}

	.shell {
		--sidebar-width: 240px;

		height: 100dvh;
		overflow: hidden;
		transition: --sidebar-width var(--speed-drawer) cubic-bezier(0.4, 0, 0.2, 1);
	}

	/* Folded, the column is just wide enough for the icons. */
	.shell.mini {
		--sidebar-width: 56px;
	}

	@media (prefers-reduced-motion: reduce) {
		.shell {
			transition: none;
		}
	}

	/* Pages fill the width, the way PocketBase does: a table is easier to
	   read with room for its columns than centred in a narrow column, and it
	   runs to the edges rather than sitting in a box. */
	main {
		height: 100%;
		overflow-y: auto;
		overscroll-behavior-y: auto;
		padding: calc(var(--header-height) + var(--space-4)) 0 var(--space-4);
		scrollbar-gutter: stable;
	}

	/* The dashboard brings its own sidebar and pads its own content.
	   .dashboard belongs to a child component, which is why :has has to match
	   it globally. */
	main:has(:global(.dashboard)) {
		padding: var(--header-height) 0 0;
	}
</style>
