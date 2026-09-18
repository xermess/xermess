<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { RiAddLine, RiRefreshLine, RiSearchLine } from 'svelte-remixicon';
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import type { LoginFlow } from '$lib/api';
	import { Alert, Button, Icon, IconButton } from '$lib/components/ui';
	import FlowDrawer from '$lib/components/flows/FlowDrawer.svelte';
	import FlowTable from '$lib/components/flows/FlowTable.svelte';
	import { can } from '$lib/permissions';
	import { keys, loginFlowsOptions } from '$lib/query';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	// The list is a query seeded with what the server rendered, as on the
	// social page: a save or the refresh button refills the cache rather than
	// reloading the page.
	const flows = createQuery(() =>
		loginFlowsOptions({ flows: data.flows, step_kinds: data.stepKinds })
	);

	const canWrite = $derived(can(data.admin, 'login_flows.write'));

	// A writable derived: typing updates it, and it follows the URL again
	// whenever that changes, so the back button and a shared link both put the
	// right term in the box.
	let search = $derived(data.search);

	/** The rows the search leaves. There are as many flows as somebody has
	    written, so narrowing them here is cheaper than asking again — and it
	    runs while the page is rendered on the server too, so a shared link
	    arrives already filtered. */
	const visible = $derived.by(() => {
		const term = data.search.trim().toLowerCase();
		if (term === '') return flows.data.flows;

		return flows.data.flows.filter((flow) =>
			[flow.name, flow.slug, flow.description].some((value) => value.toLowerCase().includes(term))
		);
	});

	/** The steps named by some flow that the server does not run yet, so the
	    page can say so once rather than every row saying it again. */
	const planned = $derived(new Set(flows.data.flows.flatMap((flow) => flow.planned)));

	let editing = $state<LoginFlow | null>(null);
	let drawerOpen = $state(false);

	async function apply(changes: { search?: string }) {
		const params = new SvelteURLSearchParams(page.url.searchParams);

		for (const [key, value] of Object.entries(changes)) {
			if (value) params.set(key, value);
			else params.delete(key);
		}

		const query = params.toString();
		const path = resolve('/admin/(panel)/dashboard/flows');

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

	function open(flow: LoginFlow | null) {
		editing = flow;
		drawerOpen = true;
	}

	let refreshing = $state(false);

	async function refresh() {
		refreshing = true;

		try {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.flows.all }),
				// A button that does something invisible in 20ms reads as one
				// that did nothing.
				new Promise((done) => setTimeout(done, 400))
			]);
		} finally {
			refreshing = false;
		}
	}
</script>

<svelte:head><title>Login flows · xermess admin</title></svelte:head>

<header>
	<div class="title">
		<h1>Login flows</h1>
		<span class="total">{flows.data.flows.length} total</span>

		<IconButton
			icon={RiRefreshLine}
			label="Refresh the data"
			onclick={refresh}
			loading={refreshing}
			disabled={refreshing}
		/>
	</div>

	{#if canWrite}
		<div class="actions">
			<Button onclick={() => open(null)}>
				<Icon icon={RiAddLine} />
				New flow
			</Button>
		</div>
	{/if}
</header>

<!-- A flow is a record of what a sign-in should be. Most of it is read by the
     sign-in pages already; the steps are not walked yet, and saying so here
     once is better than every row hedging. -->
<div class="gutter note">
	<Alert tone="info">
		The sign-in pages read a flow's options — whether an account can be made, whether a password can
		be reset, and which providers are offered. Walking the steps themselves is still being built:
		{#if planned.size > 0}
			the {planned.size}
			{planned.size === 1 ? 'step' : 'steps'} marked “not run yet”
			{planned.size === 1 ? 'is' : 'are'} a plan.
		{:else}
			a flow that names a step marked “not run yet” is a plan.
		{/if}
	</Alert>
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
			placeholder="Search name, identifier or description…"
			bind:value={search}
			oninput={debounced}
			aria-label="Search login flows"
		/>
	</form>
</div>

<FlowTable
	flows={visible}
	kinds={flows.data.step_kinds}
	onOpen={open}
	empty={flows.data.flows.length === 0 ? 'No flows yet.' : 'No flows match this.'}
/>

<FlowDrawer bind:open={drawerOpen} flow={editing} kinds={flows.data.step_kinds} />

<style>
	header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}

	.title {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	.total {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.actions {
		display: flex;
		gap: var(--space-2);
	}

	.gutter {
		padding-inline: var(--page-gutter);
	}

	.note {
		margin-bottom: var(--space-3);
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
</style>
