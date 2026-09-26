<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { RiRefreshLine } from 'svelte-remixicon';
	import { BRAND } from '$lib/brand';
	import ActivityTable from '$lib/components/activity/ActivityTable.svelte';
	import {
		categoryLabels,
		categoryOf,
		describe,
		type Category
	} from '$lib/components/activity/actions';
	import { IconButton, PageHeader, SearchInput, Toolbar } from '$lib/components/ui';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let category = $state<Category | 'all'>('all');
	let search = $state('');
	let refreshing = $state(false);

	/** The categories that have entries, in the order the panel lists them,
	    each with how many. */
	const categories = $derived(
		(Object.keys(categoryLabels) as Category[])
			.map((key) => ({
				key,
				label: categoryLabels[key],
				count: data.logs.filter((event) => categoryOf(event.action) === key).length
			}))
			.filter((it) => it.count > 0)
	);

	const shown = $derived.by(() => {
		const term = search.trim().toLowerCase();

		return data.logs.filter((event) => {
			if (category !== 'all' && categoryOf(event.action) !== category) return false;
			if (!term) return true;

			const said = describe(event);
			return [said.label, said.actor, said.verb, said.subject, said.detail, event.ip, event.action]
				.join(' ')
				.toLowerCase()
				.includes(term);
		});
	});

	async function refresh() {
		refreshing = true;
		try {
			await Promise.all([invalidateAll(), new Promise((done) => setTimeout(done, 400))]);
		} finally {
			refreshing = false;
		}
	}
</script>

<svelte:head>
	<title>Logs · {BRAND.name}</title>
</svelte:head>

<PageHeader
	crumbs={['Dashboard', 'Logs']}
	count={data.logs.length}
	description="Every sign-in and every change administrators made, newest first."
>
	{#snippet secondary()}
		<IconButton
			icon={RiRefreshLine}
			label="Refresh"
			onclick={refresh}
			loading={refreshing}
			disabled={refreshing}
		/>
	{/snippet}
</PageHeader>

<Toolbar>
	<SearchInput
		label="Search the log"
		placeholder="Search who, what or where…"
		bind:value={search}
	/>

	<div class="chips" role="group" aria-label="Filter by kind">
		<button type="button" class:on={category === 'all'} onclick={() => (category = 'all')}>
			All <span>{data.logs.length}</span>
		</button>
		{#each categories as it (it.key)}
			<button type="button" class:on={category === it.key} onclick={() => (category = it.key)}>
				{it.label} <span>{it.count}</span>
			</button>
		{/each}
	</div>
</Toolbar>

<ActivityTable
	events={shown}
	empty={data.logs.length === 0 ? 'Nothing has happened yet.' : 'No entries match.'}
/>

<style>
	.chips {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}

	.chips button {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		height: 32px;
		padding: 0 12px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-pill);
		background: transparent;
		color: var(--color-text-hint);
		font: inherit;
		font-size: var(--text-sm);
		cursor: pointer;
		transition:
			background-color var(--speed-fast),
			color var(--speed-fast);
	}

	.chips button:hover {
		color: var(--color-text);
	}

	.chips button.on {
		border-color: var(--color-primary);
		background: var(--color-primary);
		color: var(--color-primary-text);
	}

	.chips span {
		font-size: var(--text-xs);
		opacity: 0.7;
	}
</style>
