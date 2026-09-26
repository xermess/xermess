<script lang="ts">
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { RiDeleteBinLine, RiLinksLine } from 'svelte-remixicon';
	import { ApiError, apisApi, type API, type SigningAlgorithm } from '$lib/api';
	import {
		Alert,
		Button,
		FieldGrid,
		FormSection,
		Icon,
		Input,
		Note,
		Select,
		SwitchField,
		Textarea
	} from '$lib/components/ui';
	import { keys } from '$lib/query';
	import { algorithms, apiInput, roleBasedAccess } from './scopes';

	type Props = {
		api: API;
		editable: boolean;
	};

	let { api, editable }: Props = $props();

	const queryClient = useQueryClient();

	let name = $state('');
	let description = $state('');
	let enforceRoles = $state(true);
	let algorithm = $state<SigningAlgorithm>('RS256');
	let lifetimeMinutes = $state('');
	let offlineAccess = $state(false);

	let error = $state('');
	let saved = $state(false);
	let saving = $state(false);

	let confirmingDelete = $state(false);
	let deleting = $state(false);

	function fill(from: API) {
		name = from.name;
		description = from.description;
		enforceRoles = from.enforce_roles;
		algorithm = from.signing_algorithm;
		lifetimeMinutes = from.token_lifetime > 0 ? String(Math.round(from.token_lifetime / 60)) : '';
		offlineAccess = from.allow_offline_access;
	}

	// Filled once per API: a refetch in the background leaves what is being
	// typed alone.
	const id = $derived(api.id);
	$effect(() => {
		void id;
		untrack(() => fill(api));
	});

	const lifetime = $derived(lifetimeMinutes.trim() === '' ? 0 : Number(lifetimeMinutes) * 60);

	const dirty = $derived(
		name.trim() !== api.name ||
			description.trim() !== api.description ||
			enforceRoles !== api.enforce_roles ||
			algorithm !== api.signing_algorithm ||
			lifetime !== api.token_lifetime ||
			offlineAccess !== api.allow_offline_access
	);

	const save = createMutation(() => ({
		mutationFn: () =>
			apisApi.update(
				api.id,
				apiInput(api, {
					name: name.trim(),
					description: description.trim(),
					enforce_roles: enforceRoles,
					signing_algorithm: algorithm,
					token_lifetime: lifetime,
					allow_offline_access: offlineAccess
				})
			),
		onSuccess: async (result) => {
			queryClient.setQueryData(keys.apis.one(api.id), result.api);
			fill(result.api);
			saved = true;
			await queryClient.invalidateQueries({ queryKey: keys.apis.all });
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not save this API';
		},
		onSettled: () => {
			saving = false;
		}
	}));

	const remove = createMutation(() => ({
		mutationFn: () => apisApi.remove(api.id),
		onSuccess: async () => {
			await goto(resolve('/admin/(panel)/dashboard/apis'));
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.apis.all }),
				queryClient.invalidateQueries({ queryKey: keys.roles.all }),
				queryClient.invalidateQueries({ queryKey: keys.applications.all })
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not delete this API';
			deleting = false;
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (!editable || saving || !dirty || name.trim() === '') return;

		error = '';
		saved = false;
		saving = true;
		save.mutate();
	}
</script>

<form onsubmit={submit}>
	{#if error}
		<div class="message"><Alert>{error}</Alert></div>
	{:else if saved && !dirty}
		<p class="message saved">Settings saved.</p>
	{/if}

	<fieldset disabled={!editable}>
		<FormSection title="General" description="How the API is named and described in this panel.">
			<Input label="Name" bind:value={name} required />

			<Input
				label="Identifier"
				icon={RiLinksLine}
				value={api.identifier}
				readOnly
				copyable
				hint="The audience of its tokens. It cannot change: every service validating them compares against it. Register a new API to use another."
			/>

			<Textarea label="Description" bind:value={description} rows={2} />
		</FormSection>

		<FormSection title="Access" description="What decides the scopes a user's token receives.">
			<SwitchField
				label={roleBasedAccess.label}
				description={roleBasedAccess.description}
				bind:checked={enforceRoles}
			/>

			<Note>
				Either way, the API must still enforce permissions: it checks the scope an endpoint needs is
				in the token.
			</Note>
		</FormSection>

		<FormSection
			title="Tokens"
			description="How access tokens for this API are signed and how long they last."
		>
			<FieldGrid>
				<Select
					label="Signing algorithm"
					bind:value={algorithm}
					options={algorithms}
					readOnly={!editable}
				/>

				<Input
					label="Access token lifetime"
					bind:value={lifetimeMinutes}
					type="number"
					min="1"
					max="1440"
					suffix="minutes"
					placeholder="Application's"
					hint="Empty uses each application's lifetime."
				/>
			</FieldGrid>

			<SwitchField
				label="Allow refresh tokens"
				description="Let applications with the refresh_token grant get a refresh token for this API by asking for offline_access."
				bind:checked={offlineAccess}
			/>
		</FormSection>
	</fieldset>

	{#if editable}
		<div class="actions">
			<Button variant="subtle" onclick={() => fill(api)} disabled={!dirty || saving}>Discard</Button
			>
			<Button type="submit" loading={saving} disabled={!dirty || saving || name.trim() === ''}>
				{saving ? 'Saving…' : 'Save changes'}
			</Button>
		</div>

		<section class="danger">
			<div>
				<strong>Delete this API</strong>
				<p>
					{api.application_count > 0
						? `${api.application_count} ${api.application_count === 1 ? 'application loses' : 'applications lose'} access, and`
						: 'Its'}
					{api.application_count > 0
						? 'roles lose its scopes.'
						: 'scopes are removed from every role.'}
					This cannot be undone.
				</p>
			</div>

			{#if confirmingDelete}
				<div class="confirm">
					<Button
						variant="subtle"
						size="sm"
						onclick={() => (confirmingDelete = false)}
						disabled={deleting}
					>
						Keep it
					</Button>
					<Button
						colorPalette="danger"
						size="sm"
						loading={deleting}
						disabled={deleting}
						onclick={() => {
							error = '';
							deleting = true;
							remove.mutate();
						}}
					>
						{deleting ? 'Deleting…' : 'Delete API'}
					</Button>
				</div>
			{:else}
				<Button
					colorPalette="danger"
					variant="subtle"
					size="sm"
					onclick={() => (confirmingDelete = true)}
				>
					<Icon icon={RiDeleteBinLine} />
					Delete
				</Button>
			{/if}
		</section>
	{/if}
</form>

<style>
	form {
		max-width: 44rem;
	}

	fieldset {
		min-width: 0;
		margin: 0;
		padding: 0;
		border: none;
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

	.actions {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-2);
		margin-top: var(--space-4);
	}

	.danger {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: var(--space-3);
		margin-top: var(--space-6);
		padding: var(--space-3) var(--space-4);
		border: 1px solid color-mix(in srgb, var(--color-danger) 40%, transparent);
		border-radius: var(--radius-md);
	}

	.danger p {
		margin: 2px 0 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.confirm {
		display: flex;
		gap: var(--space-2);
	}
</style>
