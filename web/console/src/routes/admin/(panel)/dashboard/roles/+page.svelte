<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import {
		RiAddLine,
		RiDeleteBinLine,
		RiAppsLine,
		RiDownloadLine,
		RiGlobalLine,
		RiRefreshLine,
		RiSearchLine
	} from 'svelte-remixicon';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { ApiError, rolesApi, type Role } from '$lib/api';
	import { applicationChoicesOptions, keys, roleChoicesOptions, rolesOptions } from '$lib/query';
	import { canEditRoles } from '$lib/components/roles/roles';
	import { Alert, Button, Icon, IconButton, PageHeader, SelectionBar } from '$lib/components/ui';
	import RoleDrawer from '$lib/components/roles/RoleDrawer.svelte';
	import RoleTable from '$lib/components/roles/RoleTable.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	// The list and every role are queries, seeded with what the server has
	// already rendered: the first paint costs no request, and everything
	// after is the cache being refilled.
	const roles = createQuery(() =>
		rolesOptions(
			{
				scope: data.tab,
				application: data.application,
				search: data.search,
				isDefault: data.isDefault
			},
			data.page
		)
	);
	const choices = createQuery(() => roleChoicesOptions(data.choices));
	const applications = createQuery(() => applicationChoicesOptions(data.applications));

	const globalCount = $derived(choices.data.filter((role) => role.application_id === null).length);
	const appCount = $derived(choices.data.length - globalCount);

	/** Whether this administrator can make a role anywhere at all. */
	const canCreate = $derived(
		canEditRoles(data.admin, null) ||
			applications.data.some((app) => canEditRoles(data.admin, app.id))
	);

	/** Rows can be ticked for deleting when the administrator may change the
	    roles on screen: the global roles, or every application's. Someone who
	    looks after one application removes its roles from the role panel. */
	const canWrite = $derived(canEditRoles(data.admin, null));

	// A writable derived: typing updates it, and it goes back to following the
	// URL whenever that changes.
	let search = $derived(data.search);

	let editing = $state<Role | null>(null);
	let roleOpen = $state(false);

	/** The rows that are ticked, by id. */
	let selected = $state<string[]>([]);

	/** The ticked rows that are actually on screen: a search can take a
	    ticked row out of view, and deleting what nobody can see is not
	    something a panel should offer. */
	const visible = $derived(new Set(roles.data.roles.map((role) => role.id)));
	const chosen = $derived(selected.filter((id) => visible.has(id)));
	let confirmingDelete = $state(false);
	let error = $state('');

	/** True while the chosen roles are being deleted. */
	let busy = $state(false);

	/** How many users lose a role if the chosen ones are deleted, which is
	    worth saying before it happens. */
	const holders = $derived(
		roles.data.roles
			.filter((role) => chosen.includes(role.id))
			.reduce((sum, role) => sum + role.user_count, 0)
	);

	/** The search and the filter are the URL, so the server renders the
	    result and the back button walks through it. */
	async function apply(changes: {
		tab?: string;
		application?: string;
		search?: string;
		default?: string;
	}) {
		const params = new SvelteURLSearchParams(page.url.searchParams);

		for (const [key, value] of Object.entries(changes)) {
			if (value) params.set(key, value);
			else params.delete(key);
		}

		const query = params.toString();
		const path = resolve('/admin/(panel)/dashboard/roles');

		// resolve() has already applied any base path; the query is only ever
		// appended to what it returned.
		// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto(query ? `${path}?${query}` : path, { keepFocus: true, noScroll: true });
	}

	/** Searching as you type, but only once you have paused. */
	let timer: ReturnType<typeof setTimeout>;

	function debounced() {
		clearTimeout(timer);
		timer = setTimeout(() => apply({ search }), 250);
	}

	function openRole(role: Role | null) {
		editing = role;
		roleOpen = true;
	}

	let refreshing = $state(false);

	/** Asks for the rows again. The spin is held for a moment even when the
	    answer comes back at once, so the button reads as having done
	    something. */
	async function refresh() {
		refreshing = true;

		try {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.roles.all }),
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

	/** The roles behind the ticked ids, in the order the table shows them. */
	function chosenRoles(): Role[] {
		return roles.data.roles.filter((role) => chosen.includes(role.id));
	}

	/** Deleting what is ticked, one call each. Users who held a deleted role
	    change too, so their list is refilled as well. */
	const removeSelected = createMutation(() => ({
		mutationFn: async (ids: string[]) => {
			for (const id of ids) {
				await rolesApi.remove(id);
			}
		},
		onSuccess: () => {
			reset();
			return Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.roles.all }),
				queryClient.invalidateQueries({ queryKey: keys.users.all })
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not delete these roles';
		},
		onSettled: () => {
			busy = false;
		}
	}));

	/** Hands the chosen roles to the browser as a file, built from what the
	    page already has. */
	function download() {
		const blob = new Blob([JSON.stringify(chosenRoles(), null, 2)], { type: 'application/json' });
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');

		link.href = url;
		link.download = `roles-${new Date().toISOString().slice(0, 10)}.json`;
		link.click();

		URL.revokeObjectURL(url);
	}

	const filters = [
		{ value: '', label: 'All' },
		{ value: 'true', label: 'Default' },
		{ value: 'false', label: 'Not default' }
	];
</script>

<svelte:head><title>Roles · xermess admin</title></svelte:head>

<div class="heading">
	<PageHeader crumbs={['Dashboard', 'Roles']}>
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
			{#if canCreate}
				<Button onclick={() => openRole(null)}>
					<Icon icon={RiAddLine} />
					New role
				</Button>
			{/if}
		{/snippet}
	</PageHeader>
</div>

<p class="lead">
	Roles say what users may do in your applications. A global role is the same everywhere; an
	application role only means something in its own application.
</p>

<div class="tabs" role="tablist" aria-label="Kinds of role">
	<button
		type="button"
		role="tab"
		aria-selected={data.tab === 'global'}
		class:on={data.tab === 'global'}
		onclick={() => apply({ tab: '', application: '', search: '', default: '' })}
	>
		<Icon icon={RiGlobalLine} />
		Global roles
		<span class="count">{globalCount}</span>
	</button>

	{#if applications.data.length > 0}
		<button
			type="button"
			role="tab"
			aria-selected={data.tab === 'application'}
			class:on={data.tab === 'application'}
			onclick={() => apply({ tab: 'application', application: '', search: '', default: '' })}
		>
			<Icon icon={RiAppsLine} />
			Application roles
			<span class="count">{appCount}</span>
		</button>
	{/if}
</div>

{#if data.tab === 'application' && data.application}
	<p class="narrowed">
		Showing the roles of
		<strong>{applications.data.find((app) => app.id === data.application)?.name}</strong>.
		<button type="button" onclick={() => apply({ application: '' })}>
			Show every application
		</button>
	</p>
{/if}

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
			placeholder="Search roles…"
			bind:value={search}
			oninput={debounced}
			aria-label="Search roles"
		/>
	</form>

	<div class="filter" role="group" aria-label="Filter by default">
		{#each filters as filter (filter.value)}
			<button
				type="button"
				class:selected={data.isDefault === filter.value}
				aria-pressed={data.isDefault === filter.value}
				onclick={() => apply({ default: filter.value })}
			>
				{filter.label}
			</button>
		{/each}
	</div>
</div>

{#if error}
	<div class="gutter error"><Alert>{error}</Alert></div>
{/if}

<RoleTable
	roles={roles.data.roles}
	showApplication={data.tab === 'application'}
	applications={applications.data}
	onOpen={openRole}
	selected={canWrite ? chosen : undefined}
	onSelect={canWrite ? (ids) => (selected = ids) : undefined}
/>

<SelectionBar count={chosen.length} onReset={reset}>
	{#if confirmingDelete}
		{#if holders > 0}
			<span class="warning">{holders} {holders === 1 ? 'user' : 'users'} will lose a role</span>
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
		<Button size="sm" onclick={download} disabled={busy}>
			<Icon icon={RiDownloadLine} />
			JSON
		</Button>
	{/if}
</SelectionBar>

<RoleDrawer
	role={editing}
	initialScope={data.tab === 'application'
		? data.application || (applications.data[0]?.id ?? null)
		: null}
	roles={choices.data}
	applications={applications.data}
	apis={data.apis}
	admin={data.admin}
	bind:open={roleOpen}
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

	/* The two kinds of role are the first choice on the page, so they are
	   tabs across it rather than a field in the toolbar. */
	.tabs {
		display: flex;
		gap: var(--space-1);
		overflow-x: auto;
		scrollbar-width: none;
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
		border-bottom: 1px solid var(--color-border);
	}

	.tabs button {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		height: 42px;
		margin-bottom: -1px;
		padding: 0 var(--space-3);
		border: none;
		border-bottom: 2px solid transparent;
		background: transparent;
		color: var(--color-text-hint);
		font: inherit;
		font-size: var(--text-base);
		font-weight: 600;
		white-space: nowrap;
		cursor: pointer;
	}

	.tabs button:hover {
		color: var(--color-text);
	}

	.tabs button.on {
		border-bottom-color: var(--color-text);
		color: var(--color-text);
	}

	.count {
		display: inline-grid;
		place-items: center;
		min-width: 20px;
		height: 20px;
		padding: 0 6px;
		border-radius: var(--radius-pill);
		background: var(--color-secondary);
		color: var(--color-text-hint);
		font-size: var(--text-xs);
	}

	.narrowed {
		margin: 0 0 var(--space-3);
		padding-inline: var(--page-gutter);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.narrowed strong {
		color: var(--color-text);
	}

	.narrowed button {
		margin-left: var(--space-1);
		padding: 0;
		border: none;
		background: none;
		color: var(--color-text);
		font: inherit;
		text-decoration: underline;
		text-underline-offset: 3px;
		cursor: pointer;
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
