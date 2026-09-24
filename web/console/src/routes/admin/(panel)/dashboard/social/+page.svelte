<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import {
		RiAddLine,
		RiDeleteBinLine,
		RiEyeLine,
		RiEyeOffLine,
		RiRefreshLine
	} from 'svelte-remixicon';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { ApiError, socialApi, type SocialProvider } from '$lib/api';
	import {
		Alert,
		Button,
		Icon,
		IconButton,
		PageHeader,
		SearchInput,
		SelectionBar
	} from '$lib/components/ui';
	import SocialDrawer from '$lib/components/social/SocialDrawer.svelte';
	import SocialTable from '$lib/components/social/SocialTable.svelte';
	import { can } from '$lib/permissions';
	import { keys, socialProvidersOptions } from '$lib/query';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	// The list is a query seeded with what the server rendered, as on the
	// users page: a save or the refresh button refills the cache rather than
	// reloading the page.
	const social = createQuery(() =>
		socialProvidersOptions({ providers: data.providers, kinds: data.kinds })
	);

	const canWrite = $derived(can(data.admin, 'social.write'));

	// A writable derived: typing updates it, and it follows the URL again
	// whenever that changes, so the back button and a shared link both put the
	// right term in the box.
	let search = $derived(data.search);

	/** The rows the search and the filter leave.

	    Both live in the URL, as everywhere else in the panel, but the filtering
	    itself is done here: there are as many providers as an installation has
	    registered — a handful — and the endpoint answers with all of them, so
	    asking the server again would be a round trip to do less than this line
	    does. It runs while the page is rendered on the server too, so a shared
	    link arrives already filtered. */
	const providers = $derived.by(() => {
		const term = data.search.trim().toLowerCase();

		return social.data.providers.filter((provider) => {
			if (data.status === 'enabled' && !provider.enabled) return false;
			if (data.status === 'off' && provider.enabled) return false;

			if (term === '') return true;

			return [provider.name, provider.slug, provider.kind, provider.client_id].some((value) =>
				value.toLowerCase().includes(term)
			);
		});
	});

	let editing = $state<SocialProvider | null>(null);
	let drawerOpen = $state(false);

	/** The rows that are ticked, by id, and the ones of those still on screen:
	    a search can take a ticked row out of view, and acting on what nobody
	    can see is not something a panel should offer. */
	let selected = $state<string[]>([]);
	const visible = $derived(new Set(providers.map((provider) => provider.id)));
	const chosen = $derived(selected.filter((id) => visible.has(id)));

	let confirmingRemove = $state(false);
	let error = $state('');
	let busy = $state(false);

	/** How many accounts the ticked providers sign in, so removing them says
	    what it costs before it is done. */
	const chosenAccounts = $derived(
		social.data.providers
			.filter((provider) => chosen.includes(provider.id))
			.reduce((sum, provider) => sum + provider.identities, 0)
	);

	/** The search and the filter are the URL, so the back button walks through
	    them and a filtered list can be linked to. */
	async function apply(changes: { search?: string; status?: string }) {
		const params = new SvelteURLSearchParams(page.url.searchParams);

		for (const [key, value] of Object.entries(changes)) {
			if (value) params.set(key, value);
			else params.delete(key);
		}

		const query = params.toString();
		const path = resolve('/admin/(panel)/dashboard/social');

		// resolve() has already applied any base path; the query is only ever
		// appended to what it returned.
		// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto(query ? `${path}?${query}` : path, { keepFocus: true, noScroll: true });
	}

	let timer: ReturnType<typeof setTimeout>;

	function debounced() {
		clearTimeout(timer);
		timer = setTimeout(() => apply({ search }), 250);
	}

	function open(provider: SocialProvider | null) {
		editing = provider;
		drawerOpen = true;
	}

	let refreshing = $state(false);

	async function refresh() {
		refreshing = true;

		try {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.social.all }),
				// A button that does something invisible in 20ms reads as one
				// that did nothing.
				new Promise((done) => setTimeout(done, 400))
			]);
		} finally {
			refreshing = false;
		}
	}

	function reset() {
		selected = [];
		confirmingRemove = false;
		error = '';
	}

	/** Offering or withdrawing the ticked providers: one call each, because
	    that is what the API takes, and the cache is refilled afterwards so a
	    half-finished change still leaves the list right.
	
	    The call says only what it changes — the endpoint is a PATCH, and
	    everything it does not mention is left as it is. */
	const setEnabled = createMutation(() => ({
		mutationFn: async (enabled: boolean) => {
			for (const id of chosen) {
				const provider = social.data.providers.find((one) => one.id === id);
				if (!provider || provider.enabled === enabled) continue;

				await socialApi.update(id, { enabled });
			}
		},
		onSuccess: () => {
			reset();
			return queryClient.invalidateQueries({ queryKey: keys.social.all });
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not change these providers';
		},
		onSettled: () => {
			busy = false;
		}
	}));

	const removeSelected = createMutation(() => ({
		mutationFn: async (ids: string[]) => {
			for (const id of ids) {
				await socialApi.remove(id);
			}
		},
		onSuccess: () => {
			reset();
			// Users lose a way in, so their records change with them.
			return Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.social.all }),
				queryClient.invalidateQueries({ queryKey: keys.users.all })
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not remove these providers';
		},
		onSettled: () => {
			busy = false;
		}
	}));

	function run(what: 'enable' | 'disable' | 'remove') {
		if (busy) return;

		error = '';
		busy = true;

		if (what === 'remove') {
			removeSelected.mutate(chosen);
			return;
		}

		setEnabled.mutate(what === 'enable');
	}

	const filters = [
		{ value: '', label: 'All' },
		{ value: 'enabled', label: 'Offered' },
		{ value: 'off', label: 'Off' }
	];
</script>

<svelte:head><title>Social · xermess admin</title></svelte:head>

<div class="heading">
	<PageHeader crumbs={['Dashboard', 'Social']}>
		{#snippet secondary()}
			<span class="total">{social.data.providers.length} total</span>

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
				<Button onclick={() => open(null)}>
					<Icon icon={RiAddLine} />
					Add provider
				</Button>
			{/if}
		{/snippet}
	</PageHeader>
</div>

<div class="toolbar">
	<SearchInput
		label="Search providers"
		placeholder="Search name, identifier or client id…"
		bind:value={search}
		onsubmit={() => apply({ search })}
		oninput={debounced}
	/>

	<div class="filter" role="group" aria-label="Filter by whether it is offered">
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

<SocialTable
	{providers}
	kinds={social.data.kinds}
	onOpen={open}
	empty={social.data.providers.length === 0
		? 'No providers yet. Add one to offer a button beside the password form.'
		: 'No providers match this.'}
	selected={canWrite ? chosen : undefined}
	onSelect={canWrite ? (ids) => (selected = ids) : undefined}
/>

<SelectionBar count={chosen.length} onReset={reset}>
	{#if confirmingRemove}
		<span class="warning">
			{chosenAccounts > 0
				? `${chosenAccounts} ${chosenAccounts === 1 ? 'person signs' : 'people sign'} in with these. Their accounts stay; this way in goes.`
				: 'Nobody signs in with these yet.'}
		</span>
		<Button variant="subtle" size="sm" onclick={() => (confirmingRemove = false)} disabled={busy}>
			Keep them
		</Button>
		<Button colorPalette="danger" size="sm" onclick={() => run('remove')} disabled={busy}>
			{busy ? 'Removing…' : `Remove ${chosen.length}`}
		</Button>
	{:else}
		<Button size="sm" variant="subtle" onclick={() => run('enable')} disabled={busy}>
			<Icon icon={RiEyeLine} />
			Offer
		</Button>
		<Button size="sm" variant="subtle" onclick={() => run('disable')} disabled={busy}>
			<Icon icon={RiEyeOffLine} />
			Turn off
		</Button>
		<Button
			colorPalette="danger"
			size="sm"
			onclick={() => (confirmingRemove = true)}
			disabled={busy}
		>
			<Icon icon={RiDeleteBinLine} />
			Remove
		</Button>
	{/if}
</SelectionBar>

<SocialDrawer bind:open={drawerOpen} provider={editing} kinds={social.data.kinds} />

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

	.warning {
		margin-right: var(--space-2);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	@media (max-width: 40rem) {
		.toolbar {
			flex-direction: column;
			align-items: stretch;
		}
	}
</style>
