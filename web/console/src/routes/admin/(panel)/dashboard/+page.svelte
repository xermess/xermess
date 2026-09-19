<script lang="ts">
	import { invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import {
		RiAppsLine,
		RiBarChartBoxLine,
		RiCodeBoxLine,
		RiComputerLine,
		RiFileList3Line,
		RiGroupLine,
		RiHistoryLine,
		RiRefreshLine,
		RiShieldCheckLine,
		RiUserStarLine
	} from 'svelte-remixicon';
	import ActivityChart from '$lib/components/activity/ActivityChart.svelte';
	import ActivityFeed from '$lib/components/activity/ActivityFeed.svelte';
	import SignInHealth from '$lib/components/activity/SignInHealth.svelte';
	import TopActors from '$lib/components/activity/TopActors.svelte';
	import SessionList from '$lib/components/profile/SessionList.svelte';
	import {
		Icon,
		IconButton,
		LinkButton,
		PageHeader,
		Panel,
		StatCard,
		Tag
	} from '$lib/components/ui';
	import { can, canAnywhere } from '$lib/permissions';
	import { useTranslator } from '$lib/i18n';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const t = useTranslator();

	const admin = $derived(data.admin);
	const overview = $derived(data.overview);

	let refreshing = $state(false);

	async function refresh() {
		refreshing = true;
		try {
			await Promise.all([invalidateAll(), new Promise((done) => setTimeout(done, 400))]);
		} finally {
			refreshing = false;
		}
	}

	const totals = $derived(
		overview
			? {
					events: overview.daily.reduce((sum, day) => sum + day.events, 0),
					refused: overview.daily.reduce((sum, day) => sum + day.failures, 0)
				}
			: { events: 0, refused: 0 }
	);
</script>

<svelte:head>
	<title>Activity · xermess admin</title>
</svelte:head>

<div class="page">
	<PageHeader crumbs={[t('nav.dashboard'), t('nav.activity')]}>
		{#snippet secondary()}
			<IconButton
				icon={RiRefreshLine}
				label="Refresh"
				onclick={refresh}
				loading={refreshing}
				disabled={refreshing}
			/>
		{/snippet}

		{#snippet actions()}
			{#if overview}
				<LinkButton href={resolve('/admin/(panel)/dashboard/logs')} variant="subtle">
					<Icon icon={RiFileList3Line} />
					Logs
				</LinkButton>
			{/if}
		{/snippet}
	</PageHeader>

	{#if overview}
		{@const counts = overview.counts}
		<div class="stats">
			<StatCard
				label="Users"
				value={counts.users}
				icon={RiGroupLine}
				href={can(admin, 'users.read') ? resolve('/admin/(panel)/dashboard/users') : undefined}
			>
				<Tag small>{counts.active_users.toLocaleString()} active</Tag>
				{#if counts.new_users > 0}
					<Tag small tone="success">+{counts.new_users.toLocaleString()} this week</Tag>
				{/if}
			</StatCard>

			<StatCard
				label="Applications"
				value={counts.applications}
				icon={RiAppsLine}
				href={canAnywhere(admin, 'applications.read')
					? resolve('/admin/(panel)/dashboard/applications')
					: undefined}
			>
				<Tag small>{counts.enabled_applications.toLocaleString()} enabled</Tag>
				<Tag small>{counts.user_roles.toLocaleString()} roles</Tag>
			</StatCard>

			<StatCard
				label="APIs"
				value={counts.apis}
				icon={RiCodeBoxLine}
				href={can(admin, 'apis.read') ? resolve('/admin/(panel)/dashboard/apis') : undefined}
			>
				<Tag small>{counts.api_scopes.toLocaleString()} scopes</Tag>
			</StatCard>

			<StatCard
				label="Active sessions"
				value={counts.active_sessions}
				icon={RiComputerLine}
				href={admin.is_super_admin ? resolve('/admin/(panel)/dashboard/admins') : undefined}
			>
				<Tag small>{counts.admins.toLocaleString()} {counts.admins === 1 ? 'admin' : 'admins'}</Tag>
				{#if counts.locked_admins > 0}
					<Tag small tone="warning">{counts.locked_admins} locked</Tag>
				{/if}
			</StatCard>
		</div>

		<div class="row">
			<Panel title="Activity · last 14 days" icon={RiBarChartBoxLine}>
				{#snippet meta()}
					<Tag>{totals.events.toLocaleString()} events</Tag>
					{#if totals.refused > 0}
						<Tag tone="danger" dot>{totals.refused.toLocaleString()} refused</Tag>
					{/if}
				{/snippet}
				<ActivityChart daily={overview.daily} />
			</Panel>

			<Panel title="Sign-ins · last 7 days" icon={RiShieldCheckLine}>
				<SignInHealth signIns={overview.sign_ins} locked={counts.locked_admins} />
			</Panel>
		</div>

		<div class="row">
			<Panel title="Recent activity" icon={RiHistoryLine} flush>
				{#snippet meta()}
					<LinkButton href={resolve('/admin/(panel)/dashboard/logs')} variant="ghost" size="sm">
						View all
					</LinkButton>
				{/snippet}
				<ActivityFeed events={overview.activity} />
			</Panel>

			<div class="side">
				<Panel title="Most active · last 7 days" icon={RiUserStarLine} flush>
					<TopActors actors={overview.top_actors} />
				</Panel>

				<Panel title="Your sessions" icon={RiComputerLine} flush>
					{#snippet meta()}
						<LinkButton href={resolve('/admin/(panel)/profile')} variant="ghost" size="sm">
							Manage
						</LinkButton>
					{/snippet}
					<SessionList sessions={data.sessions} limit={4} />
				</Panel>
			</div>
		</div>
	{:else}
		<div class="narrow">
			<Panel title="Your sessions" icon={RiComputerLine} flush>
				{#snippet meta()}
					<LinkButton href={resolve('/admin/(panel)/profile')} variant="ghost" size="sm">
						Manage
					</LinkButton>
				{/snippet}
				<SessionList sessions={data.sessions} limit={8} />
			</Panel>

			<p class="hint">
				Your roles do not include reading activity, so the numbers and the log are not shown here.
			</p>
		</div>
	{/if}
</div>

<style>
	.page {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
		padding-inline: var(--page-gutter);
		padding-bottom: var(--space-5);
	}

	.stats {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: var(--space-4);
	}

	.row {
		display: grid;
		grid-template-columns: minmax(0, 1.7fr) minmax(0, 1fr);
		align-items: stretch;
		gap: var(--space-4);
	}

	.side {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
		min-width: 0;
	}

	.narrow {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		max-width: 32rem;
	}

	.hint {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	@media (max-width: 80rem) {
		.stats {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (max-width: 64rem) {
		.row {
			grid-template-columns: minmax(0, 1fr);
		}
	}

	@media (max-width: 34rem) {
		.stats,
		.row,
		.side,
		.page {
			gap: var(--space-3);
		}
	}
</style>
