<script lang="ts">
	import type { Snippet } from 'svelte';
	import { QueryClientProvider } from '@tanstack/svelte-query';
	import { createQueryClient } from '$lib/query';
	import '$lib/styles/app.css';

	let { children }: { children: Snippet } = $props();

	// One cache for this visitor, built here so that rendering on the server
	// gives each request its own.
	const queryClient = createQueryClient();
</script>

<!-- No <title> here on purpose. A title in this layout is only applied once
     the app has booted, so it lands between the one the browser already has
     and the one the page sets, which shows as a second flicker. The default
     lives in app.html instead, where it is in the very first byte. -->
<QueryClientProvider client={queryClient}>
	{@render children()}
</QueryClientProvider>
