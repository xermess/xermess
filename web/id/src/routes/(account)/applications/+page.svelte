<script lang="ts">
	import { messageOf, useTranslator } from '$lib/i18n';
	import { invalidateAll } from '$app/navigation';
	import { account } from '$lib/api';
	import { Alert, AppMark, Button, Icon, Panel } from '$lib/components';
	import { formatDate, timeAgo } from '$lib/utils/format';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const t = useTranslator();

	/** What each scope lets an application do, said plainly. */
	/** What each scope lets an application do, said to the person whose
	    account it is. The scope's own name is shown for one nobody has
	    written a sentence for. */
	const scopeKeys: Record<string, string> = {
		openid: 'apps.scope_openid',
		profile: 'apps.scope_profile',
		email: 'apps.scope_email',
		offline_access: 'apps.scope_offline',
		roles: 'apps.scope_roles'
	};

	let confirming = $state<string | null>(null);
	let disconnecting = $state<string | null>(null);
	let error = $state('');
	let notice = $state('');

	async function disconnect(clientId: string, name: string) {
		disconnecting = clientId;
		error = '';
		notice = '';

		try {
			await account.disconnect(clientId);
			notice = `${name} can no longer act for you. You’ll be asked to sign in the next time you use it.`;
			confirming = null;
			await invalidateAll();
		} catch (err) {
			error = messageOf(err, t);
		} finally {
			disconnecting = null;
		}
	}
</script>

<svelte:head>
	<title>{t('apps.head')} · {t('account.nav')}</title>
</svelte:head>

<div class="page">
	<div class="heading">
		<h1>{t('apps.title')}</h1>
		<p>{t('apps.description')}</p>
	</div>

	{#if error}<Alert>{error}</Alert>{/if}
	{#if notice}<Alert tone="success">{notice}</Alert>{/if}

	{#if data.applications.length === 0}
		<Panel title={t('apps.empty')}>
			<p class="empty">{t('apps.empty_body')}</p>
		</Panel>
	{:else}
		{#each data.applications as app (app.client_id)}
			<Panel
				title={app.name}
				description={t('apps.last_used', { when: timeAgo(app.last_used_at, t) })}
			>
				{#snippet aside()}
					<AppMark name={app.name} logo={app.logo_uri} size={44} />
				{/snippet}

				<div class="access">
					<h3>{t('apps.it_can')}</h3>
					<ul>
						{#each app.scopes as scope (scope)}
							<li>
								<Icon name="check" size="1rem" />
								{scopeKeys[scope] ? t(scopeKeys[scope]) : scope}
							</li>
						{/each}
					</ul>
				</div>

				<p class="meta">
					{t('apps.connected_since', { when: formatDate(app.authorized_at, t, data.timezone) })}
					{#if app.client_uri}
						·
						<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
						<a href={app.client_uri} target="_blank" rel="noopener noreferrer">
							{t('apps.visit')}
							<Icon name="external" size="0.8rem" />
						</a>
					{/if}
				</p>

				{#snippet footer()}
					{#if confirming === app.client_id}
						<span class="confirm">{t('apps.disconnect_confirm', { app: app.name })}</span>
						<Button variant="secondary" size="sm" onclick={() => (confirming = null)}>
							{t('apps.keep')}
						</Button>
						<Button
							variant="danger"
							size="sm"
							loading={disconnecting === app.client_id}
							onclick={() => disconnect(app.client_id, app.name)}>{t('apps.disconnect')}</Button
						>
					{:else}
						<Button variant="danger" size="sm" onclick={() => (confirming = app.client_id)}
							>{t('apps.disconnect')}</Button
						>
					{/if}
				{/snippet}
			</Panel>
		{/each}
	{/if}
</div>

<style>
	.page {
		display: flex;
		flex-direction: column;
		gap: var(--space-5);
	}

	h1 {
		font-size: var(--text-2xl);
	}

	.heading p {
		margin-top: var(--space-1);
		color: var(--color-text-hint);
	}

	.empty {
		color: var(--color-text-hint);
	}

	h3 {
		margin: 0 0 var(--space-2);
		font-size: var(--text-sm);
		color: var(--color-text-hint);
		font-weight: 600;
	}

	ul {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
		gap: var(--space-2);
		margin: 0;
		padding: 0;
		list-style: none;
	}

	li {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	li :global(svg) {
		color: var(--color-success);
	}

	.meta {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.meta a {
		display: inline-flex;
		align-items: center;
		gap: 2px;
	}

	.confirm {
		margin-right: auto;
		align-self: center;
		font-weight: 600;
	}
</style>
