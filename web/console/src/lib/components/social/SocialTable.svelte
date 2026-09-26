<script lang="ts">
	import { RiGroupLine, RiKey2Line, RiShareLine, RiToggleLine } from 'svelte-remixicon';
	import type { SocialProvider, SocialSpec } from '$lib/api';
	import { Badge, type Column, DataTable, Icon, Tag } from '$lib/components/ui';
	import { hostOf, markFor, missingFrom } from './providers';

	type Props = {
		providers: SocialProvider[];
		/** What this server knows about each kind, for the mark and the host
		    each one signs people in at. */
		kinds: SocialSpec[];
		/** Called with the provider whose row was chosen. */
		onOpen: (provider: SocialProvider) => void;
		/** Shown in place of the rows when there are none. */
		empty: string;
		/** The ids of the ticked rows, and how to say that they changed. Left
		    out, the rows cannot be ticked. */
		selected?: string[];
		onSelect?: (ids: string[]) => void;
	};

	let { providers, kinds, onOpen, empty, selected, onSelect }: Props = $props();

	const columns: Column[] = [
		{ key: 'provider', label: 'Provider', icon: RiShareLine, min: '13rem' },
		{ key: 'client_id', label: 'Client ID', icon: RiKey2Line, min: '16rem' },
		{ key: 'accounts', label: 'Accounts', icon: RiGroupLine, min: '8rem' },
		{ key: 'status', label: 'Status', icon: RiToggleLine, min: '11rem' }
	];

	function kindOf(provider: SocialProvider): SocialSpec | undefined {
		return kinds.find((kind) => kind.kind === provider.kind);
	}
</script>

<DataTable
	{columns}
	rows={providers}
	{empty}
	{onOpen}
	label={(provider) => `Edit ${provider.name}`}
	{selected}
	{onSelect}
>
	{#snippet row(provider)}
		<td>
			<span class="provider">
				<span class="mark"><Icon icon={markFor(provider.kind)} size="1.05rem" /></span>
				<span class="names">
					<strong>{provider.name}</strong>
					<span class="slug">{provider.slug}</span>
				</span>
			</span>
		</td>

		<td><span class="chip">{provider.client_id}</span></td>

		<td>
			{#if provider.identities > 0}
				<span class="text">{provider.identities.toLocaleString()}</span>
			{:else}
				<span class="empty">N/A</span>
			{/if}
		</td>

		<td>
			<span class="status">
				{#if missingFrom(provider, kindOf(provider))}
					<!-- Registered but not usable yet: the button would fail. -->
					<Tag tone="warning" strong>{missingFrom(provider, kindOf(provider))}</Tag>
				{:else if provider.enabled}
					<Badge tone="success">offered</Badge>
				{:else}
					<Badge>off</Badge>
				{/if}

				{#if !kindOf(provider)?.custom}
					<span class="host">{hostOf(kindOf(provider)?.authorize_url ?? '')}</span>
				{:else}
					<span class="host">{hostOf(provider.authorize_url)}</span>
				{/if}
			</span>
		</td>
	{/snippet}
</DataTable>

<style>
	.provider {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		min-width: 0;
	}

	/* The provider's own mark, in a square the size of the two lines beside
	   it, so every row starts at the same place whatever the logo is. */
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

	.slug,
	.host {
		color: var(--color-text-hint);
		font-family: var(--font-mono);
		font-size: var(--text-sm);
	}

	.chip {
		display: inline-block;
		max-width: 100%;
		padding: 3px 6px;
		border-radius: var(--radius-sm);
		background: var(--color-secondary-alt);
		color: var(--color-text-hint);
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		vertical-align: middle;
	}

	.status {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		min-width: 0;
	}

	.host {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.empty {
		color: var(--color-text-disabled);
	}
</style>
