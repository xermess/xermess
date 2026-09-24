<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { RiAddLine, RiDeleteBinLine, RiRefreshLine } from 'svelte-remixicon';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { ApiError, apisApi, type API } from '$lib/api';
	import { apisOptions, keys } from '$lib/query';
	import { can } from '$lib/permissions';
	import {
		Alert,
		Button,
		Icon,
		IconButton,
		PageHeader,
		SearchInput,
		SelectionBar
	} from '$lib/components/ui';
	import ApiDrawer from '$lib/components/apis/ApiDrawer.svelte';
	import ApiTable from '$lib/components/apis/ApiTable.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	const apis = createQuery(() => apisOptions(data.search, data.apis));

	/** Changing APIs is a whole-panel permission. */
	const canWrite = $derived(can(data.admin, 'apis.write'));

	let search = $derived(data.search);

	let drawerOpen = $state(false);

	let selected = $state<string[]>([]);
	const visible = $derived(new Set(apis.data.map((api) => api.id)));
	const chosen = $derived(selected.filter((id) => visible.has(id)));

	/** How many applications lose access if the chosen APIs go. */
	const affected = $derived(
		apis.data
			.filter((api) => chosen.includes(api.id))
			.reduce((sum, api) => sum + api.application_count, 0)
	);

	let confirmingDelete = $state(false);
	let error = $state('');
	let busy = $state(false);

	async function apply(changes: { search?: string }) {
		const params = new SvelteURLSearchParams(page.url.searchParams);

		for (const [key, value] of Object.entries(changes)) {
			if (value) params.set(key, value);
			else params.delete(key);
		}

		const query = params.toString();
		const path = resolve('/admin/(panel)/dashboard/apis');

		// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto(query ? `${path}?${query}` : path, { keepFocus: true, noScroll: true });
	}

	let timer: ReturnType<typeof setTimeout>;

	function debounced() {
		clearTimeout(timer);
		timer = setTimeout(() => apply({ search }), 250);
	}

	/** An API is managed on its own page; the drawer only registers one. */
	function openApi(api: API) {
		goto(resolve('/admin/(panel)/dashboard/apis/[id]', { id: api.id }));
	}

	let refreshing = $state(false);

	async function refresh() {
		refreshing = true;

		try {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.apis.all }),
				new Promise((done) => setTimeout(done, 400))
			]);
		} finally {
			refreshing = false;
		}
	}

	function reset() {
		selected = [];
		confirmingDelete = false;
		error = '';
	}

	const removeSelected = createMutation(() => ({
		mutationFn: async (ids: string[]) => {
			for (const id of ids) {
				await apisApi.remove(id);
			}
		},
		onSuccess: () => {
			reset();
			return Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.apis.all }),
				queryClient.invalidateQueries({ queryKey: keys.roles.all }),
				queryClient.invalidateQueries({ queryKey: keys.applications.all })
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not delete these APIs';
		},
		onSettled: () => {
			busy = false;
		}
	}));
</script>

<svelte:head><title>APIs · xermess admin</title></svelte:head>

<div class="heading">
	<PageHeader crumbs={['Dashboard', 'APIs']}>
		{#snippet secondary()}
			<span class="total">{apis.data.length} total</span>

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
				<Button onclick={() => (drawerOpen = true)}>
					<Icon icon={RiAddLine} />
					New API
				</Button>
			{/if}
		{/snippet}
	</PageHeader>
</div>

<p class="lead">
	APIs represent the protected resources that applications can access. Applications request access
	tokens for an API, and each token is issued for a specific audience.
</p>

<div class="toolbar">
	<SearchInput
		label="Search APIs"
		placeholder="Search name, identifier or description…"
		bind:value={search}
		onsubmit={() => apply({ search })}
		oninput={debounced}
	/>
</div>

{#if error}
	<div class="gutter error"><Alert>{error}</Alert></div>
{/if}

<ApiTable
	apis={apis.data}
	onOpen={openApi}
	selected={canWrite ? chosen : undefined}
	onSelect={canWrite ? (ids) => (selected = ids) : undefined}
/>

<SelectionBar count={chosen.length} onReset={reset}>
	{#if confirmingDelete}
		<span class="warning">
			{affected > 0
				? `${affected} ${affected === 1 ? 'application loses' : 'applications lose'} access, and roles lose these scopes`
				: 'Roles lose these scopes'}
		</span>
		<Button variant="subtle" size="sm" onclick={() => (confirmingDelete = false)} disabled={busy}>
			Keep them
		</Button>
		<Button
			colorPalette="danger"
			size="sm"
			onclick={() => {
				if (busy) return;
				error = '';
				busy = true;
				removeSelected.mutate(chosen);
			}}
			disabled={busy}
		>
			{busy ? 'Deleting…' : `Delete ${chosen.length}`}
		</Button>
	{:else}
		<Button
			colorPalette="danger"
			size="sm"
			onclick={() => (confirmingDelete = true)}
			disabled={busy}
		>
			<Icon icon={RiDeleteBinLine} />
			Delete
		</Button>
	{/if}
</SelectionBar>

{#if canWrite}
	<ApiDrawer bind:open={drawerOpen} />
{/if}

<style>
	.heading {
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}

	.error {
		margin-bottom: var(--space-3);
	}

	.lead {
		margin: 0 0 var(--space-3);
		padding-inline: var(--page-gutter);
		color: var(--color-text-hint);
		font-size: var(--text-base);
	}

	.warning {
		color: var(--color-danger);
		font-size: var(--text-sm);
	}

	.toolbar {
		display: flex;
		gap: var(--space-2);
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}

	@media (max-width: 40rem) {
		.heading,
		.error {
			margin-bottom: var(--space-3);
		}

		.toolbar {
			flex-direction: column;
			align-items: stretch;
		}
	}
</style>
