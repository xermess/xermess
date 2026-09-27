<script lang="ts">
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { RiShieldKeyholeLine } from 'svelte-remixicon';
	import {
		ApiError,
		adminsApi,
		type AdminPermission,
		type AdminPermissionName,
		type AdminRole
	} from '$lib/api';
	import { Alert, Badge, Button, Drawer, FormSection, Input, Textarea } from '$lib/components/ui';
	import Choice from '$lib/components/roles/Choice.svelte';
	import { tidyName } from '$lib/components/roles/roles';
	import { keys } from '$lib/query';
	import { byGroup } from './permissions';

	type Props = {
		/** The role being edited, or null to create one. */
		role: AdminRole | null;
		/** The permission catalog, to offer as ones to grant. */
		catalog: AdminPermission[];
		open: boolean;
	};

	let { role, catalog, open = $bindable(false) }: Props = $props();

	const queryClient = useQueryClient();

	let name = $state('');
	let description = $state('');
	let granted = $state<AdminPermissionName[]>([]);
	let error = $state('');

	/** True while the role is being written. */
	let saving = $state(false);

	const editing = $derived(role !== null);

	/** super_admin is built in: it is shown, never changed. */
	const locked = $derived(role?.is_builtin ?? false);

	const groups = $derived(byGroup(catalog));

	$effect(() => {
		if (!open) return;

		name = role?.name ?? '';
		description = role?.description ?? '';
		granted = role ? [...role.permissions] : [];
		error = '';
	});

	function toggle(permission: AdminPermissionName, on: boolean) {
		granted = on ? [...granted, permission] : granted.filter((it) => it !== permission);
	}

	function payload() {
		return { name: tidyName(name), description: description.trim(), permissions: granted };
	}

	/** A role changes what its administrators may do, and the administrators
	    list shows what each holds, so both are refilled. */
	const save = createMutation(() => ({
		mutationFn: () =>
			role ? adminsApi.updateRole(role.id, payload()) : adminsApi.createRole(payload()),
		onSuccess: async () => {
			await queryClient.invalidateQueries({ queryKey: keys.admins.all });
			open = false;
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not save this role';
		},
		onSettled: () => {
			saving = false;
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (locked || saving || name.trim() === '') return;

		error = '';
		saving = true;
		save.mutate();
	}
</script>

<Drawer
	bind:open
	title={locked ? 'Built-in role' : editing ? 'Edit admin role' : 'New admin role'}
	meta={role
		? `${role.admin_count} ${role.admin_count === 1 ? 'administrator' : 'administrators'}`
		: undefined}
	onsubmit={submit}
>
	{#if error}
		<div class="error"><Alert>{error}</Alert></div>
	{/if}

	{#if locked}
		<p class="note">
			super_admin grants every permission, and is the only role that can manage administrators and
			admin roles. It cannot be changed or removed.
		</p>
	{/if}

	<fieldset disabled={locked}>
		<FormSection title="Role" description="What administrators holding this role are called.">
			<Input
				label="Name"
				icon={RiShieldKeyholeLine}
				bind:value={name}
				autocomplete="off"
				autocapitalize="none"
				spellcheck={false}
				placeholder="moderator"
				hint="Lower case letters, numbers, dashes and underscores."
				required
			/>

			<Textarea
				label="Description"
				bind:value={description}
				rows={2}
				placeholder="What administrators holding this role look after"
			/>

			{#if editing}
				<Input label="Role ID" value={role?.id ?? ''} readOnly copyable />
			{/if}
		</FormSection>

		<FormSection
			title="Permissions"
			description="What administrators holding this role may do. Held for one application, the role grants only the permissions marked per application, and only there. Managing administrators stays with super_admin."
		>
			{#each groups as [group, permissions] (group)}
				<div class="group">
					<span class="label">{group}</span>

					<div class="list">
						{#each permissions as permission (permission.name)}
							<Choice
								name={permission.name}
								description={permission.description}
								checked={granted.includes(permission.name)}
								onChange={(on) => toggle(permission.name, on)}
								disabled={locked}
							>
								{#snippet badges()}
									{#if permission.scopable}
										<Badge tone="success">per application</Badge>
									{/if}
									{#if permission.name.endsWith('.write')}
										<Badge>write</Badge>
									{/if}
								{/snippet}
							</Choice>
						{/each}
					</div>
				</div>
			{/each}
		</FormSection>
	</fieldset>

	{#snippet footer()}
		<span class="spacer"></span>

		<Button variant="subtle" onclick={() => (open = false)} disabled={saving}>
			{locked ? 'Close' : 'Cancel'}
		</Button>

		{#if !locked}
			<Button type="submit" loading={saving} disabled={saving || name.trim() === ''}>
				{saving ? 'Saving…' : editing ? 'Save changes' : 'Create role'}
			</Button>
		{/if}
	{/snippet}
</Drawer>

<style>
	fieldset {
		min-width: 0;
		margin: 0;
		padding: 0;
		border: none;
	}

	.error {
		margin-bottom: var(--space-4);
	}

	.note {
		margin: 0 0 var(--space-4);
		padding: var(--space-3);
		border-radius: var(--radius-sm);
		background: var(--surface-info);
		font-size: var(--text-sm);
		line-height: 1.4;
	}

	.group + .group {
		margin-top: var(--space-2);
	}

	.group {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}

	.label {
		padding-left: var(--space-2);
		font-size: var(--text-sm);
		font-weight: 600;
	}

	.list {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.spacer {
		flex: 1;
	}
</style>
