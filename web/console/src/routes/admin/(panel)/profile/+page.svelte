<script lang="ts">
	import {
		RiComputerLine,
		RiGlobalLine,
		RiLockPasswordLine,
		RiMailLine,
		RiPaletteLine,
		RiShieldKeyholeLine,
		RiUserSettingsLine
	} from 'svelte-remixicon';
	import SessionList from '$lib/components/profile/SessionList.svelte';
	import SignOutButton from '$lib/components/profile/SignOutButton.svelte';
	import {
		Button,
		List,
		ListItem,
		PageContainer,
		PageHeader,
		Panel,
		Select,
		Tag,
		Thumb
	} from '$lib/components/ui';
	import { useTranslator } from '$lib/i18n';
	import { setLanguage } from '$lib/state/language.svelte';
	import { theme } from '$lib/state/theme.svelte';
	import { formatDateTime, formatRelative } from '$lib/utils/format';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const t = useTranslator();

	const admin = $derived(data.admin);

	/** Two letters for the account's thumb. */
	const initials = $derived(
		admin.full_name
			.split(/\s+/)
			.filter(Boolean)
			.slice(0, 2)
			.map((part) => part[0])
			.join('')
			.toUpperCase() || admin.email.slice(0, 2).toUpperCase()
	);

	const activeSessions = $derived(data.sessions.filter((session) => session.active).length);
</script>

<svelte:head><title>{t('profile.head')} · xermess admin</title></svelte:head>

<PageContainer>
	<div class="page">
		<PageHeader crumbs={[t('profile.crumb_account'), t('profile.head')]}>
			{#snippet actions()}
				<SignOutButton />
			{/snippet}
		</PageHeader>

		<!-- Who is signed in, at a glance. -->
		<List bordered label={t('profile.account')}>
			<ListItem title={admin.full_name} description={admin.email}>
				{#snippet lead()}<Thumb text={initials} size="md" />{/snippet}
				{#snippet end()}
					<Tag tone={admin.status === 'active' ? 'success' : 'neutral'} dot strong>
						{admin.status}
					</Tag>
					{#if admin.is_super_admin}
						<Tag tone="info">{t('profile.super_admin')}</Tag>
					{/if}
				{/snippet}
			</ListItem>
		</List>

		<Panel title={t('profile.information')} icon={RiUserSettingsLine} flush>
			<List label={t('profile.information')}>
				<ListItem>
					<span class="detail"><span>{t('profile.username')}</span><b>{admin.username}</b></span>
				</ListItem>
				<ListItem>
					<span class="detail"><span>{t('profile.name')}</span><b>{admin.full_name}</b></span>
				</ListItem>
				<ListItem>
					<span class="detail">
						<span>{t('profile.roles')}</span>
						<span class="tags">
							{#each admin.roles as role (role)}
								<Tag small>{role}</Tag>
							{:else}
								<b>—</b>
							{/each}
						</span>
					</span>
				</ListItem>
				<ListItem>
					<span class="detail">
						<span>{t('profile.last_signed_in')}</span>
						{#if admin.last_login_at}
							<b title={formatDateTime(admin.last_login_at)}>
								{formatDateTime(admin.last_login_at)}
								<small>· {formatRelative(admin.last_login_at)}</small>
							</b>
						{:else}
							<b>{t('profile.never')}</b>
						{/if}
					</span>
				</ListItem>
			</List>
		</Panel>

		<Panel title={t('profile.security')} icon={RiShieldKeyholeLine} flush>
			{#snippet meta()}<Tag small>{t('profile.not_available')}</Tag>{/snippet}
			<List label={t('profile.security')}>
				<ListItem
					title={t('profile.email')}
					description={t('profile.email_hint', { email: admin.email })}
				>
					{#snippet lead()}<Thumb icon={RiMailLine} />{/snippet}
					{#snippet end()}
						<Button size="sm" variant="subtle" disabled>{t('action.change')}</Button>
					{/snippet}
				</ListItem>
				<ListItem title={t('profile.password')} description={t('profile.password_hint')}>
					{#snippet lead()}<Thumb icon={RiLockPasswordLine} />{/snippet}
					{#snippet end()}
						<Button size="sm" variant="subtle" disabled>{t('action.change')}</Button>
					{/snippet}
				</ListItem>
			</List>
		</Panel>

		<Panel title={t('profile.preferences')} icon={RiPaletteLine} flush>
			<List label={t('profile.preferences')}>
				<ListItem title={t('profile.theme')} description={t('profile.theme_hint')}>
					{#snippet lead()}<Thumb icon={RiPaletteLine} />{/snippet}
					{#snippet end()}
						<span class="choices" role="group" aria-label={t('profile.theme')}>
							{#each ['light', 'dark'] as const as option (option)}
								<button
									type="button"
									class="control choice"
									data-size="sm"
									data-variant="outline"
									data-palette="neutral"
									class:selected={theme.current === option}
									aria-pressed={theme.current === option}
									onclick={() => theme.set(option)}
								>
									<span class="swatch {option}"></span>
									{option === 'light' ? t('profile.theme_light') : t('profile.theme_dark')}
								</button>
							{/each}
						</span>
					{/snippet}
				</ListItem>
				<ListItem title={t('profile.language')} description={t('profile.language_hint')}>
					{#snippet lead()}<Thumb icon={RiGlobalLine} />{/snippet}
					{#snippet end()}
						<!-- Every language with some of the panel translated, as the
						     server lists them. What users are offered on the sign-in
						     pages is the Languages page's business, not this one's. -->
						<div class="language">
							<Select
								label={t('profile.language')}
								value={data.language}
								options={data.panelLanguages.map((language) => ({
									value: language.code,
									label: language.native,
									description: language.name === language.native ? undefined : language.name
								}))}
								onChange={(code) => code && setLanguage(code)}
							/>
						</div>
					{/snippet}
				</ListItem>
			</List>
		</Panel>

		<Panel title={t('profile.sessions')} icon={RiComputerLine} flush>
			{#snippet meta()}
				<Tag tone={activeSessions > 0 ? 'success' : 'neutral'} dot>
					{t('profile.active_sessions', { count: activeSessions })}
				</Tag>
			{/snippet}
			<SessionList sessions={data.sessions} />
		</Panel>
	</div>
</PageContainer>

<style>
	.page {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}

	/* One fact about the account: its name in the hint colour on the left,
	   the value beside it. */
	.detail {
		display: grid;
		grid-template-columns: 10rem minmax(0, 1fr);
		align-items: center;
		gap: var(--space-3);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.detail b {
		color: var(--color-text);
		font-size: var(--text-base);
		font-weight: normal;
	}

	.detail small {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.tags {
		display: flex;
		flex-wrap: wrap;
		gap: 5px;
	}

	.choices {
		display: flex;
		gap: 5px;
	}

	/* A list that grows with every language added, so a select rather than a
	   row of buttons. */
	.language {
		width: 14rem;
	}

	/* The shape is the shared control; being the chosen one is this page's
	   own business. */
	.choice.selected {
		border-color: var(--color-text);
	}

	.swatch {
		width: 14px;
		height: 14px;
		border: 1px solid var(--color-border);
		border-radius: 4px;
	}

	.swatch.light {
		background: #fff;
	}

	.swatch.dark {
		background: #1c1c1c;
	}

	@media (max-width: 34rem) {
		.detail {
			grid-template-columns: 1fr;
			gap: 2px;
		}
	}
</style>
