<script lang="ts">
	import { signIn, messageOf } from '$lib/api';
	import { Alert, AuthCard, Button, TextField } from '$lib/components';
	import { authHref } from '$lib/utils/links';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const app = $derived(data.signInRequest?.application ?? null);
	// An expired handle is not passed on: the emailed link would lead back to
	// the same dead end.
	const request = $derived(data.signInRequest ? data.request : null);

	// svelte-ignore state_referenced_locally
	let email = $state(data.signInRequest?.login_hint ?? '');
	let error = $state('');
	let submitting = $state(false);
	let sentTo = $state('');

	const canSubmit = $derived(email.trim() !== '' && !submitting);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!canSubmit) return;

		error = '';
		submitting = true;

		try {
			await signIn.forgotPassword({ request: request ?? '', email: email.trim() });
			sentTo = email.trim();
		} catch (err) {
			error = messageOf(err);
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Reset your password{app ? ` · ${app.name}` : ''}</title>
</svelte:head>

{#if sentTo}
	<AuthCard organization={data.organization} application={app} title="Check your email">
		<!-- The same words whether or not the address has an account, so the
		     page cannot be used to find out which ones do. -->
		<Alert tone="success">
			If an account exists for <strong>{sentTo}</strong>, we’ve sent a link to reset its password.
			It works for one hour.
		</Alert>
		<p class="muted">Didn’t get it? Check your spam folder, or try again in a few minutes.</p>

		{#snippet below()}
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a href={authHref('/login', request)}>Back to sign in</a>
		{/snippet}
	</AuthCard>
{:else}
	<AuthCard
		organization={data.organization}
		application={app}
		title="Forgot your password?"
		subtitle="Enter your email and we’ll send you a link to choose a new one."
	>
		<form onsubmit={submit} novalidate>
			{#if error}<Alert>{error}</Alert>{/if}

			<TextField
				label="Email"
				bind:value={email}
				type="email"
				autocomplete="email"
				autocapitalize="none"
				spellcheck={false}
				disabled={submitting}
			/>

			<Button type="submit" block loading={submitting} disabled={!canSubmit}>
				{submitting ? 'Sending…' : 'Send reset link'}
			</Button>
		</form>

		{#snippet below()}
			Remembered it?
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a href={authHref('/login', request)}>Back to sign in</a>
		{/snippet}
	</AuthCard>
{/if}

<style>
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}

	.muted {
		margin-top: var(--space-4);
		color: var(--color-text-hint);
		text-align: center;
	}
</style>
