<script lang="ts">
	import { RiAddLine, RiDeleteBinLine } from 'svelte-remixicon';
	import type { Role, SSORoleMapping } from '$lib/api';
	import { Button, IconButton, Icon, Input, Select } from '$lib/components/ui';
	import { resolve } from '$app/paths';
	import { useTranslator } from '$lib/i18n';

	type Props = {
		mappings: SSORoleMapping[];
		/** The roles a group can be given. Empty for an administrator who may
		    not read them, who sees the ids already mapped. */
		roles: Role[];
		/** Whether the administrator may read roles, which tells an empty
		    `roles` because there are none from one they may not see. */
		canRead: boolean;
		readOnly?: boolean;
	};

	let { mappings = $bindable([]), roles, canRead, readOnly = false }: Props = $props();

	const t = useTranslator();

	const options = $derived(
		roles.map((role) => ({
			value: role.id,
			label: role.name,
			description: role.application_id ? t('sso.role_application') : t('sso.role_global')
		}))
	);

	function add() {
		mappings = [...mappings, { group: '', role_id: roles[0]?.id ?? '' }];
	}

	function remove(index: number) {
		mappings = mappings.filter((_, i) => i !== index);
	}
</script>

<div class="mappings">
	{#each mappings as mapping, index (index)}
		<div class="row">
			<Input
				label={t('sso.mapping_group')}
				bind:value={mapping.group}
				placeholder="engineering"
				{readOnly}
				autocomplete="off"
			/>
			{#if roles.length > 0}
				<Select label={t('sso.mapping_role')} bind:value={mapping.role_id} {options} {readOnly} />
			{:else}
				<Input label={t('sso.mapping_role')} value={mapping.role_id} readOnly />
			{/if}
			{#if !readOnly}
				<IconButton
					icon={RiDeleteBinLine}
					label={t('sso.mapping_remove')}
					onclick={() => remove(index)}
				/>
			{/if}
		</div>
	{/each}

	{#if roles.length === 0 && !readOnly && !canRead}
		<p class="note">{t('sso.no_roles')}</p>
	{:else if roles.length === 0 && !readOnly}
		<p class="note">
			{t('sso.no_roles_yet')}
			<a href={resolve('/admin/dashboard/roles')}>{t('sso.open_roles')}</a>
		</p>
	{:else if !readOnly}
		<div>
			<Button variant="subtle" size="sm" onclick={add}>
				<Icon icon={RiAddLine} />
				{t('sso.mapping_add')}
			</Button>
		</div>
	{/if}
</div>

<style>
	.mappings {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}

	.row {
		display: grid;
		grid-template-columns: 1fr 1fr auto;
		align-items: center;
		gap: var(--space-2);
	}

	.note {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.note a {
		color: var(--color-text);
	}

	@media (max-width: 36rem) {
		.row {
			grid-template-columns: 1fr;
		}
	}
</style>
