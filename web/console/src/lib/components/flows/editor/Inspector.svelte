<script lang="ts">
	import { resolve } from '$app/paths';
	import { RiDeleteBinLine } from 'svelte-remixicon';
	import type { LoginStep, LoginStepSpec } from '$lib/api';
	import {
		Alert,
		Button,
		FormSection,
		Icon,
		Input,
		Select,
		SwitchField,
		Textarea
	} from '$lib/components/ui';
	import {
		MAX_SESSION_HOURS,
		describe,
		labelFor,
		slugFrom,
		specFor,
		type FlowDraft
	} from '../steps';

	/**
	 * The settings of whatever is selected on the canvas. A flow's options sit
	 * on the node they govern: making an account and a verified address on the
	 * step that asks who is signing in, the reset link on the password, the
	 * session's length on the end the flow arrives at, and the flow's own name
	 * on its start.
	 */
	type Props = {
		draft: FlowDraft;
		kinds: LoginStepSpec[];
		selected: string;
		editable: boolean;
		/** A flow not saved yet, whose identifier can still be chosen. */
		creating: boolean;
		/** Whether the flow is the default in the database, which only making
		    another one the default can change. */
		savedDefault: boolean;
		applications: number;
		onRemove: (step: LoginStep) => void;
	};

	let {
		draft = $bindable(),
		kinds,
		selected,
		editable,
		creating,
		savedDefault,
		applications,
		onRemove
	}: Props = $props();

	const step = $derived(draft.steps.find((one) => one === selected));
	const spec = $derived(step ? specFor(step, kinds) : undefined);
	const removable = $derived(
		editable && step !== undefined && step !== draft.steps[0] && !spec?.fixed
	);

	/** A session length in the unit it reads best in, and back. */
	let unit = $state<'hours' | 'days'>(
		draft.session_lifetime_hours % 24 === 0 && draft.session_lifetime_hours >= 24 ? 'days' : 'hours'
	);
	const amount = $derived(
		unit === 'days' ? draft.session_lifetime_hours / 24 : draft.session_lifetime_hours
	);

	function setAmount(value: string) {
		const n = Math.round(Number(value));
		if (Number.isFinite(n)) draft.session_lifetime_hours = unit === 'days' ? n * 24 : n;
	}

	function setUnit(next: 'hours' | 'days') {
		const current = amount;
		unit = next;
		draft.session_lifetime_hours = next === 'days' ? current * 24 : current;
	}

	let slugTouched = $state(false);
</script>

<aside class="inspector" aria-label="Settings">
	{#if selected === 'start'}
		<FormSection
			title="Flow"
			description="What the flow is called, here and wherever it is exported."
		>
			<Input
				label="Name"
				bind:value={
					() => draft.name,
					(value) => {
						draft.name = value;
						if (creating && !slugTouched) draft.slug = slugFrom(value);
					}
				}
				required
				readOnly={!editable}
			/>
			<Input
				label="Identifier"
				bind:value={
					() => draft.slug,
					(value) => {
						draft.slug = value;
						slugTouched = true;
					}
				}
				hint={creating
					? 'Made from the name unless you change it.'
					: 'It names the flow in exports, so it stays as it was made.'}
				readOnly={!editable || !creating}
			/>
			<Textarea label="Description" bind:value={draft.description} rows={3} disabled={!editable} />
		</FormSection>

		<FormSection title="Where it is used">
			<!-- Two switches that sound alike and are not: one says which flow
			     an application follows, the other says whether the people
			     following it get in at all. -->
			<SwitchField
				label="Sign-ins are open"
				description="Off, nobody gets in through this flow — no password, no provider, no code, and no new accounts."
				bind:checked={draft.allow_sign_in}
				disabled={!editable}
			/>
			{#if !draft.allow_sign_in}
				<Alert tone="warning">
					While this is off, everyone following this flow is locked out. An administrator signs in
					to this panel, not through a login flow, so you will not lock yourself out of here.
				</Alert>
			{/if}
			<SwitchField
				label="On"
				description="Off, the applications that name it use the default flow instead."
				bind:checked={draft.enabled}
				disabled={!editable || draft.is_default}
			/>
			<SwitchField
				label="The default flow"
				description={savedDefault
					? 'It is the default. Make another flow the default to change that.'
					: 'Used by every application that names no flow of its own.'}
				bind:checked={draft.is_default}
				disabled={!editable || savedDefault}
			/>
			<p class="note">
				{draft.is_default
					? `The default: used by every application without its own flow, and named by ${applications} more.`
					: `Named by ${applications} applications.`}
			</p>
		</FormSection>
	{:else if selected === 'end'}
		<FormSection
			title="Session"
			description="How long somebody stays signed in after going through this flow."
		>
			<div class="pair">
				<Input
					label="Lasts"
					type="number"
					min="1"
					value={String(amount)}
					oninput={(event) => setAmount((event.currentTarget as HTMLInputElement).value)}
					readOnly={!editable}
				/>
				<Select
					label="Unit"
					value={unit}
					options={[
						{ value: 'hours', label: 'hours' },
						{ value: 'days', label: 'days' }
					]}
					readOnly={!editable}
					onChange={(value) => setUnit(value as 'hours' | 'days')}
				/>
			</div>
			<p class="note">{`At most ${MAX_SESSION_HOURS / 24} days.`}</p>
		</FormSection>
	{:else if step}
		<FormSection title={labelFor(step, kinds)} description={describe(step, kinds)}>
			{#if spec && !spec.implemented}
				<Alert tone="warning"
					>The sign-in pages do not run this step yet. It can be placed as a plan; the flow still
					needs Password or Other accounts to let anybody in.</Alert
				>
			{/if}

			{#if step === 'identifier'}
				<SwitchField
					label="People can create an account"
					description="Offers “Create one” on the sign-in page, where the application allows it too."
					bind:checked={draft.allow_registration}
					disabled={!editable}
				/>
				<SwitchField
					label="Confirm the address of a new account"
					description="Sends a link when somebody signs up. They are signed in either way; the link is waiting for them."
					bind:checked={draft.verify_email_on_register}
					disabled={!editable}
				/>
				<SwitchField
					label="Require a verified address"
					description="An account whose address is unconfirmed is sent a link instead of being signed in."
					bind:checked={draft.require_verified_email}
					disabled={!editable}
				/>
				<SwitchField
					label="People can change their email"
					description="Offers “Change” beside the address on their own account page. The new one is confirmed by a link before it takes effect."
					bind:checked={draft.allow_email_change}
					disabled={!editable}
				/>
				<p class="note">
					Every flow starts by asking who is signing in, so this step cannot move or be removed.
				</p>
			{:else if step === 'password'}
				<SwitchField
					label="Offer “Forgot password”"
					description="A link to reset the password by email. Completing it also confirms the address."
					bind:checked={draft.allow_password_reset}
					disabled={!editable}
				/>
				<SwitchField
					label="Offer “Stay signed in”"
					description="A box beside the password. Left unticked — or not offered — the session ends when the browser closes."
					bind:checked={draft.allow_remember_me}
					disabled={!editable}
				/>
			{:else if step === 'social'}
				<p class="note">
					Shows a button for each provider turned on at
					<a href={resolve('/admin/(panel)/dashboard/social')}>Social</a>
				</p>
			{/if}

			{#if removable}
				<div>
					<Button variant="subtle" colorPalette="danger" size="sm" onclick={() => onRemove(step)}>
						<Icon icon={RiDeleteBinLine} />
						Remove step
					</Button>
				</div>
			{/if}
		</FormSection>
	{/if}
</aside>

<style>
	.inspector {
		overflow-y: auto;
		padding: var(--space-4);
		border-left: 1px solid var(--color-border);
		background: var(--color-surface);
	}

	.pair {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-2);
	}

	.note {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.note a {
		color: var(--color-text);
	}
</style>
