<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import { RiAddLine } from 'svelte-remixicon';
	import type { Language, LanguageList, LocaleApp } from '$lib/api';
	import { Button, Icon, SearchInput } from '$lib/components/ui';
	import { languagesOptions } from '$lib/query';
	import LanguageDrawer from './LanguageDrawer.svelte';
	import LanguageTable from './LanguageTable.svelte';
	import NewLanguageDrawer from './NewLanguageDrawer.svelte';

	type Props = {
		/** What the server rendered, to seed the query with. */
		initial: LanguageList;
		/** Whether the administrator may add, change or remove a language. */
		canWrite: boolean;
	};

	let { initial, canWrite }: Props = $props();

	const list = createQuery(() => languagesOptions(initial));

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
</script>

<div class="languages">
	<p class="lead">
		The languages the sign-in pages can be shown in. Their text lives in the database: add a
		language, translate it here or import a file, and it is on the next page anybody opens — nothing
		is rebuilt. This panel is English and is not one of them.
	</p>

	<div class="toolbar">
		<SearchInput label="Search" placeholder="Search name or code…" bind:value={search} />

		<span class="total">{`${list.data.languages.length} available`}</span>

		{#if canWrite}
			<Button onclick={() => (creating = true)}>
				<Icon icon={RiAddLine} />
				New language
			</Button>
		{/if}
	</div>

	<LanguageTable
		languages={visible}
		apps={list.data.apps}
		onOpen={open}
		empty="No languages match this."
	/>
</div>

<LanguageDrawer bind:open={drawerOpen} language={editing} tab={drawerTab} {canWrite} />

<NewLanguageDrawer
	bind:open={creating}
	languages={list.data.languages}
	shipped={list.data.shipped}
	onCreated={created}
/>

<style>
	.languages {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}

	.lead {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		line-height: 1.5;
	}

	.toolbar {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	.total {
		margin-left: auto;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		white-space: nowrap;
	}
</style>
