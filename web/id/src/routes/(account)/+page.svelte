<script lang="ts">
	import { messageOf, useTranslator } from '$lib/i18n';
	import { invalidateAll } from '$app/navigation';
	import { account } from '$lib/api';
	import { Alert, Button, Icon, Panel, TextField } from '$lib/components';
	import { formatDate, initials } from '$lib/utils/format';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const t = useTranslator();

	const user = $derived(data.user);

	// svelte-ignore state_referenced_locally
	let firstName = $state(data.user.first_name);
	// svelte-ignore state_referenced_locally
	let lastName = $state(data.user.last_name);
	let saving = $state(false);
	let error = $state('');
	let saved = $state(false);

	const changed = $derived(
		firstName.trim() !== user.first_name || lastName.trim() !== user.last_name
	);

	async function save(event: SubmitEvent) {
		event.preventDefault();
		if (!changed || saving) return;

		saving = true;
		error = '';
		saved = false;

		try {
			const { user: updated } = await account.updateProfile({
				first_name: firstName.trim(),
				last_name: lastName.trim()
			});
			firstName = updated.first_name;
			lastName = updated.last_name;
			saved = true;
			await invalidateAll();
		} catch (err) {
			error = messageOf(err, t);
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>{t('profile.head')} · {t('account.nav')}</title>
</svelte:head>

<div class="page">
	<div class="intro">
		<span class="avatar" aria-hidden="true">{initials(user)}</span>
		<div>
			<h1>{`${user.first_name} ${user.last_name}`.trim() || t('account.your_account')}</h1>
			<p>
				{user.email}
				{#if user.email_verified}
					<span class="badge verified">
						<Icon name="check" size="0.8rem" />
						{t('profile.verified')}
					</span>
				{:else}
					<span class="badge">{t('profile.not_verified')}</span>
				{/if}
			</p>
		</div>
	</div>

	<form onsubmit={save}>
		<Panel title={t('profile.personal_title')} description={t('profile.personal_description')}>
			{#if error}<Alert>{error}</Alert>{/if}
			{#if saved && !changed}<Alert tone="success">{t('profile.saved')}</Alert>{/if}

			<div class="pair">
				<TextField
					label={t('field.first_name')}
					bind:value={firstName}
					autocomplete="given-name"
					maxlength={100}
					disabled={saving}
				/>
				<TextField
					label={t('field.last_name')}
					bind:value={lastName}
					autocomplete="family-name"
					maxlength={100}
					disabled={saving}
				/>
			</div>

			<TextField
				label={t('field.email')}
				value={user.email}
				readonly
				hint={t('profile.email_hint')}
			/>

			{#snippet footer()}
				<Button
					variant="secondary"
					disabled={!changed || saving}
					onclick={() => {
						firstName = user.first_name;
						lastName = user.last_name;
					}}>{t('action.cancel')}</Button
				>
				<Button type="submit" loading={saving} disabled={!changed}>
					{t('action.save_changes')}
				</Button>
			{/snippet}
		</Panel>
	</form>

	<Panel title={t('profile.details_title')}>
		<dl>
			<div>
				<dt>{t('profile.member_since')}</dt>
				<dd>{formatDate(user.created_at, t)}</dd>
			</div>
			<div>
				<dt>{t('profile.last_sign_in')}</dt>
				<dd>{user.last_login_at ? formatDate(user.last_login_at, t) : t('time.never')}</dd>
			</div>
			<div>
				<dt>{t('profile.account_id')}</dt>
				<dd><code>{user.id}</code></dd>
			</div>
		</dl>
	</Panel>
</div>

<style>
	.page {
		display: flex;
		flex-direction: column;
		gap: var(--space-5);
	}

	.intro {
		display: flex;
		align-items: center;
		gap: var(--space-4);
		margin-bottom: var(--space-2);
	}

	.avatar {
		display: grid;
		place-items: center;
		flex-shrink: 0;
		width: 64px;
		height: 64px;
		border-radius: 50%;
		background: var(--color-accent);
		color: var(--color-accent-text);
		font-size: var(--text-xl);
		font-weight: 700;
	}

	h1 {
		font-size: var(--text-2xl);
		line-height: 1.2;
	}

	.intro p {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-2);
		margin-top: var(--space-1);
		color: var(--color-text-hint);
		overflow-wrap: anywhere;
	}

	.badge {
		display: inline-flex;
		align-items: center;
		gap: 2px;
		padding: 1px 8px;
		border-radius: 999px;
		background: var(--color-input);
		color: var(--color-text-hint);
		font-size: var(--text-xs);
		font-weight: 600;
	}

	.badge.verified {
		background: color-mix(in srgb, var(--color-success), transparent 86%);
		color: var(--color-success);
	}

	.pair {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-4);
	}

	dl {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: var(--space-4);
		margin: 0;
	}

	dt {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	dd {
		margin: var(--space-1) 0 0;
		font-weight: 600;
		overflow-wrap: anywhere;
	}

	code {
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		font-weight: 400;
	}

	@media (max-width: 30rem) {
		.pair {
			grid-template-columns: 1fr;
		}
	}
</style>
