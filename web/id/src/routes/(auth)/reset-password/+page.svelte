<script lang="ts">
	import { resolve } from '$app/paths';
	import { signIn, messageOf } from '$lib/api';
	import { Alert, AuthCard, Button, PasswordField } from '$lib/components';
	import { authHref } from '$lib/utils/links';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const app = $derived(data.signInRequest?.application ?? null);
	const request = $derived(data.signInRequest ? data.request : null);

	const minLength = 8;

	let password = $state('');
	let confirm = $state('');
	let error = $state('');
	let submitting = $state(false);
	let done = $state(false);

	const mismatch = $derived(confirm !== '' && confirm !== password);
	const canSubmit = $derived(password.length >= minLength && confirm === password && !submitting);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!canSubmit) return;

		error = '';
		submitting = true;

		try {
			await signIn.resetPassword({ token: data.token, password });
			done = true;
		} catch (err) {
			error = messageOf(err);
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Choose a new password{app ? ` · ${app.name}` : ''}</title>
</svelte:head>

{#if done}
	<AuthCard organization={data.organization} application={app} title="Password updated">
		<Alert tone="success">
			Your password has been changed, and you’ve been signed out everywhere else.
		</Alert>

		<div class="action">
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a class="primary" href={authHref('/login', request)}>
				Sign in{app ? ` to ${app.name}` : ''}
			</a>
		</div>
	</AuthCard>
{:else if !data.valid}
	<AuthCard organization={data.organization} application={app} title="This link has expired">
		<Alert tone="info">
			Reset links work once, for one hour. Ask for a new one and use the latest email.
		</Alert>

		{#snippet below()}
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a href={authHref('/forgot-password', request)}>Send a new link</a>
			· <a href={resolve('/login')}>Sign in</a>
		{/snippet}
	</AuthCard>
{:else}
	<AuthCard
		organization={data.organization}
		application={app}
		title="Choose a new password"
		subtitle={data.temporary
			? 'Your password was set by an administrator. Choose your own to continue.'
			: undefined}
	>
		<form onsubmit={submit} novalidate>
			{#if error}<Alert>{error}</Alert>{/if}

			<PasswordField
				label="New password"
				bind:value={password}
				autocomplete="new-password"
				disabled={submitting}
				hint="At least {minLength} characters"
			/>

			<PasswordField
				label="Confirm new password"
				bind:value={confirm}
				autocomplete="new-password"
				disabled={submitting}
				error={mismatch ? 'The passwords do not match' : undefined}
			/>

			<Button type="submit" block loading={submitting} disabled={!canSubmit}>
				{submitting ? 'Saving…' : 'Update password'}
			</Button>
		</form>
	</AuthCard>
{/if}

<style>
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}

	.action {
		display: flex;
		margin-top: var(--space-5);
	}

	.primary {
		display: flex;
		flex: 1;
		align-items: center;
		justify-content: center;
		height: var(--control-height);
		border-radius: var(--radius-md);
		background: var(--color-primary);
		color: var(--color-primary-text);
		font-weight: 600;
		text-decoration: none;
	}

	.primary:hover {
		background: var(--color-primary-hover);
		text-decoration: none;
	}
</style>
