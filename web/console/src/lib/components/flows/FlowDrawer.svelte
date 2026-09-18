<script lang="ts">
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { RiDeleteBinLine } from 'svelte-remixicon';
	import {
		ApiError,
		flowsApi,
		type LoginFlow,
		type LoginFlowInput,
		type LoginStep,
		type LoginStepSpec
	} from '$lib/api';
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
	import StepBuilder from './StepBuilder.svelte';

	type Props = {
		open: boolean;
		/** The flow being changed, or null to write one. */
		flow?: LoginFlow | null;
		/** The steps a flow can be made of. */
		kinds: LoginStepSpec[];
	};

	let { open = $bindable(false), flow = null, kinds }: Props = $props();

	const queryClient = useQueryClient();

	const editing = $derived(flow !== null);

	let name = $state('');
	let slug = $state('');
	let description = $state('');
	let steps = $state<LoginStep[]>(['identifier', 'password']);
	let isDefault = $state(false);
	let enabled = $state(false);
	let allowRegistration = $state(true);
	let allowPasswordReset = $state(true);
	let requireVerifiedEmail = $state(false);
	let lifetimeHours = $state('336');

	let error = $state('');
	let saving = $state(false);
	let confirmingDelete = $state(false);

	/** What the drawer will not let be saved, so the button can say why
	    before the server does. The server checks all of it again. */
	const ready = $derived(name.trim() !== '' && steps.length > 1 && Number(lifetimeHours) > 0);

	// The form is filled in each time the panel opens: with the flow being
	// changed, or with what a new one starts as.
	$effect(() => {
		if (!open) return;

		error = '';
		confirmingDelete = false;

		if (flow) {
			name = flow.name;
			slug = flow.slug;
			description = flow.description;
			steps = [...flow.steps];
			isDefault = flow.is_default;
			enabled = flow.enabled;
			allowRegistration = flow.allow_registration;
			allowPasswordReset = flow.allow_password_reset;
			requireVerifiedEmail = flow.require_verified_email;
			lifetimeHours = String(flow.session_lifetime_hours);
			return;
		}

		name = '';
		slug = '';
		description = '';
		steps = ['identifier', 'password'];
		isDefault = false;
		enabled = false;
		allowRegistration = true;
		allowPasswordReset = true;
		requireVerifiedEmail = false;
		lifetimeHours = '336';
	});

	function input(): LoginFlowInput {
		const body: LoginFlowInput = {
			name: name.trim(),
			description: description.trim(),
			steps,
			is_default: isDefault,
			enabled,
			allow_registration: allowRegistration,
			allow_password_reset: allowPasswordReset,
			require_verified_email: requireVerifiedEmail,
			session_lifetime_hours: Number(lifetimeHours)
		};

		// The identifier is only ever read when the flow is written: an
		// application stores it to name its flow.
		if (!editing) body.slug = slug.trim().toLowerCase();

		return body;
	}

	const save = createMutation(() => ({
		mutationFn: () => (flow ? flowsApi.update(flow.id, input()) : flowsApi.create(input())),
		onSuccess: async () => {
			open = false;
			// A flow can take the default mark from another, and an
			// application's row names the flow it holds.
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.flows.all }),
				queryClient.invalidateQueries({ queryKey: keys.applications.all })
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not save this flow';
		},
		onSettled: () => {
			saving = false;
		}
	}));

	const remove = createMutation(() => ({
		mutationFn: () => flowsApi.remove(flow?.id ?? ''),
		onSuccess: async () => {
			open = false;
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.flows.all }),
				queryClient.invalidateQueries({ queryKey: keys.applications.all })
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not remove this flow';
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (!ready || saving) return;

		error = '';
		saving = true;
		save.mutate();
	}

	/** Marking a flow as the default is also turning it on: it is what every
	    application without one of its own falls back to, so the server
	    refuses a default that is off. */
	function defaultChanged(on: boolean) {
		if (on) enabled = true;
	}

	const lifetimes = [
		{ value: '12', label: '12 hours' },
		{ value: '24', label: 'A day' },
		{ value: '168', label: 'A week' },
		{ value: '336', label: 'Two weeks' },
		{ value: '720', label: 'A month' },
		{ value: '2160', label: 'Three months' }
	];
</script>

<Drawer
	bind:open
	title={editing ? flow!.name : 'New login flow'}
	description={editing
		? 'What this flow asks for, and what it lets people do along the way.'
		: 'A set of steps an application can sign its users in with.'}
	meta={editing ? flow!.slug : undefined}
	width="36rem"
	onsubmit={submit}
>
	{#if error}
		<div class="message"><Alert>{error}</Alert></div>
	{/if}

	<FormSection
		title="Flow"
		description={editing
			? 'The identifier is what an application stores to name this flow, so it stays as it is.'
			: 'The identifier is made from the name unless you give one.'}
	>
		<div class="pair">
			<Input label="Name" bind:value={name} required hint="What it is called in the panel." />
			<Input
				label="Identifier"
				bind:value={slug}
				readOnly={editing}
				placeholder="staff-sign-in"
				hint="Lower case, dashes."
			/>
		</div>

		<Textarea
			label="Description"
			bind:value={description}
			rows={2}
			hint="What this flow is for, for whoever reads the list next."
		/>
	</FormSection>

	<FormSection
		title="Steps"
		description="What somebody is taken through, in order. Every flow starts by asking who is signing in, and has to ask for at least one of those to be proved."
	>
		<StepBuilder bind:steps {kinds} />
	</FormSection>

	<FormSection title="What it lets people do">
		<SwitchField
			label="Let people create an account"
			description="Offer “Create an account” on the sign-in page. Off, an account is made by an administrator or not at all."
			bind:checked={allowRegistration}
		/>

		<SwitchField
			label="Let people reset a password"
			description="Offer “Forgotten your password”. Off, a password is changed by an administrator."
			bind:checked={allowPasswordReset}
		/>

		<SwitchField
			label="Require a confirmed address"
			description="Refuse a sign-in until the email address has been confirmed, rather than letting the account in and asking afterwards."
			bind:checked={requireVerifiedEmail}
		/>

		<Select label="Sessions last" bind:value={lifetimeHours} options={lifetimes} />
	</FormSection>

	<FormSection title="Where it is used">
		<SwitchField
			label="The default flow"
			description="Every application that names no flow of its own signs people in with this one. Making this the default takes the mark from whichever flow has it."
			bind:checked={isDefault}
			onChange={defaultChanged}
		/>

		<SwitchField
			label="Offer this flow"
			description={isDefault
				? 'The default flow is always on: it is what everything else falls back to.'
				: 'An application can be pointed at this flow. Off, the applications holding it fall back to the default.'}
			bind:checked={enabled}
			disabled={isDefault}
		/>

		{#if editing && flow!.applications > 0}
			<p class="note">
				{flow!.applications}
				{flow!.applications === 1 ? 'application signs' : 'applications sign'} people in with this flow.
			</p>
		{/if}
	</FormSection>

	{#snippet footer()}
		{#if editing && !flow!.is_default}
			{#if confirmingDelete}
				<div class="confirm">
					<span>
						{flow!.applications > 0
							? `${flow!.applications} ${flow!.applications === 1 ? 'application falls' : 'applications fall'} back to the default flow.`
							: 'Nothing is using this flow.'}
					</span>
					<Button variant="subtle" size="sm" onclick={() => (confirmingDelete = false)}>
						Keep it
					</Button>
					<Button colorPalette="danger" size="sm" onclick={() => remove.mutate()}>Remove</Button>
				</div>
			{:else}
				<Button
					colorPalette="danger"
					variant="subtle"
					size="sm"
					onclick={() => (confirmingDelete = true)}
				>
					<Icon icon={RiDeleteBinLine} />
					Remove
				</Button>
			{/if}
		{/if}

		<div class="actions">
			<Button variant="subtle" onclick={() => (open = false)}>Cancel</Button>
			<Button type="submit" loading={saving} disabled={!ready || saving}>
				{editing ? 'Save changes' : 'Create flow'}
			</Button>
		</div>
	{/snippet}
</Drawer>

<style>
	.message {
		margin-bottom: var(--space-4);
	}

	.pair {
		display: grid;
		grid-template-columns: 1fr 1fr;
		align-items: start;
		gap: var(--space-3);
	}

	.note {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.confirm {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin-right: auto;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.actions {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		margin-left: auto;
	}

	@media (max-width: 36rem) {
		.pair {
			grid-template-columns: 1fr;
		}
	}
</style>
