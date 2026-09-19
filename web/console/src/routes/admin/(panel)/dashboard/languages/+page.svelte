<script lang="ts">
	import { RiAddLine, RiRefreshLine, RiSearchLine } from 'svelte-remixicon';
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import type { Language, LocaleApp } from '$lib/api';
	import { Button, Icon, IconButton } from '$lib/components/ui';
	import LanguageDrawer from '$lib/components/languages/LanguageDrawer.svelte';
	import LanguageTable from '$lib/components/languages/LanguageTable.svelte';
	import NewLanguageDrawer from '$lib/components/languages/NewLanguageDrawer.svelte';
	import { useTranslator } from '$lib/i18n';
	import { can } from '$lib/permissions';
	import { keys, languagesOptions } from '$lib/query';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();
	const t = useTranslator();

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

<svelte:head><title>{t('languages.head')} · xermess admin</title></svelte:head>

<header>
	<div class="title">
		<h1>{t('languages.title')}</h1>
		<span class="total">{t('languages.total', { count: list.data.languages.length })}</span>

		<IconButton
			icon={RiRefreshLine}
			label={t('action.refresh')}
			onclick={refresh}
			loading={refreshing}
			disabled={refreshing}
		/>
	</div>

	{#if canWrite}
		<div class="actions">
			<Button onclick={() => (creating = true)}>
				<Icon icon={RiAddLine} />
				{t('languages.new')}
			</Button>
		</div>
	{/if}
</header>

<p class="lead">{t('languages.lead')}</p>

<div class="toolbar">
	<label class="search">
		<Icon icon={RiSearchLine} />
		<input
			type="search"
			placeholder={t('languages.search')}
			bind:value={search}
			aria-label={t('action.search')}
		/>
	</label>
</div>

<LanguageTable
	languages={visible}
	apps={list.data.apps}
	onOpen={open}
	empty={t('languages.empty')}
/>

<LanguageDrawer bind:open={drawerOpen} language={editing} tab={drawerTab} {canWrite} />

<NewLanguageDrawer
	bind:open={creating}
	languages={list.data.languages}
	shipped={list.data.shipped}
	onCreated={created}
/>

<style>
	header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		margin-bottom: var(--space-2);
		padding-inline: var(--page-gutter);
	}

	.title {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	.total {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.actions {
		display: flex;
		gap: var(--space-2);
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

	.search {
		flex: 1;
		display: flex;
		align-items: center;
		gap: var(--space-2);
		height: var(--control-height);
		padding: 0 13px;
		border-radius: var(--radius-sm);
		background: var(--color-input);
		color: var(--color-text-hint);
		transition: background-color var(--speed-fast);
	}

	.search:focus-within {
		background: var(--color-input-focus);
	}

	.search input {
		flex: 1;
		min-width: 0;
		border: none;
		background: transparent;
		color: var(--color-text);
		font: inherit;
		font-family: var(--font-sans);
		font-size: var(--text-base);
	}

	.search input:focus {
		outline: none;
	}
</style>
