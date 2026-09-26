<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import {
		RiAddLine,
		RiCloseLine,
		RiDeleteBinLine,
		RiDownloadLine,
		RiRefreshLine,
		RiSettings3Line
	} from 'svelte-remixicon';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { ApiError, usersApi, type UserRecord } from '$lib/api';
	import { BRAND } from '$lib/brand';
	import {
		applicationChoicesOptions,
		keys,
		roleChoicesOptions,
		userFieldsOptions,
		usersOptions
	} from '$lib/query';
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
	import FieldsDrawer from '$lib/components/users/FieldsDrawer.svelte';
	import UserDrawer from '$lib/components/users/UserDrawer.svelte';
	import UserTable from '$lib/components/users/UserTable.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	// The list and the fields are queries, seeded with what the server has
	// already rendered: the first paint costs no request, and everything
	// after — a save, the refresh button, coming back to the tab — is the
	// cache being refilled rather than the page being reloaded.
	const users = createQuery(() =>
		usersOptions({ search: data.search, verified: data.verified, role: data.role }, data.page)
	);
	const fields = createQuery(() => userFieldsOptions(data.fields));
	const roles = createQuery(() => roleChoicesOptions(data.roles));
	const applications = createQuery(() => applicationChoicesOptions(data.applications));

	/** What this administrator may change. Without users.write the page is
	    for looking; without user_fields.write the field settings are hidden. */
	const canWrite = $derived(can(data.admin, 'users.write'));
	const canEditFields = $derived(can(data.admin, 'user_fields.write'));

	/** The role the list is filtered to, when the URL names one. */
	const filterRole = $derived(roles.data.find((role) => role.id === data.role));
	const filterApp = $derived(
		applications.data.find((app) => app.id === filterRole?.application_id)
	);

	// A writable derived: typing updates it, and it goes back to following the
	// URL whenever that changes, so the back button and a shared link both put
	// the right term in the box.
	let search = $derived(data.search);

	let editing = $state<UserRecord | null>(null);
	let userOpen = $state(false);
	let fieldsOpen = $state(false);

	/** The rows that are ticked, by id. */
	let selected = $state<string[]>([]);

	/** The ticked rows that are actually on screen.
	
	    A search or a filter can take a ticked row out of view, and deleting
	    what nobody can see is not something a panel should offer. Everything
	    the bar says and does goes through this rather than through the raw
	    list, so what is counted is what is shown. */
	const visible = $derived(new Set(users.data.users.map((user) => user.id)));
	const chosen = $derived(selected.filter((id) => visible.has(id)));
	let confirmingDelete = $state(false);
	let error = $state('');

	/** True while the chosen records are being deleted. Ours rather than the
	    mutation's own isPending, so the bar cannot be left disabled. */
	let busy = $state(false);

	/** The search and the filter are the URL, so the server renders the
	    result and the back button walks through it. */
	async function apply(changes: { search?: string; verified?: string; role?: string }) {
		const params = new SvelteURLSearchParams(page.url.searchParams);

		for (const [key, value] of Object.entries(changes)) {
			if (value) params.set(key, value);
			else params.delete(key);
		}

		const query = params.toString();
		const path = resolve('/admin/(panel)/dashboard/users');

		// resolve() has already applied any base path; the query is only ever
		// appended to what it returned.
		// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto(query ? `${path}?${query}` : path, { keepFocus: true, noScroll: true });
	}

	/** Searching as you type, but only once you have paused: every keystroke
	    is a round trip to the server otherwise. */
	let timer: ReturnType<typeof setTimeout>;

	function debounced() {
		clearTimeout(timer);
		timer = setTimeout(() => apply({ search }), 250);
	}

	function openUser(user: UserRecord | null) {
		editing = user;
		userOpen = true;
	}

	let refreshing = $state(false);

	/** Asks for the rows again, so a record someone else changed shows up
	    without leaving the page.

	    The spin is held for a moment even when the answer comes back at once:
	    a button that does something invisible in 20ms reads as a button that
	    did nothing. */
	async function refresh() {
		refreshing = true;

		try {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.users.all }),
				queryClient.invalidateQueries({ queryKey: keys.roles.choices }),
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

	/** The records behind the ticked ids, in the order the table shows them. */
	function chosenRecords(): UserRecord[] {
		return users.data.users.filter((user) => chosen.includes(user.id));
	}

	/** Deleting what is ticked, one call each — the API removes one record at
	    a time, and a half-finished delete should still leave the list right,
	    which is what refilling the cache afterwards is for. */
	const removeSelected = createMutation(() => ({
		mutationFn: async (ids: string[]) => {
			for (const id of ids) {
				await usersApi.remove(id);
			}
		},
		onSuccess: () => {
			reset();
			// A role's count of users changes with them.
			return Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.users.all }),
				queryClient.invalidateQueries({ queryKey: keys.roles.all })
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not delete these users';
		},
		onSettled: () => {
			busy = false;
		}
	}));

	/** Hands the chosen records to the browser as a file. Nothing leaves the
	    machine: the JSON is built here from what the page already has. */
	function download() {
		const blob = new Blob([JSON.stringify(chosenRecords(), null, 2)], { type: 'application/json' });
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');

		link.href = url;
		link.download = `users-${new Date().toISOString().slice(0, 10)}.json`;
		link.click();

		URL.revokeObjectURL(url);
	}

	const filters = [
		{ value: '', label: 'All' },
		{ value: 'true', label: 'Verified' },
		{ value: 'false', label: 'Unverified' }
	];
</script>

<svelte:head><title>Users · {BRAND.name}</title></svelte:head>

<div class="heading">
	<PageHeader crumbs={['Dashboard', 'Users']}>
		{#snippet secondary()}
			<span class="total">{users.data.total} total</span>

			{#if canEditFields}
				<IconButton
					icon={RiSettings3Line}
					label="Field settings"
					onclick={() => (fieldsOpen = true)}
				/>
			{/if}

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
				<Button onclick={() => openUser(null)}>
					<Icon icon={RiAddLine} />
					New user
				</Button>
			{/if}
		{/snippet}
	</PageHeader>
</div>

<div class="toolbar">
	<SearchInput
		label="Search users"
		placeholder="Search email or any field…"
		bind:value={search}
		onsubmit={() => apply({ search })}
		oninput={debounced}
	/>

	{#if data.role}
		<!-- Arriving from the roles page filters the list to one role's
		     holders; this says so, and is the way back to everyone. -->
		<button
			type="button"
			class="chip"
			title="Clear the role filter"
			onclick={() => apply({ role: '' })}
		>
			role: <strong>{filterRole?.name ?? 'unknown'}</strong>
			{#if filterApp}in {filterApp.name}{/if}
			<Icon icon={RiCloseLine} />
		</button>
	{/if}

	<div class="filter" role="group" aria-label="Filter by verified">
		{#each filters as filter (filter.value)}
			<button
				type="button"
				class:selected={data.verified === filter.value}
				aria-pressed={data.verified === filter.value}
				onclick={() => apply({ verified: filter.value })}
			>
				{filter.label}
			</button>
		{/each}
	</div>
</div>

{#if error}
	<div class="gutter error"><Alert>{error}</Alert></div>
{/if}

<UserTable
	users={users.data.users}
	fields={fields.data}
	applications={applications.data}
	onOpen={openUser}
	selected={canWrite ? chosen : undefined}
	onSelect={canWrite ? (ids) => (selected = ids) : undefined}
/>

<SelectionBar count={chosen.length} onReset={reset}>
	{#if confirmingDelete}
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
		<Button size="sm" onclick={download} disabled={busy}>
			<Icon icon={RiDownloadLine} />
			JSON
		</Button>
	{/if}
</SelectionBar>

<UserDrawer
	user={editing}
	fields={fields.data}
	applications={applications.data}
	roles={roles.data}
	admin={data.admin}
	editable={canWrite}
	bind:open={userOpen}
/>
{#if canEditFields}
	<FieldsDrawer fields={fields.data} bind:open={fieldsOpen} />
{/if}

<style>
	.heading {
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}

	.error {
		margin-bottom: var(--space-3);
	}

	.toolbar {
		display: flex;
		gap: var(--space-2);
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
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
		border-radius: var(--radius-sm);
		background: var(--color-secondary);
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
