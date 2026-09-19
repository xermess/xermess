<script lang="ts">
	import {
		RiComputerLine,
		RiGlobalLine,
		RiLoginBoxLine,
		RiLogoutBoxRLine,
		RiTimeLine,
		RiUserLine
	} from 'svelte-remixicon';
	import type { UserSessionRecord } from '$lib/api';
	import { Button, DataTable, Icon, type Column } from '$lib/components/ui';
	import { useTranslator } from '$lib/i18n';
	import { describeUserAgent, formatDateTime, formatRelative } from '$lib/utils/format';

	type Props = {
		sessions: UserSessionRecord[];
		/** What the table says with no rows: nobody signed in, or nobody
		    matching the search. */
		empty: string;
		/** Given when the administrator may end sessions. */
		onEnd?: (session: UserSessionRecord) => void;
		onSignOutUser?: (session: UserSessionRecord) => void;
		/** Narrows the list to one user's sessions. */
		onUser: (session: UserSessionRecord) => void;
		/** The session being ended, whose button shows it. */
		ending?: string;
	};

	let { sessions, empty, onEnd, onSignOutUser, onUser, ending }: Props = $props();

	const t = useTranslator();

	const columns: Column[] = $derived([
		{ key: 'user', label: t('sessions.column_user'), icon: RiUserLine, min: '16rem' },
		{ key: 'device', label: t('sessions.column_device'), icon: RiComputerLine, min: '12rem' },
		{ key: 'ip', label: t('sessions.column_ip'), icon: RiGlobalLine, min: '8rem' },
		{ key: 'signed_in', label: t('sessions.column_signed_in'), icon: RiLoginBoxLine, min: '9rem' },
		{ key: 'expires', label: t('sessions.column_expires'), icon: RiTimeLine, min: '11rem' },
		...(onEnd ? [{ key: 'actions', label: '', min: '16rem', align: 'end' as const }] : [])
	]);
</script>

<DataTable {columns} rows={sessions} {empty}>
	{#snippet row(session)}
		<td>
			<button
				type="button"
				class="user"
				title={t('sessions.only_user', { email: session.user.email })}
				onclick={() => onUser(session)}
			>
				<span class="email">{session.user.email}</span>
				{#if session.user.name}<span class="name">{session.user.name}</span>{/if}
			</button>
		</td>

		<td>{describeUserAgent(session.user_agent)}</td>

		<td><span class="mono">{session.ip || t('sessions.unknown_ip')}</span></td>

		<td>
			<time datetime={session.signed_in_at} title={formatDateTime(session.signed_in_at)}>
				{formatRelative(session.signed_in_at)}
			</time>
		</td>

		<td><span class="mono">{formatDateTime(session.expires_at)}</span></td>

		{#if onEnd && onSignOutUser}
			<td class="actions">
				<Button
					variant="subtle"
					size="sm"
					title={t('sessions.end_title')}
					loading={ending === session.id}
					onclick={() => onEnd(session)}
				>
					{t('sessions.end')}
				</Button>
				<Button
					variant="subtle"
					colorPalette="danger"
					size="sm"
					title={t('sessions.sign_out_user_title', { email: session.user.email })}
					onclick={() => onSignOutUser(session)}
				>
					<Icon icon={RiLogoutBoxRLine} />
					{t('sessions.sign_out_user')}
				</Button>
			</td>
		{/if}
	{/snippet}
</DataTable>

<style>
	.user {
		display: inline-flex;
		flex-direction: column;
		align-items: flex-start;
		padding: 0;
		border: none;
		background: none;
		color: var(--color-text);
		font: inherit;
		text-align: left;
		cursor: pointer;
	}

	.user:hover .email {
		text-decoration: underline;
	}

	.name {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.mono {
		font-family: var(--font-mono);
		font-size: var(--text-sm);
	}

	.actions {
		text-align: right;
		white-space: nowrap;
	}

	.actions :global(button + button) {
		margin-left: var(--space-1);
	}
</style>
