<script lang="ts">
	import { resolve } from '$app/paths';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { RiAppsLine } from 'svelte-remixicon';
	import {
		ApiError,
		apisApi,
		applicationsApi,
		type Admin,
		type API,
		type APIApplication
	} from '$lib/api';
	import { Alert, Badge, Checkbox, Icon, SearchInput, Switch } from '$lib/components/ui';
	import { can } from '$lib/permissions';
	import { keys } from '$lib/query';
	import { types } from '$lib/components/applications/applications';

	type Props = {
		api: API;
		admin: Admin;
	};

	let { api, admin }: Props = $props();

	const queryClient = useQueryClient();

	const applications = createQuery(() => ({
		queryKey: keys.apis.applications(api.id),
		queryFn: async () => (await apisApi.applications(api.id)).applications
	}));

	let search = $state('');
	let error = $state('');

	/** The application being saved, so only its row shows the wait. */
	let saving = $state<string | null>(null);

	const shown = $derived.by(() => {
		const all = applications.data ?? [];
		const term = search.trim().toLowerCase();
		const matching = term
			? all.filter((app) => app.name.toLowerCase().includes(term) || app.client_id.includes(term))
			: all;

		// Authorised first: they are what this tab is usually opened to check.
		return [...matching].sort((a, b) => Number(b.authorized) - Number(a.authorized));
	});

	const authorized = $derived((applications.data ?? []).filter((app) => app.authorized).length);

	/** Saved as they are made, through the application's own endpoints, which
	    check the administrator may change that application. */
	const change = createMutation(() => ({
		mutationFn: ({ app, scopes }: { app: APIApplication; scopes: string[] | null }) =>
			scopes === null
				? applicationsApi.revokeAPI(app.id, api.id)
				: applicationsApi.authorizeAPI(app.id, api.id, scopes),
		onSuccess: async (_result, { app }) => {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.apis.applications(api.id) }),
				queryClient.invalidateQueries({ queryKey: keys.apis.one(api.id) }),
				queryClient.invalidateQueries({ queryKey: keys.apis.logs(api.id) }),
				queryClient.invalidateQueries({ queryKey: keys.apis.list('') }),
				queryClient.invalidateQueries({ queryKey: keys.applications.access(app.id) })
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not change this access';
		},
		onSettled: () => {
			saving = null;
		}
	}));

	function save(app: APIApplication, scopes: string[] | null) {
		if (saving) return;

		error = '';
		saving = app.id;
		change.mutate({ app, scopes });
	}

	function toggleScope(app: APIApplication, id: string, on: boolean) {
		save(app, on ? [...app.allowed, id] : app.allowed.filter((it) => it !== id));
	}
</script>

<p class="lead">
	The applications that may request access tokens for this API, and the most each token may carry.
	Authorising here is the same as on the application's API access tab.
</p>

{#if error}
	<div class="message"><Alert>{error}</Alert></div>
{/if}

{#if applications.isPending}
	<p class="muted">Loading applications…</p>
{:else if applications.isError}
	<Alert>Could not load the applications.</Alert>
{:else if applications.data.length === 0}
	<div class="empty">
		<Icon icon={RiAppsLine} size="1.25rem" />
		<div>
			<strong>No applications to authorise</strong>
			<p>
				Register one on the
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
				<a href={resolve('/admin/(panel)/dashboard/applications')}>Applications page</a>, then allow
				it to use this API here.
			</p>
		</div>
	</div>
{:else}
	<div class="toolbar">
		<p class="summary">
			<strong>{authorized}</strong> of {applications.data.length}
			{applications.data.length === 1 ? 'application' : 'applications'} authorised
		</p>

		{#if applications.data.length > 5}
			<div class="search-slot">
				<SearchInput
					label="Filter applications"
					placeholder="Filter applications…"
					bind:value={search}
					size="sm"
				/>
			</div>
		{/if}
	</div>

	<div class="list">
		{#each shown as app (app.id)}
			{@const editable = can(admin, 'applications.write', app.id)}
			{@const kind = types[app.type]}
			<section class:authorized={app.authorized} class:busy={saving === app.id}>
				<header>
					<span class="kind" title={kind.label}><Icon icon={kind.icon} size="1rem" /></span>

					<div class="identity">
						<span class="name">
							<strong>{app.name}</strong>
							{#if !app.enabled}<Badge tone="danger">Disabled</Badge>{/if}
						</span>
						<code>{app.client_id}</code>
					</div>

					<Switch
						label={app.authorized ? 'Authorised' : 'Not authorised'}
						checked={app.authorized}
						disabled={!editable || saving !== null}
						onChange={(on) => save(app, on ? [] : null)}
					/>
				</header>

				{#if app.authorized && api.scopes.length > 0}
					<div class="scopes">
						<span class="count">
							{app.allowed.length} of {api.scopes.length} allowed
						</span>
						<ul>
							{#each api.scopes as scope (scope.id)}
								<li>
									<Checkbox
										checked={app.allowed.includes(scope.id)}
										onChange={(on) => toggleScope(app, scope.id, on)}
										disabled={!editable || saving !== null}
										title={scope.name}
									/>
									<code>{scope.name}</code>
									{#if scope.default}<Badge>default</Badge>{/if}
								</li>
							{/each}
						</ul>
					</div>
				{/if}
			</section>
		{:else}
			<p class="muted">No application matches “{search}”.</p>
		{/each}
	</div>
{/if}

<style>
	.search-slot {
		display: flex;
		width: min(18rem, 100%);
	}

	.lead {
		margin: 0 0 var(--space-4);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		line-height: 1.5;
	}

	.message {
		margin-bottom: var(--space-3);
	}

	.muted {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.toolbar {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-2);
		margin-bottom: var(--space-3);
	}

	.summary {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.summary strong {
		color: var(--color-text);
	}

	.list {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}

	section {
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		transition: opacity var(--speed-fast);
	}

	section.authorized {
		border-color: color-mix(in srgb, var(--color-success), var(--color-border) 55%);
	}

	section.busy {
		opacity: 0.6;
	}

	header {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		padding: var(--space-2) var(--space-3);
	}

	.kind {
		display: grid;
		flex-shrink: 0;
		place-items: center;
		width: 32px;
		height: 32px;
		border-radius: var(--radius-sm);
		background: var(--color-secondary-alt);
		color: var(--color-text-hint);
	}

	.identity {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 1px;
		min-width: 0;
	}

	.name {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	code {
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		overflow-wrap: anywhere;
	}

	.identity code {
		color: var(--color-text-hint);
		font-size: var(--text-xs);
	}

	.scopes {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
		padding: var(--space-2) var(--space-3) var(--space-3);
		border-top: 1px solid var(--color-border);
	}

	.count {
		color: var(--color-text-hint);
		font-size: var(--text-xs);
	}

	ul {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(13rem, 1fr));
		gap: var(--space-2) var(--space-4);
		margin: 0;
		padding: 0;
		list-style: none;
	}

	li {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		min-width: 0;
	}

	.empty {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		padding: var(--space-4);
		border: 1px dashed var(--color-border);
		border-radius: var(--radius-md);
		color: var(--color-text-hint);
	}

	.empty strong {
		color: var(--color-text);
	}

	.empty p {
		margin: 2px 0 0;
		font-size: var(--text-sm);
	}

	.empty a {
		color: var(--color-text);
	}
</style>
