<script lang="ts">
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { ApiError, requiredSSO, signIn, type CodeChallenge, type SignedIn } from '$lib/api';
	import {
		Alert,
		AuthCard,
		Button,
		Checkbox,
		CodeForm,
		PasswordField,
		TextField,
		SocialButtons,
		SSOButtons
	} from '$lib/components';
	import { messageOf, useTranslator } from '$lib/i18n';
	import { authHref, leaveTo, ssoHref } from '$lib/utils/links';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const t = useTranslator();

	const app = $derived(data.signInRequest?.application ?? null);

	/** What the login flow behind this sign-in lets the page offer. The
	    application's own "allow registration" still has to agree: a flow says
	    what this installation allows, an application what it wants. */
	const flow = $derived(data.login);
	const offerSocial = $derived(flow.steps.includes('social'));
	/** A flow without a password step signs people in through the buttons
	    alone, and makes accounts that way too. */
	const offerPassword = $derived(flow.steps.includes('password'));
	const offerRegistration = $derived(
		offerPassword && flow.allow_registration && app?.allow_registration === true
	);
	/** Sign-ins are closed for this flow: nothing on this page would work, so
	     nothing is offered. The server refuses every way in regardless. */
	const closed = $derived(!flow.allow_sign_in);
	const offerRemember = $derived(flow.allow_remember_me);

	// The application may already know who is signing in (login_hint).
	// svelte-ignore state_referenced_locally
	let email = $state(data.signInRequest?.login_hint ?? '');
	let password = $state('');
	/** "Stay signed in": off by default, because the machine somebody is
	    signing in on is not known to be theirs. */
	let remember = $state(false);
	let error = $state('');
	/** Said instead of an error when the address has to be confirmed first:
	    a link is on its way, which is news rather than a failure. */
	let notice = $state('');
	let submitting = $state(false);
	/** The identity provider the browser is being sent to, when the address
	    typed belongs to a domain that has to sign in through it. */
	let redirecting = $state('');
	/** A sign-in the flow held back for a code emailed to the address: the
	    page swaps the password form for the code form until it is typed. */
	let challenge = $state<CodeChallenge | null>(null);

	const canSubmit = $derived(email.trim() !== '' && password !== '' && !submitting);

	/** What to do with a sign-in that the server has answered: the same three
	    endings whether it was the password that finished it or the code, so
	    both forms hand what they got to this.

	    `next` is read by the caller before anything is awaited: once the
	    session changes, this page's data can be replaced under it. */
	async function arrive(result: SignedIn, next: string) {
		if (result.password_change_required && result.reset_token) {
			// A temporary password: the user chooses their own before going on.
			let query = `token=${encodeURIComponent(result.reset_token)}&reason=temporary`;
			if (data.request) query += `&request=${encodeURIComponent(data.request)}`;
			// eslint-disable-next-line svelte/no-navigation-without-resolve -- resolved, with a query
			await goto(`${resolve('/reset-password')}?${query}`);
			return;
		}

		if (result.code) {
			// The flow asks for a code emailed to the address. Nobody is
			// signed in until it is typed back.
			challenge = result.code;
			password = '';
			submitting = false;
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
	}

	/** The code was typed right, or the sign-in it belonged to ran out. */
	function signedInWithCode(result: SignedIn) {
		void arrive(result, data.next);
	}

	function codeExpired(message: string) {
		challenge = null;
		submitting = false;
		password = '';
		error = message;
	}

	async function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!canSubmit) return;

		error = '';
		notice = '';
		submitting = true;

		// Read before anything is awaited: once the session changes, this
		// page's data can be replaced under it.
		const next = data.next;

		try {
			const result = await signIn.login({
				request: data.request ?? '',
				email: email.trim(),
				password,
				remember
			});

			await arrive(result, next);
		} catch (err) {
			// A domain that signs in through its organisation's provider: go
			// there, with the address, rather than asking for a password it
			// does not use here.
			const sso = requiredSSO(err);
			if (sso) {
				redirecting = sso.name;
				password = '';
				leaveTo(ssoHref(sso.slug, { request: data.request, next, email: email.trim() }));
				return;
			}

			if (err instanceof ApiError && err.code === 'email_not_verified') {
				notice = messageOf(err, t);
			} else {
				error = messageOf(err, t);
			}
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
{:else if closed}
	<!-- Nothing on this page could work, so none of it is drawn: no form, no
	     provider buttons, no way to a new account. -->
	<AuthCard
		organization={data.organization}
		application={app}
		title={t('login.closed_title')}
		subtitle={app ? t('login.subtitle_app', { app: app.name }) : t('login.subtitle_account')}
	>
		<Alert tone="warning">{t('login.closed_body')}</Alert>
	</AuthCard>
{:else if challenge}
	<!-- The sign-in is held for the code that was emailed. Nothing else on
	     this page can finish it, so nothing else is offered: the buttons that
	     start another sign-in would only lose this one. -->
	<AuthCard
		organization={data.organization}
		application={app}
		title={t('login.code_title')}
		subtitle={t('login.code_subtitle', { email: challenge.email })}
	>
		<CodeForm {challenge} onsignedin={signedInWithCode} onexpired={codeExpired} />

		{#snippet below()}
			<a href={resolve('/login')}>{t('login.code_back')}</a>
		{/snippet}
	</AuthCard>
{:else}
	<AuthCard
		organization={data.organization}
		application={app}
		title={t('login.title')}
		subtitle={app ? t('login.subtitle_app', { app: app.name }) : t('login.subtitle_account')}
	>
		{#if offerPassword}
			<form onsubmit={submit} novalidate>
				{#if error}<Alert>{error}</Alert>{/if}
				{#if notice}<Alert tone="info">{notice}</Alert>{/if}
				{#if redirecting}
					<Alert tone="info">{t('login.sso_redirecting', { name: redirecting })}</Alert>
				{/if}

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

				{#if offerRemember}
					<Checkbox bind:checked={remember} disabled={submitting}>
						{t('login.remember')}
						<span class="aside">{t('login.remember_hint')}</span>
					</Checkbox>
				{/if}

				<Button type="submit" block loading={submitting} disabled={!canSubmit}>
					{submitting ? t('login.submitting') : t('login.submit')}
				</Button>
			</form>
		{/if}

		<SocialButtons
			providers={offerSocial ? data.socialProviders : []}
			request={data.request}
			next={data.next ?? null}
			label={t('login.social')}
			disabled={submitting}
		/>

		<SSOButtons
			connections={data.ssoConnections}
			request={data.request}
			next={data.next ?? null}
			disabled={submitting}
		/>

		{#if data.ssoAvailable}
			<p class="sso-link">
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
				<a href={authHref('/sso', data.request)}>{t('login.sso')}</a>
			</p>
		{/if}

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
	.sso-link {
		margin: var(--space-4) 0 0;
		text-align: center;
		font-size: var(--text-sm);
	}

	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}

	/* The line under "Stay signed in": what ticking it means on a machine
	   that is not yours. */
	.aside {
		display: block;
		color: var(--color-text-muted);
		font-size: var(--text-sm);
	}
</style>
