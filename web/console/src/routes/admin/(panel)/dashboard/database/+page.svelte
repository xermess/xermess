<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { RiDatabase2Line, RiListUnordered, RiRefreshLine, RiSearchLine } from 'svelte-remixicon';
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import type { DatabaseTable } from '$lib/api';
	import { type Column, DataTable, Icon, IconButton } from '$lib/components/ui';
	import { describes } from '$lib/components/database/format';
	import { databaseTablesOptions, keys } from '$lib/query';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	const tables = createQuery(() => databaseTablesOptions({ tables: data.tables }));

	/** The DataTable identifies rows by `id`; a table's name is what
	    identifies it, so that is what it is given. */
	type Row = DatabaseTable & { id: string };

	let search = $state('');

	const rows = $derived.by<Row[]>(() => {
		const term = search.trim().toLowerCase();

		return tables.data.tables
			.filter((table) => term === '' || table.name.toLowerCase().includes(term))
			.map((table) => ({ ...table, id: table.name }));
	});

	const total = $derived(tables.data.tables.reduce((sum, table) => sum + table.rows, 0));

	const columns: Column[] = [
		{ key: 'table', label: 'table', icon: RiDatabase2Line, min: '16rem' },
		{ key: 'holds', label: 'holds', min: '20rem' },
		{ key: 'rows', label: 'rows', icon: RiListUnordered, min: '7rem', align: 'end' },
		{ key: 'columns', label: 'columns', min: '8rem', align: 'end' }
	];

	async function open(row: Row) {
		await goto(resolve('/admin/(panel)/dashboard/database/[table]', { table: row.name }));
	}

	let refreshing = $state(false);

	async function refresh() {
		refreshing = true;

		try {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.database.all }),
				new Promise((done) => setTimeout(done, 400))
			]);
		} finally {
			refreshing = false;
		}
	}
</script>

<svelte:head><title>Database · xermess admin</title></svelte:head>

<header>
	<div class="title">
		<h1>Database</h1>
		<span class="total">
			{tables.data.tables.length} tables · {total.toLocaleString()} rows
		</span>

		<IconButton
			icon={RiRefreshLine}
			label="Refresh the data"
			onclick={refresh}
			loading={refreshing}
			disabled={refreshing}
		/>
	</div>
</header>

<p class="lead">
	What the migration built, as the database holds it. It is read only: a record is changed on the
	page that knows what it is. Passwords, keys and tokens are never read.
</p>

<div class="toolbar">
	<label class="search">
		<Icon icon={RiSearchLine} />
		<input
			type="search"
			placeholder="Search tables…"
			bind:value={search}
			aria-label="Search tables"
		/>
	</label>
</div>

<DataTable
	{columns}
	{rows}
	onOpen={open}
	label={(row) => `Open ${row.name}`}
	empty={tables.data.tables.length === 0 ? 'No tables.' : 'No tables match this.'}
>
	{#snippet row(table)}
		<td><strong class="name">{table.name}</strong></td>
		<td class="holds">{describes[table.name] ?? ''}</td>
		<td class="end count">{table.rows.toLocaleString()}</td>
		<td class="end count">{table.columns}</td>
	{/snippet}
</DataTable>

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
		font-family: var(--font-sans);
		font-size: var(--text-base);
	}

	.search input:focus {
		outline: none;
	}

	.name {
		font-family: var(--font-mono);
	}

	.holds {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.count {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		font-variant-numeric: tabular-nums;
	}
</style>
