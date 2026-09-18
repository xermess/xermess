<script lang="ts">
	import { createMutation } from '@tanstack/svelte-query';
	import { RiShieldKeyholeLine } from 'svelte-remixicon';
	import { ApiError, adminsApi, type AdminSecurity } from '$lib/api';
	import { Alert, Button, Panel, SwitchField, Tag } from '$lib/components/ui';

	type Props = {
		/** How administrators are made to sign in, as the page loaded it. */
		security: AdminSecurity;
		/** Whether the administrator reading this has an authenticator, so the
		    panel can warn them what turning it on does to their own session. */
		selfHasMFA: boolean;
	};

	let { security, selfHasMFA }: Props = $props();

	// svelte-ignore state_referenced_locally
	let current = $state(security);

	let error = $state('');
	let saving = $state(false);
	let confirming = $state(false);

	/** How many administrators have nothing to sign in with yet. A change that
	    would shut them out of everything but the setup page is worth saying
	    out loud before it is made. */
	const without = $derived(Math.max(current.administrators - current.with_mfa, 0));

	const save = createMutation(() => ({
		mutationFn: (required: boolean) => adminsApi.updateSecurity({ mfa_required: required }),
		onSuccess: (result) => {
			current = result;
			confirming = false;
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not change this setting';
		},
		onSettled: () => {
			saving = false;
		}
	}));

	function set(required: boolean) {
		// Turning it on is asked about first; turning it off takes nothing
		// away from anybody and is done as asked.
		if (required && !confirming) {
			confirming = true;
			return;
		}

		error = '';
		saving = true;
		save.mutate(required);
	}
</script>

<Panel title="Two-factor sign-in" icon={RiShieldKeyholeLine}>
	{#snippet meta()}
		<Tag tone={current.with_mfa === current.administrators ? 'success' : 'neutral'} small>
			{current.with_mfa} of {current.administrators}
			{current.administrators === 1 ? 'administrator has one' : 'administrators have one'}
		</Tag>
	{/snippet}

	{#if error}
		<div class="message"><Alert>{error}</Alert></div>
	{/if}

	<SwitchField
		label="Require an authenticator"
		description="Every administrator sets one up before they can do anything else, and none of them can turn it off again."
		checked={current.mfa_required}
		disabled={saving}
		onChange={(on) => set(on)}
	/>

	{#if confirming}
		<Alert tone="warning">
			{#if without > 0}
				{without}
				{without === 1 ? 'administrator has' : 'administrators have'} no authenticator yet. They will
				be able to do nothing but set one up — including you, if that is you.
			{:else}
				Every administrator already has one, so nothing changes for them today. New ones will be
				asked to set one up before anything else.
			{/if}

			{#if !selfHasMFA}
				<strong>You have none yet:</strong> the next page you load will ask you to set one up, and you
				will not be able to turn this off again until you have.
			{/if}

			<span class="confirm">
				<Button size="sm" variant="subtle" onclick={() => (confirming = false)} disabled={saving}>
					Cancel
				</Button>
				<Button size="sm" loading={saving} onclick={() => set(true)}>Require it</Button>
			</span>
		</Alert>
	{/if}
</Panel>

<style>
	.message {
		margin-bottom: var(--space-3);
	}

	.confirm {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		margin-top: var(--space-3);
	}
</style>
