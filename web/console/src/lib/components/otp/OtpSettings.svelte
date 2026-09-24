<script lang="ts">
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { resolve } from '$app/paths';
	import { RiGitBranchLine, RiHashtag, RiShieldKeyholeLine, RiTimerLine } from 'svelte-remixicon';
	import { ApiError, otpApi, type OTPResponse, type OTPSettingsInput } from '$lib/api';
	import { Alert, Button, Input, List, ListItem, Panel, Tag, Thumb } from '$lib/components/ui';
	import { keys } from '$lib/query';

	type Props = {
		settings: OTPResponse;
	};

	let { settings }: Props = $props();

	const queryClient = useQueryClient();

	const stored = $derived(settings.otp);
	const limits = $derived(settings.limits);

	/** The stored record as the form holds it: every field is a number, and a
	    field binds text. */
	function formOf(otp: OTPResponse['otp']) {
		return {
			length: String(otp.length),
			lifetime_minutes: String(otp.lifetime_minutes),
			max_attempts: String(otp.max_attempts),
			resend_seconds: String(otp.resend_seconds)
		};
	}

	// Seeded once, when the panel is first drawn: a refetch in the background
	// leaves what is being typed alone, and a save fills it in again.
	// svelte-ignore state_referenced_locally
	let form = $state(formOf(settings.otp));

	let error = $state('');
	let saved = $state(false);
	let saving = $state(false);

	function discard() {
		error = '';
		saved = false;
		form = formOf(stored);
	}

	const input = $derived<Required<OTPSettingsInput>>({
		length: Number(form.length),
		lifetime_minutes: Number(form.lifetime_minutes),
		max_attempts: Number(form.max_attempts),
		resend_seconds: Number(form.resend_seconds)
	});

	const dirty = $derived(
		(Object.keys(input) as (keyof typeof input)[]).some((key) => input[key] !== stored[key])
	);

	/** The same bounds the server holds these to, so a value it would refuse
	    is refused here first — with the button, rather than with an error. */
	function within(value: number, min: number, max: number): boolean {
		return Number.isInteger(value) && value >= min && value <= max;
	}

	const complete = $derived(
		within(input.length, limits.min_length, limits.max_length) &&
			within(input.lifetime_minutes, 1, limits.max_lifetime_minutes) &&
			within(input.max_attempts, 1, limits.max_attempts) &&
			within(input.resend_seconds, 0, limits.max_resend_seconds)
	);

	/** How many codes there are to guess at, against how many guesses one
	    sign-in gets. It is the number these settings actually decide, so the
	    page says it rather than leaving it to be worked out. */
	const odds = $derived.by(() => {
		if (!complete) return '';

		const codes = 10 ** input.length;
		const chance = Math.round(codes / input.max_attempts);

		return `A code is 1 of ${codes.toLocaleString()}, and one sign-in gets ${input.max_attempts} ${
			input.max_attempts === 1 ? 'try' : 'tries'
		} — about 1 in ${chance.toLocaleString()} of being guessed.`;
	});

	const save = createMutation(() => ({
		mutationFn: () => otpApi.update(input),
		onSuccess: async (result) => {
			queryClient.setQueryData(keys.otp.settings, result);
			form = formOf(result.otp);
			saved = true;
			// The log gains an entry for the change.
			await queryClient.invalidateQueries({ queryKey: keys.admin.overview });
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not save these settings';
		},
		onSettled: () => {
			saving = false;
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (saving || !dirty || !complete) return;

		error = '';
		saved = false;
		saving = true;
		save.mutate();
	}
</script>

<form onsubmit={submit}>
	{#if error}
		<Alert>{error}</Alert>
	{:else if saved && !dirty}
		<Alert tone="success">Settings saved. The next code sent follows them.</Alert>
	{/if}

	{#if settings.flows.length === 0}
		<!-- Settings nothing reads are a note rather than a setting, so the
		     page says so instead of looking as though it is doing something. -->
		<Alert tone="warning">
			No login flow asks for an emailed code yet, so nothing reads these. Add the
			<strong>Emailed code</strong> step to a flow on the
			<a href={resolve('/admin/(panel)/dashboard/flows')}>Login flows</a> page.
		</Alert>
	{/if}

	<Panel title="The code" icon={RiHashtag}>
		<div class="grid">
			<Input
				label="Digits"
				bind:value={form.length}
				type="number"
				min={limits.min_length}
				max={limits.max_length}
				hint="Between {limits.min_length} and {limits.max_length}. Six is what most people expect."
			/>

			<Input
				label="Guesses"
				bind:value={form.max_attempts}
				type="number"
				min={1}
				max={limits.max_attempts}
				hint="Wrong codes before the sign-in is over and has to be started again."
			/>
		</div>

		{#if odds}
			<p class="note">{odds}</p>
		{/if}
	</Panel>

	<Panel title="Timing" icon={RiTimerLine}>
		<div class="grid">
			<Input
				label="Valid for"
				bind:value={form.lifetime_minutes}
				type="number"
				min={1}
				max={limits.max_lifetime_minutes}
				suffix="minutes"
				hint="A code that outlives the message it came in is a password sitting in an inbox."
			/>

			<Input
				label="Wait before another"
				bind:value={form.resend_seconds}
				type="number"
				min={0}
				max={limits.max_resend_seconds}
				suffix="seconds"
				hint="What stops “send it again” being a way to post mail to a stranger. Zero turns the wait off."
			/>
		</div>
	</Panel>

	{#if settings.flows.length > 0}
		<Panel title="Flows that ask for a code" icon={RiGitBranchLine} flush>
			{#snippet meta()}
				<Tag small>{settings.flows.length}</Tag>
			{/snippet}

			<List>
				{#each settings.flows as flow (flow.id)}
					<ListItem title={flow.name} description={flow.slug}>
						{#snippet lead()}
							<Thumb icon={RiShieldKeyholeLine} />
						{/snippet}

						{#snippet end()}
							{#if flow.is_default}
								<Tag tone="info" small>Default</Tag>
							{/if}
							{#if flow.enabled}
								<Tag tone="success" dot small>On</Tag>
							{:else}
								<Tag small>Off</Tag>
							{/if}
						{/snippet}
					</ListItem>
				{/each}
			</List>
		</Panel>
	{/if}

	{#if dirty}
		<div class="actions">
			<span class="pending">Unsaved changes</span>

			<Button variant="subtle" onclick={discard} disabled={saving}>Discard</Button>
			<Button type="submit" loading={saving} disabled={saving || !complete}>
				{saving ? 'Saving…' : 'Save changes'}
			</Button>
		</div>
	{/if}
</form>

<style>
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		align-items: start;
		gap: var(--space-3) var(--space-4);
	}

	.note {
		margin: var(--space-3) 0 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		line-height: 1.5;
	}

	.actions {
		position: sticky;
		bottom: 0;
		z-index: 1;
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: var(--space-2);
		margin-inline: calc(var(--space-4) * -1);
		padding: var(--space-3) var(--space-4);
		border-top: 1px solid var(--color-secondary-alt);
		background: var(--color-surface);
		box-shadow: var(--shadow-panel);
	}

	.pending {
		margin-right: auto;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	@media (max-width: 40rem) {
		.grid {
			grid-template-columns: 1fr;
		}
	}
</style>
