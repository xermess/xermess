<script lang="ts">
	import type { ComponentType } from 'svelte';

	type Props = {
		/** A component from svelte-remixicon, as in the console. */
		icon: ComponentType;
		size?: string;
	};

	let { icon: Glyph, size = '1rem' }: Props = $props();
</script>

<Glyph {size} class="icon" aria-hidden="true" />

<style>
	:global(.icon) {
		flex: none;
		vertical-align: -0.125em;
	}
</style>
