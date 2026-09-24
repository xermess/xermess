<script lang="ts">
	import { untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { RiArrowDownSLine, RiLinksLine } from 'svelte-remixicon';
	import { ApiError, apisApi, type SigningAlgorithm } from '$lib/api';
	import {
		Alert,
		Button,
		Drawer,
		FormSection,
		Icon,
		Input,
		Select,
		SwitchField,
		Textarea
	} from '$lib/components/ui';
	import { keys } from '$lib/query';
	import ScopeEditor from './ScopeEditor.svelte';
	import {
		algorithms,
		roleBasedAccess,
		scopeInput,
		scopeProblem,
		scopeRow,
		type ScopeRow
	} from './scopes';

	type Props = {
		open: boolean;
	};

	let { open = $bindable(false) }: Props = $props();

	const queryClient = useQueryClient();

	let name = $state('');
	let identifier = $state('');
	let description = $state('');
	let enforceRoles = $state(true);
	let rows = $state<ScopeRow[]>([]);

	let advanced = $state(false);
	let algorithm = $state<SigningAlgorithm>('RS256');
	let lifetimeMinutes = $state('');
	let offlineAccess = $state(false);

	let error = $state('');
	let saving = $state(false);

	const ready = $derived(
		name.trim() !== '' &&
			identifier.trim() !== '' &&
			rows.every((row) => scopeProblem(rows, row) === undefined)
	);

	$effect(() => {
		if (!open) return;

		untrack(() => {
			name = '';
			identifier = '';
			description = '';
			enforceRoles = true;
			rows = [scopeRow()];
			advanced = false;
			algorithm = 'RS256';
			lifetimeMinutes = '';
			offlineAccess = false;
			error = '';
		});
	});

	/** Creating the API opens its page, where its applications, logs and the
	    rest of its settings are. */
	const create = createMutation(() => ({
		mutationFn: () =>
			apisApi.create({
				name: name.trim(),
				identifier: identifier.trim(),
				description: description.trim(),
				enforce_roles: enforceRoles,
				signing_algorithm: algorithm,
				token_lifetime: lifetimeMinutes.trim() === '' ? 0 : Number(lifetimeMinutes) * 60,
				allow_offline_access: offlineAccess,
				scopes: scopeInput(rows)
			}),
		onSuccess: async (result) => {
			await queryClient.invalidateQueries({ queryKey: keys.apis.all });
			open = false;
			await goto(resolve('/admin/(panel)/dashboard/apis/[id]', { id: result.api.id }));
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not create this API';
		},
		onSettled: () => {
			saving = false;
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (saving || !ready) return;

		error = '';
		saving = true;
		create.mutate();
	}
</script>

<Drawer bind:open title="New API" onsubmit={submit}>
	{#if error}
		<div class="error"><Alert>{error}</Alert></div>
	{/if}

	<FormSection
		title="API"
		description="A protected resource applications request access tokens for. Its identifier is the audience of those tokens."
	>
		<Input label="Name" bind:value={name} placeholder="Marketplace API" required />

		<Input
			label="Identifier"
			icon={RiLinksLine}
			bind:value={identifier}
			placeholder="https://api.example.com"
			hint="The aud claim of its access tokens. Unique, and it cannot change later: every service validating tokens compares against it. A URL is usual; it does not have to resolve."
			autocapitalize="none"
			spellcheck={false}
			required
		/>

		<Textarea
			label="Description"
			bind:value={description}
			rows={2}
			placeholder="What the API does, for other administrators"
		/>

		<SwitchField
			label={roleBasedAccess.label}
			description={roleBasedAccess.description}
			bind:checked={enforceRoles}
		/>
	</FormSection>

	<FormSection
		title="Scopes"
		description="The permissions the API exposes. Applications are allowed some of them; roles grant them to users. A default scope is added to every token for the API."
	>
		<ScopeEditor bind:rows editable />
	</FormSection>

	<section class="advanced">
		<button
			type="button"
			class="toggle"
			aria-expanded={advanced}
			onclick={() => (advanced = !advanced)}
		>
			<Icon icon={RiArrowDownSLine} />
			<span>
				<strong>Advanced</strong>
				<small>Signing algorithm, token lifetime, refresh tokens</small>
			</span>
		</button>

		{#if advanced}
			<div class="advanced-body">
				<Select
					label="Token signing algorithm"
					bind:value={algorithm}
					options={algorithms}
					hint="Resource servers validate tokens with the public keys at this server's JWKS URI."
				/>

				<Input
					label="Access token lifetime"
					bind:value={lifetimeMinutes}
					type="number"
					min="1"
					max="1440"
					suffix="minutes"
					placeholder="Application's"
					hint="Leave empty to use each application's access token lifetime."
				/>

				<SwitchField
					label="Allow refresh tokens"
					description="Let applications get a refresh token alongside an access token for this API."
					bind:checked={offlineAccess}
				/>
			</div>
		{/if}
	</section>

	{#snippet footer()}
		<span class="spacer"></span>
		<Button variant="subtle" onclick={() => (open = false)} disabled={saving}>Cancel</Button>
		<Button type="submit" loading={saving} disabled={saving || !ready}>
			{saving ? 'Creating…' : 'Create API'}
		</Button>
	{/snippet}
</Drawer>

<style>
	.error {
		margin-bottom: var(--space-4);
	}

	.advanced {
		margin-top: var(--space-5);
		padding-top: var(--space-5);
		border-top: 1px solid var(--color-border);
	}

	.toggle {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		width: 100%;
		padding: 0;
		border: none;
		background: none;
		color: var(--color-text);
		font: inherit;
		text-align: left;
		cursor: pointer;
	}

	.toggle :global(svg) {
		transition: transform var(--speed-fast);
	}

	.toggle[aria-expanded='false'] :global(svg) {
		transform: rotate(-90deg);
	}

	.toggle span {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.toggle small {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.advanced-body {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		margin-top: var(--space-4);
	}

	.spacer {
		flex: 1;
	}
</style>
