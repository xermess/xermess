<script lang="ts">
	import { QueryClientProvider } from '@tanstack/svelte-query';
	import { provideTranslator, translator } from '$lib/i18n';
	import { createQueryClient } from '$lib/query';
	import '$lib/styles/app.css';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();

	// Every component below looks its text up through this one translator, and
	// it reads `data.messages` each time — so the panel re-renders in the new
	// language the moment the choice lands.
	provideTranslator(translator(() => data.messages));

	// hooks.server.ts names the document before anything is loaded; this keeps
	// it right when the language is changed without a reload.
	$effect(() => {
		document.documentElement.lang = data.language;
	});

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
