<script lang="ts">
	import { untrack } from 'svelte';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { ApiError, apisApi, type API } from '$lib/api';
	import { Alert, Button } from '$lib/components/ui';
	import { keys } from '$lib/query';
	import ScopeEditor from './ScopeEditor.svelte';
	import { apiInput, scopeInput, scopeProblem, scopeRow, type ScopeRow } from './scopes';

	type Props = {
		api: API;
		editable: boolean;
	};

	let { api, editable }: Props = $props();

	const queryClient = useQueryClient();

	let rows = $state<ScopeRow[]>([]);
	let error = $state('');
	let saved = $state(false);
	let saving = $state(false);

	function fill(from: API) {
		rows = from.scopes.map((scope) => scopeRow(scope));
	}

	const id = $derived(api.id);
	$effect(() => {
		void id;
		untrack(() => fill(api));
	});

	const valid = $derived(rows.every((row) => scopeProblem(rows, row) === undefined));

	/** Compared as the server would store them, so a stray space is no change. */
	const dirty = $derived(
		JSON.stringify(scopeInput(rows)) !==
			JSON.stringify(
				api.scopes.map((scope) => ({
					id: scope.id,
					name: scope.name,
					description: scope.description,
					is_default: scope.is_default
				}))
			)
	);

	/** Stored scopes no longer listed: saving takes them from applications and
	    roles, which is worth saying before it happens. */
	const removed = $derived(
		api.scopes.filter((scope) => !rows.some((row) => row.id === scope.id)).map((it) => it.name)
	);

	const save = createMutation(() => ({
		mutationFn: () => apisApi.update(api.id, apiInput(api, { scopes: scopeInput(rows) })),
		onSuccess: async (result) => {
			queryClient.setQueryData(keys.apis.one(api.id), result.api);
			fill(result.api);
			saved = true;
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.apis.all }),
				queryClient.invalidateQueries({ queryKey: keys.roles.all })
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not save the scopes';
		},
		onSettled: () => {
			saving = false;
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (!editable || saving || !dirty || !valid) return;

		error = '';
		saved = false;
		saving = true;
		save.mutate();
	}
</script>

<form onsubmit={submit}>
	<p class="lead">
		The permissions this API understands. Applications are allowed some of them on the Applications
		tab; roles grant them to users. A <strong>default</strong> scope is added to every token for the API
		without being asked for, as long as the application is allowed it and, with role-based access, the
		user's roles grant it.
	</p>

	{#if error}
		<div class="message"><Alert>{error}</Alert></div>
	{:else if saved && !dirty}
		<p class="message saved">Scopes saved.</p>
	{/if}

	<ScopeEditor bind:rows {editable} />

	{#if editable}
		{#if removed.length > 0}
			<p class="warning">
				Saving removes {removed.join(', ')} from every application allowed and every role granting
				{removed.length === 1 ? 'it' : 'them'}.
			</p>
		{/if}

		<div class="actions">
			<Button variant="subtle" onclick={() => fill(api)} disabled={!dirty || saving}>Discard</Button
			>
			<Button type="submit" loading={saving} disabled={!dirty || !valid || saving}>
				{saving ? 'Saving…' : 'Save scopes'}
			</Button>
		</div>
	{/if}
</form>

<style>
	form {
		max-width: 52rem;
	}

	.lead {
		margin: 0 0 var(--space-4);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		line-height: 1.5;
	}

	.message {
		margin: 0 0 var(--space-4);
	}

	.saved {
		padding: var(--space-2) var(--space-3);
		border-radius: var(--radius-sm);
		background: var(--surface-success);
		font-size: var(--text-sm);
	}

	.warning {
		margin: var(--space-3) 0 0;
		color: var(--color-danger);
		font-size: var(--text-sm);
	}

	.actions {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-2);
		margin-top: var(--space-4);
	}
</style>
