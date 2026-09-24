<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import {
		RiAddLine,
		RiCloseLine,
		RiDeleteBinLine,
		RiRefreshLine,
		RiSearchLine
	} from 'svelte-remixicon';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { ApiError, adminsApi, type AdminRecord } from '$lib/api';
	import {
		adminPermissionsOptions,
		adminRolesOptions,
		adminsOptions,
		applicationChoicesOptions,
		keys
	} from '$lib/query';
	import { Alert, Button, Icon, IconButton, PageHeader, SelectionBar } from '$lib/components/ui';
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

<svelte:head><title>Administrators · xermess admin</title></svelte:head>

<div class="heading">
	<PageHeader crumbs={['Dashboard', 'Administrators']}>
		{#snippet secondary()}
			<span class="total">{admins.data.total} total</span>

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
</div>

<div class="gutter policy">
	<SecurityPanel security={data.security} selfHasMFA={data.admin.mfa_enabled} />
</div>

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
			placeholder="Search email or name…"
			bind:value={search}
			oninput={debounced}
			aria-label="Search administrators"
		/>
	</form>

	{#if data.role}
		<button
			type="button"
			class="chip"
			title="Clear the role filter"
			onclick={() => apply({ role: '' })}
		>
			role: <strong>{filterRole?.name ?? 'unknown'}</strong>
			<Icon icon={RiCloseLine} />
		</button>
	{/if}

	<div class="filter" role="group" aria-label="Filter by status">
		{#each filters as filter (filter.value)}
			<button
				type="button"
				class:selected={data.status === filter.value}
				aria-pressed={data.status === filter.value}
				onclick={() => apply({ status: filter.value })}
			>
				{filter.label}
			</button>
		{/each}
	</div>
</div>

{#if error}
	<div class="gutter error"><Alert>{error}</Alert></div>
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
	/* The panel's own security sits above the list of who has it. */
	.policy {
		margin-bottom: var(--space-4);
	}

	.heading {
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}

	.error {
		margin-bottom: var(--space-3);
	}

	.warning {
		color: var(--color-text-hint);
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

	.chip {
		display: flex;
		align-items: center;
		gap: var(--space-1);
		height: var(--control-height);
		padding: 0 var(--space-3);
		border: none;
		border-radius: var(--radius-md);
		background: var(--surface-info);
		color: var(--color-text);
		font: inherit;
		font-size: var(--text-base);
		white-space: nowrap;
		cursor: pointer;
	}

	.chip strong {
		font-family: var(--font-mono);
		font-size: var(--text-sm);
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
