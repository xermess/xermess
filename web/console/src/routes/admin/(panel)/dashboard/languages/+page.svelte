<script lang="ts">
	import { RiAddLine, RiRefreshLine } from 'svelte-remixicon';
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import type { Language, LocaleApp } from '$lib/api';
	import { BRAND } from '$lib/brand';
	import { Button, Icon, IconButton, PageHeader, SearchInput } from '$lib/components/ui';
	import LanguageDrawer from '$lib/components/languages/LanguageDrawer.svelte';
	import LanguageTable from '$lib/components/languages/LanguageTable.svelte';
	import NewLanguageDrawer from '$lib/components/languages/NewLanguageDrawer.svelte';
	import { can } from '$lib/permissions';
	import { keys, languagesOptions } from '$lib/query';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	const list = createQuery(() =>
		languagesOptions({ languages: data.languages, apps: data.apps, shipped: data.shipped })
	);

	const canWrite = $derived(can(data.admin, 'languages.write'));

	let search = $state('');

	/** There are as many languages as somebody has added — a handful — so the
	    search narrows them here rather than asking the server again. */
	const visible = $derived.by(() => {
		const term = search.trim().toLowerCase();
		if (term === '') return list.data.languages;

		return list.data.languages.filter((language) =>
			[language.name, language.native, language.code].some((value) =>
				value.toLowerCase().includes(term)
			)
		);
	});

	let editing = $state<Language | null>(null);
	let drawerTab = $state<'settings' | LocaleApp>('settings');
	let drawerOpen = $state(false);
	let creating = $state(false);

	function open(language: Language, tab: 'settings' | LocaleApp = 'settings') {
		editing = language;
		drawerTab = tab;
		drawerOpen = true;
	}

	/** A language that was just added opens on its text, which is the next
	    thing anybody adding one does — unless it arrived complete. */
	function created(language: Language) {
		const unfinished = language.apps.find((app) => (language.missing[app] ?? 0) > 0);
		open(language, unfinished ?? 'settings');
	}

	let refreshing = $state(false);

	async function refresh() {
		refreshing = true;

		try {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.languages.all }),
				new Promise((done) => setTimeout(done, 400))
			]);
		} finally {
			refreshing = false;
		}
	}
</script>

<svelte:head><title>Languages · {BRAND.name}</title></svelte:head>

<div class="heading">
	<PageHeader crumbs={['Dashboard', 'Languages']}>
		{#snippet secondary()}
			<span class="total">{`${list.data.languages.length} available`}</span>

			<IconButton
				icon={RiRefreshLine}
				label="Refresh the data"
				onclick={refresh}
				loading={refreshing}
				disabled={refreshing}
			/>
		{/snippet}

		{#snippet actions()}
			{#if canWrite}
				<Button onclick={() => (creating = true)}>
					<Icon icon={RiAddLine} />
					New language
				</Button>
			{/if}
		{/snippet}
	</PageHeader>
</div>

<p class="lead">
	The languages the sign-in pages can be shown in. Their text lives in the database: add a language,
	translate it here or import a file, and it is on the next page anybody opens — nothing is rebuilt.
	This panel is English and is not one of them.
</p>

<div class="toolbar">
	<SearchInput label="Search" placeholder="Search name or code…" bind:value={search} />
</div>

<LanguageTable
	languages={visible}
	apps={list.data.apps}
	onOpen={open}
	empty="No languages match this."
/>

<LanguageDrawer bind:open={drawerOpen} language={editing} tab={drawerTab} {canWrite} />

<NewLanguageDrawer
	bind:open={creating}
	languages={list.data.languages}
	shipped={list.data.shipped}
	onCreated={created}
/>

<style>
	.heading {
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}

	.lead {
		max-width: 90ch;
		margin: 0 0 var(--space-3);
		padding-inline: var(--page-gutter);
		color: var(--color-text-hint);
		line-height: 1.5;
	}

	.toolbar {
		display: flex;
		gap: var(--space-2);
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}
</style>
