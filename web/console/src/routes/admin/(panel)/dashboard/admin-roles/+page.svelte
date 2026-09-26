<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { RiAddLine, RiDeleteBinLine, RiRefreshLine } from 'svelte-remixicon';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { ApiError, adminsApi, type AdminRole } from '$lib/api';
	import { BRAND } from '$lib/brand';
	import { adminPermissionsOptions, adminRolesOptions, keys } from '$lib/query';
	import {
		Alert,
		Button,
		Icon,
		IconButton,
		PageHeader,
		SearchInput,
		SelectionBar,
		Toolbar
	} from '$lib/components/ui';
	import AdminRoleDrawer from '$lib/components/admins/AdminRoleDrawer.svelte';
	import AdminRoleTable from '$lib/components/admins/AdminRoleTable.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	const roles = createQuery(() => adminRolesOptions(data.search, data.roles));
	const catalog = createQuery(() => adminPermissionsOptions(data.catalog));

	let search = $derived(data.search);

	let editing = $state<AdminRole | null>(null);
	let roleOpen = $state(false);

	let selected = $state<string[]>([]);

	/** The ticked rows on screen that can be deleted: super_admin is built in,
	    so ticking it offers nothing. */
	const deletable = $derived(
		new Set(roles.data.filter((role) => !role.builtin).map((role) => role.id))
	);
	const chosen = $derived(selected.filter((id) => deletable.has(id)));

	/** How many administrators lose a role if the chosen ones go. */
	const holders = $derived(
		roles.data
			.filter((role) => chosen.includes(role.id))
			.reduce((sum, role) => sum + role.admin_count, 0)
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
		const path = resolve('/admin/(panel)/dashboard/admin-roles');

		// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto(query ? `${path}?${query}` : path, { keepFocus: true, noScroll: true });
	}

	let timer: ReturnType<typeof setTimeout>;

	function debounced() {
		clearTimeout(timer);
		timer = setTimeout(() => apply({ search }), 250);
	}

	function openRole(role: AdminRole | null) {
		editing = role;
		roleOpen = true;
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
				await adminsApi.removeRole(id);
			}
		},
		onSuccess: () => {
			reset();
			return queryClient.invalidateQueries({ queryKey: keys.admins.all });
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not delete these roles';
		},
		onSettled: () => {
			busy = false;
		}
	}));
</script>

<svelte:head><title>Admin roles · {BRAND.name}</title></svelte:head>

<PageHeader
	crumbs={['Dashboard', 'Admin roles']}
	count={roles.data.length}
	description="What administrators may do in this panel. Each role grants permissions from a fixed catalog; managing administrators and these roles stays with super_admin."
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
		<Button onclick={() => openRole(null)}>
			<Icon icon={RiAddLine} />
			New admin role
		</Button>
	{/snippet}
</PageHeader>

<Toolbar>
	<SearchInput
		label="Search admin roles"
		placeholder="Search name or description…"
		bind:value={search}
		onsubmit={() => apply({ search })}
		oninput={debounced}
	/>
</Toolbar>

{#if error}
	<Alert>{error}</Alert>
{/if}

<AdminRoleTable
	roles={roles.data}
	onOpen={openRole}
	selected={chosen}
	onSelect={(ids) => (selected = ids)}
/>

<SelectionBar count={chosen.length} onReset={reset}>
	{#if confirmingDelete}
		{#if holders > 0}
			<span class="warning">
				{holders}
				{holders === 1 ? 'administrator' : 'administrators'} will lose a role
			</span>
		{/if}
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

<AdminRoleDrawer role={editing} catalog={catalog.data} bind:open={roleOpen} />

<style>
	.warning {
		color: var(--color-danger);
		font-size: var(--text-sm);
	}
</style>
