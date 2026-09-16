<script lang="ts">
	import { signIn, messageOf } from '$lib/api';
	import { Alert, AuthCard, Button, Checkbox, PasswordField, TextField } from '$lib/components';
	import { legalLinks } from '$lib/utils/legal';
	import { authHref, leaveTo } from '$lib/utils/links';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const app = $derived(data.signInRequest?.application ?? null);

	/** What the account is being made under: the application's agreements
	    where it has them, and the organisation's otherwise. With neither
	    there is nothing to agree to, and no box to tick. */
	const agreements = $derived(legalLinks(app, data.organization));
	const needsConsent = $derived(agreements.length > 0);

	/** The shortest password the server takes. */
	const minLength = 8;

	let firstName = $state('');
	let lastName = $state('');
	// svelte-ignore state_referenced_locally
	let email = $state(data.signInRequest?.login_hint ?? '');
	let password = $state('');
	let confirm = $state('');
	let accepted = $state(false);
	let error = $state('');
	let submitting = $state(false);

	const mismatch = $derived(confirm !== '' && confirm !== password);
	const canSubmit = $derived(
		email.trim() !== '' &&
			password.length >= minLength &&
			confirm === password &&
			(!needsConsent || accepted) &&
			!submitting
	);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!canSubmit || !data.request) return;

		error = '';
		submitting = true;

		try {
			const result = await signIn.register({
				request: data.request,
				email: email.trim(),
				password,
				first_name: firstName.trim(),
				last_name: lastName.trim(),
				accept_terms: accepted
			});

			if (result.redirect_to) {
				leaveTo(result.redirect_to);
				return;
			}

			submitting = false;
		} catch (err) {
			error = messageOf(err);
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Create an account{app ? ` · ${app.name}` : ''}</title>
</svelte:head>

{#if data.expired || !data.request}
	<AuthCard
		organization={data.organization}
		title={data.expired ? 'This sign-in has expired' : 'Start from an application'}
		subtitle="Accounts are created while signing in to an application."
	>
		<Alert tone="info">Go back to the application and choose “Sign in” again.</Alert>
	</AuthCard>
{:else if app && !app.allow_registration}
	<AuthCard organization={data.organization} application={app} title="Registration is closed">
		<Alert tone="info">
			{app.name} does not allow creating accounts here. Ask whoever runs it for access.
		</Alert>

		{#snippet below()}
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a href={authHref('/login', data.request)}>Back to sign in</a>
		{/snippet}
	</AuthCard>
{:else}
	<AuthCard
		organization={data.organization}
		application={app}
		title="Create your account"
		subtitle={app ? `to continue to ${app.name}` : undefined}
	>
		<form onsubmit={submit} novalidate>
			{#if error}<Alert>{error}</Alert>{/if}

			<div class="names">
				<TextField
					label="First name"
					bind:value={firstName}
					autocomplete="given-name"
					disabled={submitting}
				/>
				<TextField
					label="Last name"
					bind:value={lastName}
					autocomplete="family-name"
					disabled={submitting}
				/>
			</div>

			<TextField
				label="Email"
				bind:value={email}
				type="email"
				autocomplete="email"
				autocapitalize="none"
				spellcheck={false}
				disabled={submitting}
			/>

			<PasswordField
				label="Password"
				bind:value={password}
				autocomplete="new-password"
				disabled={submitting}
				hint="At least {minLength} characters"
			/>

			<PasswordField
				label="Confirm password"
				bind:value={confirm}
				autocomplete="new-password"
				disabled={submitting}
				error={mismatch ? 'The passwords do not match' : undefined}
			/>

			{#if needsConsent}
				<Checkbox bind:checked={accepted} disabled={submitting}>
					I agree to the
					{#each agreements as agreement, index (agreement.href)}
						{#if index > 0}and{/if}
						<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
						<a href={agreement.href} target="_blank" rel="noopener noreferrer">
							{agreement.label}
						</a>
					{/each}
				</Checkbox>
			{/if}

			<Button type="submit" block loading={submitting} disabled={!canSubmit}>
				{submitting ? 'Creating account…' : 'Create account'}
			</Button>
		</form>

		{#snippet below()}
			Already have an account?
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a href={authHref('/login', data.request)}>Sign in</a>
		{/snippet}
	</AuthCard>
{/if}

<style>
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}

	.names {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-3);
	}

	@media (max-width: 22rem) {
		.names {
			grid-template-columns: 1fr;
		}
	}
</style>
