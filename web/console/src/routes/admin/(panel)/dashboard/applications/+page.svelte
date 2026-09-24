<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { RiAddLine, RiDeleteBinLine, RiRefreshLine, RiSearchLine } from 'svelte-remixicon';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { ApiError, applicationsApi, type Application } from '$lib/api';
	import { applicationsOptions, keys } from '$lib/query';
	import { can } from '$lib/permissions';
	import { Alert, Button, Icon, IconButton, PageHeader, SelectionBar } from '$lib/components/ui';
	import ApplicationDrawer from '$lib/components/applications/ApplicationDrawer.svelte';
	import ApplicationTable from '$lib/components/applications/ApplicationTable.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	const applications = createQuery(() =>
		applicationsOptions({ search: data.search, type: data.type }, data.page)
	);

	/** Registering and removing applications takes applications.write for the
	    whole panel; changing one takes it for that application. */
	const canRegister = $derived(can(data.admin, 'applications.write'));

	let search = $derived(data.search);

	let editing = $state<Application | null>(null);
	let drawerOpen = $state(false);

	let selected = $state<string[]>([]);
	const visible = $derived(new Set(applications.data.applications.map((app) => app.id)));
	const chosen = $derived(selected.filter((id) => visible.has(id)));

	let confirmingDelete = $state(false);
	let error = $state('');
	let busy = $state(false);

	async function apply(changes: { search?: string; type?: string }) {
		const params = new SvelteURLSearchParams(page.url.searchParams);

		for (const [key, value] of Object.entries(changes)) {
			if (value) params.set(key, value);
			else params.delete(key);
		}

		const query = params.toString();
		const path = resolve('/admin/(panel)/dashboard/applications');

		// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto(query ? `${path}?${query}` : path, { keepFocus: true, noScroll: true });
	}

	let timer: ReturnType<typeof setTimeout>;

	function debounced() {
		clearTimeout(timer);
		timer = setTimeout(() => apply({ search }), 250);
	}

	function openApplication(application: Application | null) {
		editing = application;
		drawerOpen = true;
	}

	let refreshing = $state(false);

	async function refresh() {
		refreshing = true;

		try {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.applications.all }),
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

	/** Removing an application takes its roles, and everyone's hold on them,
	    with it, so roles and users are refilled too. */
	const removeSelected = createMutation(() => ({
		mutationFn: async (ids: string[]) => {
			for (const id of ids) {
				await applicationsApi.remove(id);
			}
		},
		onSuccess: () => {
			reset();
			return Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.applications.all }),
				queryClient.invalidateQueries({ queryKey: keys.roles.all }),
				queryClient.invalidateQueries({ queryKey: keys.users.all })
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not delete these applications';
		},
		onSettled: () => {
			busy = false;
		}
	}));

	const filters = [
		{ value: '', label: 'All' },
		{ value: 'web', label: 'Web' },
		{ value: 'spa', label: 'SPA' },
		{ value: 'native', label: 'Native' },
		{ value: 'm2m', label: 'M2M' }
	];
</script>

<svelte:head><title>Applications · xermess admin</title></svelte:head>

<div class="heading">
	<PageHeader crumbs={['Dashboard', 'Applications']}>
		{#snippet secondary()}
			<span class="total">{applications.data.total} total</span>

			<IconButton
				icon={RiRefreshLine}
				label="Refresh the data"
				onclick={refresh}
				loading={refreshing}
				disabled={refreshing}
			/>
		{/snippet}

		{#snippet actions()}
			{#if canRegister}
				<Button onclick={() => openApplication(null)}>
					<Icon icon={RiAddLine} />
					New application
				</Button>
			{/if}
		{/snippet}
	</PageHeader>
</div>

<p class="lead">
	The apps and services that sign their users in here with OAuth 2.0 and OpenID Connect. Each
	defines its own roles, and a token for it carries its roles alone.
</p>

<div class="toolbar">
	<form
		class="search"
		onsubmit={(event) => {
			event.preventDefault();
			apply({ search });
		}}
	>
		<Icon icon={RiSearchLine} />
		<input
			type="search"
			placeholder="Search name, description or client id…"
			bind:value={search}
			oninput={debounced}
			aria-label="Search applications"
		/>
	</form>

	<div class="filter" role="group" aria-label="Filter by type">
		{#each filters as filter (filter.value)}
			<button
				type="button"
				class:selected={data.type === filter.value}
				aria-pressed={data.type === filter.value}
				onclick={() => apply({ type: filter.value })}
			>
				{filter.label}
			</button>
		{/each}
	</div>
</div>

{#if error}
	<div class="gutter error"><Alert>{error}</Alert></div>
{/if}

<ApplicationTable
	applications={applications.data.applications}
	onOpen={openApplication}
	selected={canRegister ? chosen : undefined}
	onSelect={canRegister ? (ids) => (selected = ids) : undefined}
/>

<SelectionBar count={chosen.length} onReset={reset}>
	{#if confirmingDelete}
		<span class="warning">Their roles, and everyone's hold on them, go too.</span>
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

<ApplicationDrawer
	application={editing}
	editable={editing ? can(data.admin, 'applications.write', editing.id) : canRegister}
	admin={data.admin}
	bind:open={drawerOpen}
/>

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

	/* The same height as the search box beside it and the buttons above it:
	   the toolbar reads as one row rather than three sizes. */
	.filter {
		display: flex;
		gap: 2px;
		height: var(--control-height);
		padding: 3px;
		border-radius: var(--radius-md);
		background: var(--color-input);
	}

	.filter button {
		padding: 0 var(--space-4);
		border: none;
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--color-text-hint);
		font: inherit;
		font-size: var(--text-base);
		cursor: pointer;
		transition:
			background-color var(--speed-fast),
			color var(--speed-fast);
	}

	.filter button:hover {
		color: var(--color-text);
	}

	.filter button.selected {
		background: var(--color-surface);
		color: var(--color-text);
		font-weight: 600;
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
