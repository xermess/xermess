<script lang="ts">
	import { ApiError, requiredSSO, signIn } from '$lib/api';
	import {
		Alert,
		AuthCard,
		Button,
		Checkbox,
		PasswordField,
		TextField,
		SocialButtons
	} from '$lib/components';
	import { legalLinks } from '$lib/utils/legal';
	import { messageOf, useTranslator } from '$lib/i18n';
	import { authHref, leaveTo, ssoHref } from '$lib/utils/links';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const t = useTranslator();

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
	/** The account was made, and the flow wants its address confirmed before
	    anybody signs in with it: a link has been sent. */
	let sent = $state('');
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
				accept_terms: accepted,
				// Somebody who has just made an account here meant to come
				// back to it, so the session is remembered where the flow
				// offers that at all. There is no box to tick: the sign-in
				// page is where the choice belongs, and a form this long does
				// not need another line.
				remember: data.login.allow_remember_me
			});

			if (result.redirect_to) {
				leaveTo(result.redirect_to);
				return;
			}

			submitting = false;
		} catch (err) {
			// A domain that signs in through its organisation's provider makes
			// its account and keeps its password there: go there instead.
			const sso = requiredSSO(err);
			if (sso) {
				leaveTo(ssoHref(sso.slug, { request: data.request, email: email.trim() }));
				return;
			}

			if (err instanceof ApiError && err.code === 'email_not_verified') {
				sent = messageOf(err, t);
			} else {
				error = messageOf(err, t);
			}
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>{t('register.head')}{app ? ` · ${app.name}` : ''}</title>
</svelte:head>

{#if sent}
	<AuthCard organization={data.organization} application={app} title={t('verify.title')}>
		<Alert tone="info">{sent}</Alert>

		{#snippet below()}
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a href={authHref('/login', data.request)}>{t('action.back_to_sign_in')}</a>
		{/snippet}
	</AuthCard>
{:else if data.expired || !data.request}
	<AuthCard
		organization={data.organization}
		title={data.expired ? t('login.expired_title') : t('register.start_title')}
		subtitle={t('register.start_subtitle')}
	>
		<Alert tone="info">{t('login.expired_body')}</Alert>
	</AuthCard>
{:else if app && !app.allow_registration}
	<AuthCard organization={data.organization} application={app} title={t('register.closed_title')}>
		<Alert tone="info">{t('register.closed_body', { app: app.name })}</Alert>

		{#snippet below()}
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a href={authHref('/login', data.request)}>{t('action.back_to_sign_in')}</a>
		{/snippet}
	</AuthCard>
{:else}
	<AuthCard
		organization={data.organization}
		application={app}
		title={t('register.title')}
		subtitle={app ? t('register.subtitle_app', { app: app.name }) : undefined}
	>
		<form onsubmit={submit} novalidate>
			{#if error}<Alert>{error}</Alert>{/if}

			<div class="names">
				<TextField
					label={t('field.first_name')}
					bind:value={firstName}
					autocomplete="given-name"
					disabled={submitting}
				/>
				<TextField
					label={t('field.last_name')}
					bind:value={lastName}
					autocomplete="family-name"
					disabled={submitting}
				/>
			</div>

			<TextField
				label={t('field.email')}
				bind:value={email}
				type="email"
				autocomplete="email"
				autocapitalize="none"
				spellcheck={false}
				disabled={submitting}
			/>

			<PasswordField
				label={t('field.password')}
				bind:value={password}
				autocomplete="new-password"
				disabled={submitting}
				hint={t('field.password_hint', { count: minLength })}
			/>

			<PasswordField
				label={t('field.confirm_password')}
				bind:value={confirm}
				autocomplete="new-password"
				disabled={submitting}
				error={mismatch ? t('field.password_mismatch') : undefined}
			/>

			{#if needsConsent}
				<Checkbox bind:checked={accepted} disabled={submitting}>
					{t('register.consent')}
					{#each agreements as agreement, index (agreement.href)}
						{#if index > 0}{t('register.consent_and')}{/if}
						<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
						<a href={agreement.href} target="_blank" rel="noopener noreferrer">
							{t(agreement.key)}
						</a>
					{/each}
				</Checkbox>
			{/if}

			<Button type="submit" block loading={submitting} disabled={!canSubmit}>
				{submitting ? t('register.submitting') : t('register.submit')}
			</Button>
		</form>

		<SocialButtons
			providers={data.socialProviders}
			request={data.request}
			label={t('register.social')}
			disabled={submitting}
		/>

		{#snippet below()}
			{t('register.have_account')}
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a href={authHref('/login', data.request)}>{t('action.sign_in')}</a>
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
