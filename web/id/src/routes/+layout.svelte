<script lang="ts">
	import { provideTranslator, translator } from '$lib/i18n';
	import '$lib/styles/app.css';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();

	// Every component below looks its text up through this one translator. The
	// browser registers `data.messages` with svelte-i18n, and the lookup reads
	// the request's messages each time — so a page re-renders in the new
	// language the moment the choice lands, without any of them subscribing to
	// anything themselves.
	provideTranslator(
		translator(
			() => data.messages,
			() => data.language
		)
	);

	// hooks.server.ts names the document on the server; this keeps it right
	// when the language is switched without a reload.
	$effect(() => {
		document.documentElement.lang = data.language;
	});
</script>

{@render children()}
