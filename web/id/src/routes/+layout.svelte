<script lang="ts">
	import { Toaster } from '$lib/components';
	import { provideTranslator, translator } from '$lib/i18n';
	import '$lib/styles/app.css';
	import type { LayoutProps } from './$types';

	let { data, children }: LayoutProps = $props();

	// One translator for every component below; it reads the request's messages on each lookup,
	// so a language change re-renders immediately.
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

<!-- The account pages' toasts. -->
<Toaster />
