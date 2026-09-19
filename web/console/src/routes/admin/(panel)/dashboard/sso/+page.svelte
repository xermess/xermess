<script lang="ts">
	import { RiAddLine, RiBuilding2Line, RiRefreshLine, RiSearchLine } from 'svelte-remixicon';
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import type { SSOConnection } from '$lib/api';
	import { Button, Icon, IconButton } from '$lib/components/ui';
	import SSODrawer from '$lib/components/sso/SSODrawer.svelte';
	import SSOTable from '$lib/components/sso/SSOTable.svelte';
	import { useTranslator } from '$lib/i18n';
	import { can } from '$lib/permissions';
	import { keys, ssoOptions } from '$lib/query';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();
	const t = useTranslator();

	const list = createQuery(() => ssoOptions({ connections: data.connections }));

	const canWrite = $derived(can(data.admin, 'sso.write'));

	let search = $state('');

	/** There are as many connections as organisations — a handful — so the
	    search narrows them here, by name, identifier or domain. */
	const visible = $derived.by(() => {
		const term = search.trim().toLowerCase();
		if (term === '') return list.data.connections;

		return list.data.connections.filter((connection) =>
			[connection.name, connection.slug, ...connection.domains].some((value) =>
				value.toLowerCase().includes(term)
			)
		);
	});

	let editing = $state<SSOConnection | null>(null);
	let drawerOpen = $state(false);

	function open(connection: SSOConnection | null) {
		editing = connection;
		drawerOpen = true;
	}

	let refreshing = $state(false);

	async function refresh() {
		refreshing = true;

		try {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.sso.all }),
				new Promise((done) => setTimeout(done, 400))
			]);
		} finally {
			refreshing = false;
		}
	}
</script>

<svelte:head><title>{t('sso.head')} · xermess admin</title></svelte:head>

<header>
	<div class="title">
		<h1>{t('sso.title')}</h1>
		<span class="total">{t('sso.total', { count: list.data.connections.length })}</span>

		<IconButton
			icon={RiRefreshLine}
			label={t('action.refresh')}
			onclick={refresh}
			loading={refreshing}
			disabled={refreshing}
		/>
	</div>

	{#if canWrite}
		<Button onclick={() => open(null)}>
			<Icon icon={RiAddLine} />
			{t('sso.new')}
		</Button>
	{/if}
</header>

<p class="lead">{t('sso.lead')}</p>

{#if list.data.connections.length === 0}
	<section class="empty">
		<span class="mark" aria-hidden="true"><Icon icon={RiBuilding2Line} size="1.6rem" /></span>
		<div>
			<h2>{t('sso.empty_title')}</h2>
			<p>{t('sso.empty_body')}</p>
			{#if canWrite}
				<Button onclick={() => open(null)}>
					<Icon icon={RiAddLine} />
					{t('sso.new')}
				</Button>
			{/if}
		</div>
	</section>
{:else}
	<div class="toolbar">
		<label class="search">
			<Icon icon={RiSearchLine} />
			<input
				type="search"
				placeholder={t('sso.search')}
				bind:value={search}
				aria-label={t('action.search')}
			/>
		</label>
	</div>

	<SSOTable connections={visible} onOpen={open} empty={t('sso.no_match')} />
{/if}

<SSODrawer
	bind:open={drawerOpen}
	connection={editing}
	roles={data.roles}
	canReadRoles={data.mayReadRoles}
	{canWrite}
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
		font-size: var(--text-base);
	}

	.search input:focus {
		outline: none;
	}

	.empty {
		display: flex;
		align-items: flex-start;
		gap: var(--space-4);
		max-width: 60rem;
		margin-inline: var(--page-gutter);
		padding: var(--space-5) var(--space-4);
		border: 1px dashed var(--color-border);
		border-radius: var(--radius-lg);
	}

	.empty h2 {
		margin: 0;
		font-size: var(--text-lg);
	}

	.empty p {
		max-width: 40rem;
		margin: var(--space-1) 0 var(--space-3);
		color: var(--color-text-hint);
		line-height: 1.55;
	}

	.mark {
		display: grid;
		flex-shrink: 0;
		place-items: center;
		width: 56px;
		height: 56px;
		border-radius: 16px;
		background: var(--color-primary);
		color: var(--color-primary-text);
	}
</style>
