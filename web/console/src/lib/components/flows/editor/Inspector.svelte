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
	import { useTranslator } from '$lib/i18n';
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

	const t = useTranslator();

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

<aside class="inspector" aria-label={t('flows.inspector')}>
	{#if selected === 'start'}
		<FormSection title={t('flows.section_flow')} description={t('flows.section_flow_hint')}>
			<Input
				label={t('flows.name')}
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
				label={t('flows.slug')}
				bind:value={
					() => draft.slug,
					(value) => {
						draft.slug = value;
						slugTouched = true;
					}
				}
				hint={creating ? t('flows.slug_hint') : t('flows.slug_fixed')}
				readOnly={!editable || !creating}
			/>
			<Textarea
				label={t('flows.description')}
				bind:value={draft.description}
				rows={3}
				disabled={!editable}
			/>
		</FormSection>

		<FormSection title={t('flows.section_use')}>
			<SwitchField
				label={t('flows.enabled')}
				description={t('flows.enabled_hint')}
				bind:checked={draft.enabled}
				disabled={!editable || draft.is_default}
			/>
			<SwitchField
				label={t('flows.default')}
				description={savedDefault ? t('flows.default_is') : t('flows.default_hint')}
				bind:checked={draft.is_default}
				disabled={!editable || savedDefault}
			/>
			<p class="note">
				{draft.is_default
					? t('flows.used_default', { count: applications })
					: t('flows.used_by', { count: applications })}
			</p>
		</FormSection>
	{:else if selected === 'end'}
		<FormSection title={t('flows.section_session')} description={t('flows.section_session_hint')}>
			<div class="pair">
				<Input
					label={t('flows.session_length')}
					type="number"
					min="1"
					value={String(amount)}
					oninput={(event) => setAmount((event.currentTarget as HTMLInputElement).value)}
					readOnly={!editable}
				/>
				<Select
					label={t('flows.session_unit')}
					value={unit}
					options={[
						{ value: 'hours', label: t('flows.unit_hours') },
						{ value: 'days', label: t('flows.unit_days') }
					]}
					readOnly={!editable}
					onChange={(value) => setUnit(value as 'hours' | 'days')}
				/>
			</div>
			<p class="note">{t('flows.session_max', { days: MAX_SESSION_HOURS / 24 })}</p>
		</FormSection>
	{:else if step}
		<FormSection title={labelFor(step, kinds, t)} description={describe(step, kinds, t)}>
			{#if spec && !spec.implemented}
				<Alert tone="warning">{t('flows.planned_body')}</Alert>
			{/if}

			{#if step === 'identifier'}
				<SwitchField
					label={t('flows.allow_registration')}
					description={t('flows.allow_registration_hint')}
					bind:checked={draft.allow_registration}
					disabled={!editable}
				/>
				<SwitchField
					label={t('flows.require_verified')}
					description={t('flows.require_verified_hint')}
					bind:checked={draft.require_verified_email}
					disabled={!editable}
				/>
				<p class="note">{t('flows.identifier_fixed')}</p>
			{:else if step === 'password'}
				<SwitchField
					label={t('flows.allow_reset')}
					description={t('flows.allow_reset_hint')}
					bind:checked={draft.allow_password_reset}
					disabled={!editable}
				/>
			{:else if step === 'social'}
				<p class="note">
					{t('flows.social_hint')}
					<a href={resolve('/admin/(panel)/dashboard/social')}>{t('flows.social_link')}</a>
				</p>
			{/if}

			{#if removable}
				<div>
					<Button variant="subtle" colorPalette="danger" size="sm" onclick={() => onRemove(step)}>
						<Icon icon={RiDeleteBinLine} />
						{t('flows.remove_step')}
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
