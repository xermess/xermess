<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { signIn, messageOf } from '$lib/api';
	import {
		Alert,
		AuthCard,
		Button,
		PasswordField,
		TextField,
		SocialButtons
	} from '$lib/components';
	import { useTranslator } from '$lib/i18n';
	import { authHref, leaveTo } from '$lib/utils/links';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const t = useTranslator();

	const app = $derived(data.signInRequest?.application ?? null);

	/** What the login flow behind this sign-in lets the page offer. The
	    application's own "allow registration" still has to agree: a flow says
	    what this installation allows, an application what it wants. */
	const flow = $derived(data.login);
	const offerSocial = $derived(flow.steps.includes('social'));
	const offerRegistration = $derived(flow.allow_registration && app?.allow_registration === true);

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
	<title>{t('login.head')}{app ? ` · ${app.name}` : ''}</title>
</svelte:head>

{#if data.expired}
	<AuthCard
		organization={data.organization}
		title={t('login.expired_title')}
		subtitle={t('login.expired_subtitle')}
	>
		<Alert tone="info">{t('login.expired_body')}</Alert>

		{#snippet below()}
			{t('login.expired_or')}
			<a href={resolve('/login')}>{t('login.expired_link')}</a>
		{/snippet}
	</AuthCard>
{:else}
	<AuthCard
		organization={data.organization}
		application={app}
		title={t('login.title')}
		subtitle={app ? t('login.subtitle_app', { app: app.name }) : t('login.subtitle_account')}
	>
		<form onsubmit={submit} novalidate>
			{#if error}<Alert>{error}</Alert>{/if}

			<TextField
				label={t('field.email')}
				bind:value={email}
				type="email"
				name="email"
				autocomplete="username"
				autocapitalize="none"
				spellcheck={false}
				disabled={submitting}
			/>

			<PasswordField
				label={t('field.password')}
				bind:value={password}
				disabled={submitting}
				aside={flow.allow_password_reset
					? { label: t('login.forgot'), href: authHref('/forgot-password', data.request) }
					: undefined}
			/>

			<Button type="submit" block loading={submitting} disabled={!canSubmit}>
				{submitting ? t('login.submitting') : t('login.submit')}
			</Button>
		</form>

		<SocialButtons
			providers={offerSocial ? data.socialProviders : []}
			request={data.request}
			next={data.next ?? null}
			label={t('login.social')}
			disabled={submitting}
		/>

		{#snippet below()}
			{#if offerRegistration && data.request}
				{t('login.no_account')}
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
				<a href={authHref('/register', data.request)}>{t('login.create_one')}</a>
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
