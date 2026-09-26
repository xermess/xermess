<script lang="ts">
	import { RiKey2Line, RiShareLine, RiShieldUserLine } from 'svelte-remixicon';
	import type { Application, SocialKind, UserField, UserRecord } from '$lib/api';
	import { Badge, DataTable, Icon, Tooltip, type Column } from '$lib/components/ui';
	import { markFor as providerMark } from '$lib/components/social/providers';
	import FieldValue from './FieldValue.svelte';
	import { valueOf } from './fields';
	import { fieldIcons, idIcon } from './fieldIcons';

	type Props = {
		users: UserRecord[];
		fields: UserField[];
		/** The applications the roles belong to, to name them. */
		applications: Application[];
		/** Called with the user whose row was chosen. */
		onOpen: (user: UserRecord) => void;
		/** The ids of the ticked rows, and how to say that they changed. Left
		    out, the rows cannot be ticked. */
		selected?: string[];
		onSelect?: (ids: string[]) => void;
	};

	let { users, fields, applications, onOpen, selected, onSelect }: Props = $props();

	const appNames = $derived(new Map(applications.map((app) => [app.id, app.name])));

	/** The id, then every field the API listed — the built-in ones first and
	    then whatever this organisation added — then the roles. One list, so a
	    new field needs no change here. */
	const columns = $derived<Column[]>([
		{ key: 'id', label: 'ID', icon: idIcon, min: '10rem' },
		...fields.map((field) => ({
			key: field.name,
			label: field.label || field.name,
			icon: fieldIcons[field.type],
			min: field.type === 'email' ? '13rem' : field.type === 'bool' ? '8rem' : '9rem'
		})),
		{ key: 'roles', label: 'Roles', icon: RiShieldUserLine, min: '10rem' },
		{ key: 'signs_in_with', label: 'Signs in with', icon: RiShareLine, min: '9rem' }
	]);

	/** The first characters of the id, which is all anyone reads of it. */
	/** The mark for a provider is the panel's own, so a row and the drawer
	    show the same thing. */
	function markFor(kind: SocialKind) {
		return providerMark(kind);
	}

	function shortId(id: string): string {
		return id.replace(/-/g, '').slice(0, 15);
	}
</script>

<DataTable
	{columns}
	rows={users}
	empty="No users match this."
	{onOpen}
	label={(user) => `Edit ${user.email}`}
	{selected}
	{onSelect}
>
	{#snippet row(user)}
		<td><span class="chip">{shortId(user.id)}</span></td>

		{#each fields as field (field.name)}
			<td><FieldValue type={field.type} value={valueOf(user, field)} /></td>
		{/each}

		<td>
			<span class="roles">
				{#each user.roles ?? [] as role (role.id)}
					<!-- A global role reads as itself; an application role is
					     prefixed with its application, since the same name can be
					     a role in two of them. -->
					<Badge>
						{#if role.application_id}
							<span class="app">{appNames.get(role.application_id) ?? '?'}</span>
						{/if}
						{role.name}
					</Badge>
				{:else}
					<span class="empty">N/A</span>
				{/each}
			</span>
		</td>

		<!-- How this person gets in: a password, an account somewhere else,
		     or both. A record with neither cannot sign in at all. -->
		<td>
			<span class="ways">
				{#each user.social_accounts ?? [] as account (account.id)}
					<Tooltip label="{account.provider} · {account.email || 'no address'}">
						{#snippet children(trigger)}
							<span class="way" {...trigger()}>
								<Icon icon={markFor(account.kind)} size="1rem" />
							</span>
						{/snippet}
					</Tooltip>
				{/each}

				{#if user.has_password}
					<Tooltip label="Signs in with a password">
						{#snippet children(trigger)}
							<span class="way" {...trigger()}><Icon icon={RiKey2Line} size="1rem" /></span>
						{/snippet}
					</Tooltip>
				{/if}

				{#if !user.has_password && (user.social_accounts ?? []).length === 0}
					<span class="empty">N/A</span>
				{/if}
			</span>
		</td>
	{/snippet}
</DataTable>

<style>
	/* The ways in, side by side: one small mark each. */
	.ways {
		display: inline-flex;
		align-items: center;
		gap: var(--space-1);
	}

	.way {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 26px;
		height: 26px;
		border: 1px solid var(--color-secondary-alt);
		border-radius: var(--radius-sm);
		color: var(--color-text-hint);
	}

	.roles {
		display: inline-flex;
		gap: var(--space-1);
	}

	/* The application a role belongs to, quieter than the role: the same
	   name can be a role in two applications. */
	.app {
		margin-right: 4px;
		opacity: 0.65;
	}

	.app::after {
		content: ' /';
	}

	.empty {
		color: var(--color-text-disabled);
		font-size: var(--text-sm);
	}

	/* PocketBase shows the id as a small chip rather than raw text, which
	   stops it competing with the values beside it. */
	.chip {
		display: inline-flex;
		align-items: center;
		height: 25px;
		padding: 0 7px;
		border-radius: var(--radius-sm);
		background: var(--color-secondary-alt);
		color: var(--color-text);
		font-family: var(--font-mono);
		font-size: var(--text-sm);
	}
</style>
