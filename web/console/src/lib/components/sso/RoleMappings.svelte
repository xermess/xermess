<script lang="ts">
	import { RiAddLine, RiDeleteBinLine } from 'svelte-remixicon';
	import type { Role, SSORoleMapping } from '$lib/api';
	import { Button, IconButton, Icon, Input, Select } from '$lib/components/ui';
	import { resolve } from '$app/paths';

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

	const options = $derived(
		roles.map((role) => ({
			value: role.id,
			label: role.name,
			description: role.application_id ? 'Application role' : 'Global role'
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
				label="Group"
				bind:value={mapping.group}
				placeholder="engineering"
				{readOnly}
				autocomplete="off"
			/>
			{#if roles.length > 0}
				<Select label="Role" bind:value={mapping.role_id} {options} {readOnly} />
			{:else}
				<Input label="Role" value={mapping.role_id} readOnly />
			{/if}
			{#if !readOnly}
				<IconButton icon={RiDeleteBinLine} label="Remove mapping" onclick={() => remove(index)} />
			{/if}
		</div>
	{/each}

	{#if roles.length === 0 && !readOnly && !canRead}
		<p class="note">Your roles do not let you read roles, so you cannot map groups to them.</p>
	{:else if roles.length === 0 && !readOnly}
		<p class="note">
			There are no roles yet, so there is nothing to map a group to.
			<a href={resolve('/admin/dashboard/roles')}>Create one on the Roles page</a>
		</p>
	{:else if !readOnly}
		<div>
			<Button variant="subtle" size="sm" onclick={add}>
				<Icon icon={RiAddLine} />
				Add mapping
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
