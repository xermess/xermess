<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { signIn, messageOf } from '$lib/api';
	import { Alert, AuthCard, Button, PasswordField, TextField } from '$lib/components';
	import { authHref, leaveTo } from '$lib/utils/links';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const app = $derived(data.signInRequest?.application ?? null);

	// The application may already know who is signing in (login_hint).
	// svelte-ignore state_referenced_locally
	let email = $state(data.signInRequest?.login_hint ?? '');
	let password = $state('');
	let error = $state('');
	let submitting = $state(false);

	const canSubmit = $derived(email.trim() !== '' && password !== '' && !submitting);

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!canSubmit) return;

		error = '';
		submitting = true;

		// Read before anything is awaited: once the session changes, this
		// page's data can be replaced under it.
		const next = data.next;

		try {
			const result = await signIn.login({
				request: data.request ?? '',
				email: email.trim(),
				password
			});

			if (result.password_change_required && result.reset_token) {
				// A temporary password: the user chooses their own before going on.
				let query = `token=${encodeURIComponent(result.reset_token)}&reason=temporary`;
				if (data.request) query += `&request=${encodeURIComponent(data.request)}`;
				// eslint-disable-next-line svelte/no-navigation-without-resolve -- resolved, with a query
				await goto(`${resolve('/reset-password')}?${query}`);
				return;
			}

			if (result.redirect_to) {
				// Back to the application. The button stays busy until it loads.
				leaveTo(result.redirect_to);
				return;
			}

			// Signed in to the account itself. One navigation, which reloads the
			// data on the way: a separate invalidateAll would re-run this page's
			// load, whose redirect for a signed-in user races this goto.
			// eslint-disable-next-line svelte/no-navigation-without-resolve -- a checked path on this site
			await goto(next, { invalidateAll: true, replaceState: true });
		} catch (err) {
			error = messageOf(err);
			password = '';
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>Sign in{app ? ` · ${app.name}` : ''}</title>
</svelte:head>

{#if data.expired}
	<AuthCard
		organization={data.organization}
		title="This sign-in has expired"
		subtitle="Sign-in links only work for a short while."
	>
		<Alert tone="info">Go back to the application and choose “Sign in” again.</Alert>

		{#snippet below()}
			Or <a href={resolve('/login')}>sign in to your account</a>
		{/snippet}
	</AuthCard>
{:else}
	<AuthCard
		organization={data.organization}
		application={app}
		title="Sign in"
		subtitle={app ? `to continue to ${app.name}` : 'to manage your account'}
	>
		<form onsubmit={submit} novalidate>
			{#if error}<Alert>{error}</Alert>{/if}

			<TextField
				label="Email"
				bind:value={email}
				type="email"
				name="email"
				autocomplete="username"
				autocapitalize="none"
				spellcheck={false}
				disabled={submitting}
			/>

			<PasswordField
				label="Password"
				bind:value={password}
				disabled={submitting}
				aside={{ label: 'Forgot password?', href: authHref('/forgot-password', data.request) }}
			/>

			<Button type="submit" block loading={submitting} disabled={!canSubmit}>
				{submitting ? 'Signing in…' : 'Continue'}
			</Button>
		</form>

		{#snippet below()}
			{#if app?.allow_registration && data.request}
				Don’t have an account?
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
				<a href={authHref('/register', data.request)}>Create one</a>
			{/if}
		{/snippet}
	</AuthCard>
{/if}

<style>
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}
</style>
