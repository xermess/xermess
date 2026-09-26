<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { RiAddLine, RiDeleteBinLine, RiRefreshLine } from 'svelte-remixicon';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { ApiError, adminsApi, type AdminRecord } from '$lib/api';
	import { BRAND } from '$lib/brand';
	import {
		adminPermissionsOptions,
		adminRolesOptions,
		adminsOptions,
		applicationChoicesOptions,
		keys
	} from '$lib/query';
	import {
		Alert,
		Button,
		FilterChip,
		Icon,
		IconButton,
		PageHeader,
		SearchInput,
		SegmentedControl,
		SelectionBar,
		Toolbar
	} from '$lib/components/ui';
	import AdminDrawer from '$lib/components/admins/AdminDrawer.svelte';
	import AdminTable from '$lib/components/admins/AdminTable.svelte';
	import SecurityPanel from '$lib/components/admins/SecurityPanel.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	const admins = createQuery(() =>
		adminsOptions({ search: data.search, status: data.status, role: data.role }, data.page)
	);
	const roles = createQuery(() => adminRolesOptions('', data.roles));
	const catalog = createQuery(() => adminPermissionsOptions(data.catalog));
	const applications = createQuery(() => applicationChoicesOptions(data.applications));

	/** The role the list is filtered to, when the URL names one. */
	const filterRole = $derived(roles.data.find((role) => role.id === data.role));

	let search = $derived(data.search);

	let editing = $state<AdminRecord | null>(null);
	let adminOpen = $state(false);

	let selected = $state<string[]>([]);

	/** The ticked rows on screen that can be deleted. Nobody can delete their
	    own account, so ticking yourself offers nothing. */
	const deletable = $derived(
		new Set(admins.data.admins.filter((admin) => admin.id !== data.admin.id).map((it) => it.id))
	);
	const chosen = $derived(selected.filter((id) => deletable.has(id)));

	let confirmingDelete = $state(false);
	let error = $state('');
	let busy = $state(false);

	async function apply(changes: { search?: string; status?: string; role?: string }) {
		const params = new SvelteURLSearchParams(page.url.searchParams);

		for (const [key, value] of Object.entries(changes)) {
			if (value) params.set(key, value);
			else params.delete(key);
		}

		const query = params.toString();
		const path = resolve('/admin/(panel)/dashboard/admins');

		// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto(query ? `${path}?${query}` : path, { keepFocus: true, noScroll: true });
	}

	let timer: ReturnType<typeof setTimeout>;

	function debounced() {
		clearTimeout(timer);
		timer = setTimeout(() => apply({ search }), 250);
	}

	function openAdmin(admin: AdminRecord | null) {
		editing = admin;
		adminOpen = true;
	}

	let refreshing = $state(false);

	async function refresh() {
		refreshing = true;

		try {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.admins.all }),
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
				await adminsApi.remove(id);
			}
		},
		onSuccess: () => {
			reset();
			return queryClient.invalidateQueries({ queryKey: keys.admins.all });
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not delete these administrators';
		},
		onSettled: () => {
			busy = false;
		}
	}));

	const filters = [
		{ value: '', label: 'All' },
		{ value: 'active', label: 'Active' },
		{ value: 'suspended', label: 'Suspended' },
		{ value: 'disabled', label: 'Disabled' }
	];
</script>

<svelte:head><title>Administrators · {BRAND.name}</title></svelte:head>

<PageHeader crumbs={['Dashboard', 'Administrators']} count={admins.data.total}>
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
		<Button onclick={() => openAdmin(null)}>
			<Icon icon={RiAddLine} />
			New administrator
		</Button>
	{/snippet}
</PageHeader>

<SecurityPanel security={data.security} selfHasMFA={data.admin.mfa_enabled} />

<Toolbar>
	<SearchInput
		label="Search administrators"
		placeholder="Search email or name…"
		bind:value={search}
		onsubmit={() => apply({ search })}
		oninput={debounced}
	/>

	{#if data.role}
		<FilterChip title="Clear the role filter" onclear={() => apply({ role: '' })}>
			role: <strong>{filterRole?.name ?? 'unknown'}</strong>
		</FilterChip>
	{/if}

	<SegmentedControl
		label="Filter by status"
		options={filters}
		value={data.status}
		onChange={(status) => apply({ status })}
	/>
</Toolbar>

{#if error}
	<Alert>{error}</Alert>
{/if}

<AdminTable
	admins={admins.data.admins}
	self={data.admin.id}
	onOpen={openAdmin}
	selected={chosen}
	onSelect={(ids) => (selected = ids)}
/>

<SelectionBar count={chosen.length} onReset={reset}>
	{#if confirmingDelete}
		<span class="warning">Their sessions end and they can no longer sign in.</span>
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

<AdminDrawer
	admin={editing}
	self={data.admin.id}
	roles={roles.data}
	applications={applications.data}
	catalog={catalog.data}
	bind:open={adminOpen}
/>

<style>
	.warning {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}
</style>
