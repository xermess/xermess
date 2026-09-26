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
	import { BRAND } from '$lib/brand';
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
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

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

	const count = (n: number) => n.toLocaleString();

	/** One line from the parts that apply, in a middot-separated list. */
	function note(...parts: (string | false)[]) {
		return parts.filter(Boolean).join(' · ');
	}

	/** What each number is a number of, as the one line its card shows beside
	    it. A part that does not apply this week is left out rather than shown
	    as a zero. */
	const notes = $derived.by(() => {
		if (!overview) return null;

		const counts = overview.counts;

		return {
			users: note(
				`${count(counts.active_users)} active`,
				counts.new_users > 0 && `+${count(counts.new_users)} this week`
			),
			applications: note(
				`${count(counts.enabled_applications)} enabled`,
				`${count(counts.user_roles)} roles`
			),
			apis: note(`${count(counts.api_scopes)} scopes`),
			sessions: note(
				`${count(counts.admins)} ${counts.admins === 1 ? 'admin' : 'admins'}`,
				counts.locked_admins > 0 && `${counts.locked_admins} locked`
			)
		};
	});

	/** The one note that is not a count: an administrator locked out is the
	    exception on this page, so their card is the one that changes colour. */
	const sessionsTone = $derived(
		overview && overview.counts.locked_admins > 0 ? 'warning' : 'neutral'
	);
</script>

<svelte:head>
	<title>Activity · {BRAND.name}</title>
</svelte:head>

<div class="page">
	<PageHeader crumbs={['Dashboard', 'Activity']}>
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

	{#if overview && notes}
		{@const counts = overview.counts}
		<div class="stats">
			<StatCard
				label="Users"
				value={counts.users}
				icon={RiGroupLine}
				note={notes.users}
				href={can(admin, 'users.read') ? resolve('/admin/(panel)/dashboard/users') : undefined}
			/>

			<StatCard
				label="Applications"
				value={counts.applications}
				icon={RiAppsLine}
				note={notes.applications}
				href={canAnywhere(admin, 'applications.read')
					? resolve('/admin/(panel)/dashboard/applications')
					: undefined}
			/>

			<StatCard
				label="APIs"
				value={counts.apis}
				icon={RiCodeBoxLine}
				note={notes.apis}
				href={can(admin, 'apis.read') ? resolve('/admin/(panel)/dashboard/apis') : undefined}
			/>

			<StatCard
				label="Active sessions"
				value={counts.active_sessions}
				icon={RiComputerLine}
				note={notes.sessions}
				tone={sessionsTone}
				href={admin.is_super_admin ? resolve('/admin/(panel)/dashboard/admins') : undefined}
			/>
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
					<SessionList sessions={data.sessions} limit={4} />
				</Panel>
			</div>
		</div>
	{:else}
		<div class="narrow">
			<Panel title="Your sessions" icon={RiComputerLine} flush>
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
	}

	/* The metrics are a strip rather than a row of boxes, so the gap between
	   them is smaller than the gap down to the panels below: they are read
	   together, and a 5px step says so. */
	.stats {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr));
		gap: var(--space-3);
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

	/* Below this a note beside its number has nowhere to go without being cut
	   off, so the cards go down the page instead and each is the full width. */
	@media (max-width: 34rem) {
		.stats {
			grid-template-columns: minmax(0, 1fr);
		}
	}

	@media (max-width: 64rem) {
		.row {
			grid-template-columns: minmax(0, 1fr);
		}
	}

	@media (max-width: 34rem) {
		.row,
		.side,
		.page {
			gap: var(--space-3);
		}
	}
</style>
