<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { RiAddLine, RiDeleteBinLine, RiRefreshLine } from 'svelte-remixicon';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { ApiError, applicationsApi, type Application } from '$lib/api';
	import { BRAND } from '$lib/brand';
	import { applicationsOptions, keys } from '$lib/query';
	import { can } from '$lib/permissions';
	import {
		Alert,
		Button,
		Icon,
		IconButton,
		PageHeader,
		SearchInput,
		SegmentedControl,
		SelectionBar,
		Toolbar
	} from '$lib/components/ui';
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

<svelte:head><title>Applications · {BRAND.name}</title></svelte:head>

<PageHeader
	crumbs={['Dashboard', 'Applications']}
	count={applications.data.total}
	description="The apps and services that sign their users in here with OAuth 2.0 and OpenID Connect. Each defines its own roles, and a token for it carries its roles alone."
>
	{#snippet secondary()}
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

<Toolbar>
	<SearchInput
		label="Search applications"
		placeholder="Search name, description or client id…"
		bind:value={search}
		onsubmit={() => apply({ search })}
		oninput={debounced}
	/>

	<SegmentedControl
		label="Filter by type"
		options={filters}
		value={data.type}
		onChange={(type) => apply({ type })}
	/>
</Toolbar>

{#if error}
	<Alert>{error}</Alert>
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
	.warning {
		color: var(--color-danger);
		font-size: var(--text-sm);
	}
</style>
