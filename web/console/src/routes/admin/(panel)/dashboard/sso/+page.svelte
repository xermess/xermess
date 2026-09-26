<script lang="ts">
	import { RiAddLine, RiBuilding2Line, RiRefreshLine } from 'svelte-remixicon';
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import type { SSOConnection } from '$lib/api';
	import { BRAND } from '$lib/brand';
	import { Button, Icon, IconButton, PageHeader, SearchInput } from '$lib/components/ui';
	import SSODrawer from '$lib/components/sso/SSODrawer.svelte';
	import SSOTable from '$lib/components/sso/SSOTable.svelte';
	import { can } from '$lib/permissions';
	import { keys, ssoOptions } from '$lib/query';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	const list = createQuery(() => ssoOptions({ connections: data.connections }));

	const canWrite = $derived(can(data.admin, 'sso.write'));

	let search = $state('');

	/** There are as many connections as organisations — a handful — so the
	    search narrows them here, by name, identifier or domain. */
	const visible = $derived.by(() => {
		const term = search.trim().toLowerCase();
		if (term === '') return list.data.connections;

		return list.data.connections.filter((connection) =>
			[connection.name, connection.slug, ...connection.domains].some((value) =>
				value.toLowerCase().includes(term)
			)
		);
	});

	let editing = $state<SSOConnection | null>(null);
	let drawerOpen = $state(false);

	function open(connection: SSOConnection | null) {
		editing = connection;
		drawerOpen = true;
	}

	let refreshing = $state(false);

	async function refresh() {
		refreshing = true;

		try {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.sso.all }),
				new Promise((done) => setTimeout(done, 400))
			]);
		} finally {
			refreshing = false;
		}
	}
</script>

<svelte:head><title>SSO integrations · {BRAND.name}</title></svelte:head>

<div class="heading">
	<PageHeader crumbs={['Dashboard', 'SSO integrations']}>
		{#snippet secondary()}
			<span class="total">{`${list.data.connections.length} total`}</span>

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
					New connection
				</Button>
			{/if}
		{/snippet}
	</PageHeader>
</div>

<p class="lead">
	Let an organisation sign its people in through its own identity provider — Okta, Microsoft Entra
	ID, Google Workspace, ADFS — over OpenID Connect or SAML 2.0. A connection owns email domains: the
	people at them sign in through it, can be made to, and get accounts and roles from what the
	provider says.
</p>

{#if list.data.connections.length === 0}
	<section class="empty">
		<span class="mark" aria-hidden="true"><Icon icon={RiBuilding2Line} size="1.6rem" /></span>
		<div>
			<h2>No identity providers connected yet</h2>
			<p>
				Connect a company's identity provider and its people sign in with their work account: routed
				there by their email domain, given an account the first time, and roles from the groups the
				provider puts them in.
			</p>
			{#if canWrite}
				<Button onclick={() => open(null)}>
					<Icon icon={RiAddLine} />
					New connection
				</Button>
			{/if}
		</div>
	</section>
{:else}
	<div class="toolbar">
		<SearchInput
			label="Search"
			placeholder="Search name, identifier or domain…"
			bind:value={search}
		/>
	</div>

	<SSOTable connections={visible} onOpen={open} empty="No connections match this." />
{/if}

<SSODrawer
	bind:open={drawerOpen}
	connection={editing}
	roles={data.roles}
	canReadRoles={data.mayReadRoles}
	{canWrite}
/>

<style>
	.heading {
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}

	.lead {
		max-width: 90ch;
		margin: 0 0 var(--space-3);
		padding-inline: var(--page-gutter);
		color: var(--color-text-hint);
		line-height: 1.5;
	}

	.toolbar {
		display: flex;
		gap: var(--space-2);
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}

	.empty {
		display: flex;
		align-items: flex-start;
		gap: var(--space-4);
		max-width: 60rem;
		margin-inline: var(--page-gutter);
		padding: var(--space-5) var(--space-4);
		border: 1px dashed var(--color-border);
		border-radius: var(--radius-lg);
	}

	.empty h2 {
		margin: 0;
		font-size: var(--text-lg);
	}

	.empty p {
		max-width: 40rem;
		margin: var(--space-1) 0 var(--space-3);
		color: var(--color-text-hint);
		line-height: 1.55;
	}

	.mark {
		display: grid;
		flex-shrink: 0;
		place-items: center;
		width: 56px;
		height: 56px;
		border-radius: 16px;
		background: var(--color-primary);
		color: var(--color-primary-text);
	}
</style>
