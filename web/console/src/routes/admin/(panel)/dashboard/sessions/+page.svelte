<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { RiCloseLine, RiRefreshLine, RiSearchLine } from 'svelte-remixicon';
	import { createInfiniteQuery, createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { messageOf, sessionsApi, type UserSessionRecord } from '$lib/api';
	import { keys, sessionsOptions } from '$lib/query';
	import { Alert, Button, Icon, IconButton, PageHeader } from '$lib/components/ui';
	import SessionTable from '$lib/components/sessions/SessionTable.svelte';
	import { can } from '$lib/permissions';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	const sessions = createInfiniteQuery(() =>
		sessionsOptions({ search: data.search, user: data.user }, data.page)
	);
	const rows = $derived(sessions.data?.pages.flatMap((one) => one.sessions) ?? []);

	const canWrite = $derived(can(data.admin, 'users.write'));

	let search = $derived(data.search);
	let error = $state('');
	let notice = $state('');

	/** The user about to be signed out everywhere, waiting on a yes. */
	let confirming = $state<UserSessionRecord['user'] | null>(null);

	async function apply(changes: { search?: string; user?: string; email?: string }) {
		const params = new SvelteURLSearchParams(page.url.searchParams);

		for (const [key, value] of Object.entries(changes)) {
			if (value) params.set(key, value);
			else params.delete(key);
		}

		const query = params.toString();
		const path = resolve('/admin/(panel)/dashboard/sessions');

		// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto(query ? `${path}?${query}` : path, { keepFocus: true, noScroll: true });
	}

	let timer: ReturnType<typeof setTimeout>;

	function debounced() {
		clearTimeout(timer);
		timer = setTimeout(() => apply({ search }), 250);
	}

	let refreshing = $state(false);

	async function refresh() {
		refreshing = true;

		try {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.sessions.all }),
				new Promise((done) => setTimeout(done, 400))
			]);
		} finally {
			refreshing = false;
		}
	}

	const end = createMutation(() => ({
		mutationFn: (session: UserSessionRecord) => sessionsApi.end(session.id),
		onMutate: () => {
			error = '';
			notice = '';
		},
		onSuccess: () => queryClient.invalidateQueries({ queryKey: keys.sessions.all }),
		onError: (err: unknown) => {
			error = messageOf(err);
		}
	}));

	const signOut = createMutation(() => ({
		mutationFn: (user: UserSessionRecord['user']) => sessionsApi.signOutUser(user.id),
		onMutate: () => {
			error = '';
			notice = '';
		},
		onSuccess: (result, user) => {
			confirming = null;
			notice = `${user.email} was signed out of ${result.sessions} sessions, and ${result.tokens} application tokens were revoked.`;
			return queryClient.invalidateQueries({ queryKey: keys.sessions.all });
		},
		onError: (err: unknown) => {
			error = messageOf(err);
		}
	}));
</script>

<svelte:head><title>Sessions · xermess admin</title></svelte:head>

<div class="heading">
	<PageHeader crumbs={['Dashboard', 'Sessions']}>
		{#snippet secondary()}
			<span class="total">{`${rows.length} shown`}</span>

			<IconButton
				icon={RiRefreshLine}
				label="Refresh the data"
				onclick={refresh}
				loading={refreshing}
				disabled={refreshing}
			/>
		{/snippet}
	</PageHeader>

	<p class="lead">
		Everyone signed in right now: the browser session that keeps them signed in, where it is and
		when it started. Sign one out, or sign a user out everywhere — which also revokes the tokens
		their applications hold, as after a lost device or a compromised account.
	</p>
</div>

<div class="toolbar">
	<form
		class="search"
		onsubmit={(event) => {
			event.preventDefault();
			apply({ search });
		}}
	>
		<Icon icon={RiSearchLine} />
		<input
			type="search"
			placeholder="Search by the start of an email…"
			bind:value={search}
			oninput={debounced}
			aria-label="Search sessions"
		/>
	</form>

	{#if data.user}
		<button
			type="button"
			class="chip"
			title="Show everyone's sessions"
			onclick={() => apply({ user: '', email: '' })}
		>
			{`Only ${data.email || data.user}`}
			<Icon icon={RiCloseLine} />
		</button>
	{/if}
</div>

{#if confirming}
	<div class="gutter banner">
		<Alert tone="warning">
			<span class="confirm">
				{`Sign ${confirming.email} out of every browser and every application?`}
				<span class="buttons">
					<Button variant="subtle" size="sm" onclick={() => (confirming = null)}>Keep</Button>
					<Button
						colorPalette="danger"
						size="sm"
						loading={signOut.isPending}
						onclick={() => confirming && signOut.mutate(confirming)}
					>
						Sign out
					</Button>
				</span>
			</span>
		</Alert>
	</div>
{/if}

{#if error}
	<div class="gutter banner"><Alert>{error}</Alert></div>
{:else if notice}
	<div class="gutter banner"><Alert tone="success">{notice}</Alert></div>
{/if}

<SessionTable
	sessions={rows}
	empty={data.search || data.user
		? 'No session belongs to an address starting like that.'
		: 'Nobody is signed in.'}
	onEnd={canWrite ? (session) => end.mutate(session) : undefined}
	onSignOutUser={canWrite ? (session) => (confirming = session.user) : undefined}
	onUser={(session) => apply({ user: session.user.id, email: session.user.email, search: '' })}
	ending={end.isPending ? end.variables?.id : undefined}
/>

{#if sessions.hasNextPage}
	<div class="more">
		<Button
			variant="subtle"
			loading={sessions.isFetchingNextPage}
			onclick={() => sessions.fetchNextPage()}
		>
			Show more
		</Button>
	</div>
{/if}

<style>
	.heading {
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}

	.lead {
		max-width: 90ch;
		margin: var(--space-2) 0 0;
		color: var(--color-text-hint);
		font-size: var(--text-base);
	}

	.gutter {
		padding-inline: var(--page-gutter);
	}

	.banner {
		margin-bottom: var(--space-3);
	}

	.confirm {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-2);
		width: 100%;
	}

	.buttons {
		display: flex;
		gap: var(--space-1);
	}

	.toolbar {
		display: flex;
		gap: var(--space-2);
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}

	.search {
		flex: 1;
		display: flex;
		align-items: center;
		gap: var(--space-2);
		height: var(--control-height);
		padding: 0 13px;
		border-radius: var(--radius-sm);
		background: var(--color-input);
		color: var(--color-text-hint);
		transition: background-color var(--speed-fast);
	}

	.search:focus-within {
		background: var(--color-input-focus);
	}

	.search input {
		flex: 1;
		min-width: 0;
		border: none;
		background: transparent;
		color: var(--color-text);
		font: inherit;
		font-family: var(--font-sans);
		font-size: var(--text-base);
	}

	.search input:focus {
		outline: none;
	}

	.chip {
		display: flex;
		align-items: center;
		gap: var(--space-1);
		height: var(--control-height);
		padding: 0 var(--space-3);
		border: none;
		border-radius: var(--radius-md);
		background: var(--surface-info);
		color: var(--color-text);
		font: inherit;
		font-size: var(--text-base);
		white-space: nowrap;
		cursor: pointer;
	}

	.more {
		display: flex;
		justify-content: center;
		padding: var(--space-4) var(--page-gutter);
	}
</style>
