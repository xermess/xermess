<script lang="ts">
	import FlowEditor from '$lib/components/flows/editor/FlowEditor.svelte';
	import { TEMPLATES, freeSlug, slugFrom } from '$lib/components/flows/steps';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	/** The template's shape, named in the panel's language, under an
	    identifier nobody has yet. */
	const initial = $derived.by(() => {
		const template = TEMPLATES.find((one) => one.id === data.template) ?? TEMPLATES[0];
		const name = template.name;

		return {
			...template.draft,
			name,
			description: template.hint,
			slug: freeSlug(slugFrom(name), data.taken)
		};
	});
</script>

<svelte:head><title>New flow · xermess admin</title></svelte:head>

{#key data.template}
	<FlowEditor flow={null} {initial} kinds={data.stepKinds} taken={data.taken} editable />
{/key}
