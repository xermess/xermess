<script lang="ts">
	import {
		RiBuilding2Line,
		RiGlobalLine,
		RiGroupLine,
		RiShieldKeyholeLine,
		RiToggleLine
	} from 'svelte-remixicon';
	import type { SSOConnection } from '$lib/api';
	import { Badge, type Column, DataTable, Icon, Tag } from '$lib/components/ui';
	import { useTranslator } from '$lib/i18n';

	type Props = {
		connections: SSOConnection[];
		/** Called with the connection whose row was chosen. */
		onOpen: (connection: SSOConnection) => void;
		empty: string;
	};

	let { connections, onOpen, empty }: Props = $props();

	const t = useTranslator();

	const columns = $derived<Column[]>([
		{ key: 'connection', label: t('sso.column_connection'), icon: RiBuilding2Line, min: '14rem' },
		{ key: 'domains', label: t('sso.column_domains'), icon: RiGlobalLine, min: '14rem' },
		{ key: 'users', label: t('sso.column_users'), icon: RiGroupLine, min: '7rem' },
		{ key: 'status', label: t('sso.column_status'), icon: RiToggleLine, min: '14rem' }
	]);
</script>

<DataTable {columns} rows={connections} {empty} {onOpen} label={(connection) => connection.name}>
	{#snippet row(connection)}
		<td>
			<span class="connection">
				<span class="mark"><Icon icon={RiShieldKeyholeLine} size="1.05rem" /></span>
				<span class="names">
					<strong>{connection.name}</strong>
					<span class="protocol">
						{connection.protocol === 'saml' ? t('sso.protocol_saml') : t('sso.protocol_oidc')}
					</span>
				</span>
			</span>
		</td>

		<td>
			<span class="domains">
				{#each connection.domains.slice(0, 3) as domain (domain)}
					<span class="chip">{domain}</span>
				{/each}
				{#if connection.domains.length > 3}
					<span class="more">+{connection.domains.length - 3}</span>
				{/if}
			</span>
		</td>

		<td>
			{#if connection.users > 0}
				<span>{connection.users.toLocaleString()}</span>
			{:else}
				<span class="none">—</span>
			{/if}
		</td>

		<td>
			<span class="status">
				{#if connection.enabled}
					<Badge tone="success">{t('sso.status_on')}</Badge>
				{:else}
					<Badge>{t('sso.status_off')}</Badge>
				{/if}
				{#if connection.enforce_domains}
					<Tag tone="info" small>{t('sso.status_enforced')}</Tag>
				{/if}
				{#if connection.show_on_login}
					<Tag small>{t('sso.status_on_login')}</Tag>
				{/if}
			</span>
		</td>
	{/snippet}
</DataTable>

<style>
	.connection {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		min-width: 0;
	}

	.mark {
		display: grid;
		place-items: center;
		flex-shrink: 0;
		width: 30px;
		height: 30px;
		border: 1px solid var(--color-secondary-alt);
		border-radius: var(--radius-sm);
		background: var(--color-surface-alt);
		color: var(--color-text-hint);
	}

	.names {
		display: flex;
		flex-direction: column;
		min-width: 0;
		line-height: 1.3;
	}

	.protocol {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.domains,
	.status {
		display: inline-flex;
		align-items: center;
		flex-wrap: wrap;
		gap: var(--space-1);
	}

	.chip {
		padding: 3px 6px;
		border-radius: var(--radius-sm);
		background: var(--color-secondary-alt);
		color: var(--color-text-hint);
		font-family: var(--font-mono);
		font-size: var(--text-sm);
	}

	.more,
	.none {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}
</style>
