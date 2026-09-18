<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { RiArrowLeftLine, RiArrowRightLine, RiRefreshLine } from 'svelte-remixicon';
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { Button, Icon, IconButton, LinkButton } from '$lib/components/ui';
	import RowGrid from '$lib/components/database/RowGrid.svelte';
	import { describes } from '$lib/components/database/format';
	import { DATABASE_PAGE_SIZE, databaseTableOptions, keys } from '$lib/query';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	const table = createQuery(() => databaseTableOptions(data.page.table, data.offset, data.page));

	const page = $derived(table.data);

	/** Which rows this page holds, counting from one, for the line under the
	    pager. An empty table says so rather than showing "0 to 0". */
	const first = $derived(page.offset + 1);
	const last = $derived(Math.min(page.offset + page.rows.length, page.total));

	const hidden = $derived(page.columns.filter((column) => column.hidden).length);

	async function to(offset: number) {
		const path = resolve('/admin/(panel)/dashboard/database/[table]', { table: page.table });

		// resolve() has already applied any base path; the query is only ever
		// appended to what it returned.
		// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto(offset > 0 ? `${path}?offset=${offset}` : path, { noScroll: true });
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

<svelte:head><title>{page.table} · Database · xermess admin</title></svelte:head>

<header>
	<div class="title">
		<LinkButton
			href={resolve('/admin/(panel)/dashboard/database')}
			variant="ghost"
			size="sm"
			aria-label="Back to the tables"
		>
			<Icon icon={RiArrowLeftLine} />
			Database
		</LinkButton>

		<h1>{page.table}</h1>
		<span class="total">
			{page.total.toLocaleString()}
			{page.total === 1 ? 'row' : 'rows'}
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
	{describes[page.table] ?? 'One of the tables this server keeps its records in'}.
	{#if hidden > 0}
		{hidden}
		{hidden === 1 ? 'column holds' : 'columns hold'} a password, a key or a token, and
		{hidden === 1 ? 'is' : 'are'} never read.
	{/if}
</p>

<RowGrid columns={page.columns} rows={page.rows} empty="This table is empty." />

{#if page.total > DATABASE_PAGE_SIZE}
	<div class="pager">
		<span class="range">
			{#if page.rows.length > 0}
				{first.toLocaleString()}–{last.toLocaleString()} of {page.total.toLocaleString()}
			{:else}
				Past the end of the table.
			{/if}
		</span>

		<div class="buttons">
			<Button
				variant="subtle"
				size="sm"
				disabled={page.offset === 0}
				onclick={() => to(Math.max(page.offset - DATABASE_PAGE_SIZE, 0))}
			>
				<Icon icon={RiArrowLeftLine} />
				Newer
			</Button>
			<Button
				variant="subtle"
				size="sm"
				disabled={page.offset + page.rows.length >= page.total}
				onclick={() => to(page.offset + DATABASE_PAGE_SIZE)}
			>
				Older
				<Icon icon={RiArrowRightLine} />
			</Button>
		</div>
	</div>
{/if}

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

	h1 {
		font-family: var(--font-mono);
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

	.pager {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		padding: var(--space-3) var(--page-gutter);
	}

	.range {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		font-variant-numeric: tabular-nums;
	}

	.buttons {
		display: flex;
		gap: var(--space-2);
	}
</style>
