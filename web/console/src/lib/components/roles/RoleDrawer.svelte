<script lang="ts">
	import { untrack } from 'svelte';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { RiAddLine, RiCloseLine, RiGlobalLine, RiShieldUserLine } from 'svelte-remixicon';
	import { ApiError, rolesApi, type Admin, type API, type Application, type Role } from '$lib/api';
	import {
		Alert,
		Badge,
		Button,
		Drawer,
		FormSection,
		Icon,
		IconButton,
		Input,
		SearchInput,
		Select,
		SwitchField,
		Textarea,
		type SelectOption
	} from '$lib/components/ui';
	import { keys } from '$lib/query';
	import Choice from './Choice.svelte';
	import RolePicker from './RolePicker.svelte';
	import { canEditRoles, mayInherit, reachable, scopeName, tidyName, wouldCycle } from './roles';

	type Props = {
		/** The role being edited, or null to create one. */
		role: Role | null;
		/** Where a new role starts out: an application's id, or null for a
		    global role. It can be changed before the role is created. */
		initialScope: string | null;
		/** Every role the administrator can see, to offer as ones to include. */
		roles: Role[];
		/** The applications the administrator can see. */
		applications: Application[];
		/** Every API, to offer scopes from; null for an administrator who
		    cannot see APIs, whose edits then leave the role's grants alone. */
		apis: API[] | null;
		admin: Admin;
		open: boolean;
	};

	let {
		role,
		initialScope,
		roles,
		applications,
		apis,
		admin,
		open = $bindable(false)
	}: Props = $props();

	/** The ids of the API scopes the role grants. */
	let granted = $state<string[]>([]);

	/** Whether the list that adds API scopes is open, and what is picked in
	    it. */
	let addingScopes = $state(false);
	let pickedScopes = $state<string[]>([]);
	let scopeSearch = $state('');

	/** The granted scopes, grouped under their API. */
	const grantedByApi = $derived(
		(apis ?? [])
			.map((api) => ({ api, scopes: api.scopes.filter((scope) => granted.includes(scope.id)) }))
			.filter((group) => group.scopes.length > 0)
	);

	/** The scopes that can still be added, matching the search, grouped under
	    their API. */
	const scopeCandidates = $derived(
		(apis ?? [])
			.map((api) => ({
				api,
				scopes: api.scopes.filter((scope) => {
					const wanted = scopeSearch.trim().toLowerCase();
					return (
						!granted.includes(scope.id) &&
						(wanted === '' ||
							[scope.name, scope.description, api.name, api.identifier].some((text) =>
								text.toLowerCase().includes(wanted)
							))
					);
				})
			}))
			.filter((group) => group.scopes.length > 0)
	);

	const queryClient = useQueryClient();

	/** Where the role applies: "global", or an application's id. */
	let target = $state('global');
	let name = $state('');
	let description = $state('');
	let isDefault = $state(false);
	let inherits = $state<string[]>([]);

	/** Whether the list that adds included roles is open, and what is picked
	    in it. */
	let adding = $state(false);
	let picked = $state<string[]>([]);

	let error = $state('');
	let saving = $state(false);

	const editing = $derived(role !== null);

	/** The role's scope: its application, or null for a global role. */
	const scope = $derived(
		role ? role.application_id : target === 'global' ? null : target || undefined
	);

	/** The applications this administrator may make roles in. */
	const writableApps = $derived(applications.filter((app) => canEditRoles(admin, app.id)));
	const mayWriteGlobal = $derived(canEditRoles(admin, null));

	/** The scopes a role can be made in: global, when the administrator may,
	    and each application they may make roles in. An existing role shows
	    its own scope, which cannot change. */
	const targetOptions = $derived<SelectOption[]>([
		...((role ? role.application_id === null : mayWriteGlobal)
			? [
					{
						value: 'global',
						label: 'Global',
						description: 'The same role in every application',
						icon: RiGlobalLine
					}
				]
			: []),
		...(role ? applications.filter((app) => app.id === role.application_id) : writableApps).map(
			(app) => ({
				value: app.id,
				label: app.name,
				description: `Only in ${app.name}`,
				icon: RiShieldUserLine
			})
		)
	]);

	const editable = $derived(role ? canEditRoles(admin, role.application_id) : true);

	const ready = $derived(scope !== undefined && name.trim() !== '');

	/** The roles included directly, as roles. */
	const direct = $derived(roles.filter((other) => inherits.includes(other.id)));

	/** Everything a user holding this role also holds, inheritance followed. */
	const effective = $derived(
		roles.filter((other) => reachable(inherits, roles).has(other.id) && other.id !== role?.id)
	);

	/** The roles that can still be added: allowed by the scope rule, not this
	    role, not included already, and in a scope the administrator manages
	    roles of — including a role hands it out, so the server asks that too. */
	const candidates = $derived(
		scope === undefined
			? []
			: roles.filter(
					(other) =>
						other.id !== role?.id &&
						!inherits.includes(other.id) &&
						mayInherit(scope, other) &&
						canEditRoles(admin, other.application_id)
				)
	);

	$effect(() => {
		if (!open) return;

		const start = role ? role.application_id : initialScope;
		target = start ?? 'global';

		// Start somewhere the administrator may actually make a role.
		// Read without tracking: the options are refilled when the
		// applications are, and that must not reset a form being filled in.
		const options = untrack(() => targetOptions);
		if (!role && !options.some((option) => option.value === target)) {
			target = options[0]?.value ?? '';
		}

		name = role?.name ?? '';
		description = role?.description ?? '';
		isDefault = role?.is_default ?? false;
		inherits = role?.inherits.map((inherited) => inherited.id) ?? [];
		granted = role?.api_scopes.map((scope) => scope.id) ?? [];
		addingScopes = false;
		pickedScopes = [];
		scopeSearch = '';
		adding = false;
		picked = [];
		error = '';
	});

	/** Changing the scope of a new role drops what it can no longer include. */
	function keepAllowed() {
		const now = target === 'global' ? null : target;
		inherits = inherits.filter((id) => {
			const other = roles.find((it) => it.id === id);
			return other !== undefined && mayInherit(now, other);
		});
	}

	function label(other: Role): string {
		return other.application_id === scope
			? other.name
			: `${scopeName(other, applications)} / ${other.name}`;
	}

	function addPicked() {
		inherits = [...inherits, ...picked];
		picked = [];
		adding = false;
	}

	function payload() {
		return {
			application_id: scope ?? null,
			name: tidyName(name),
			description: description.trim(),
			is_default: isDefault,
			inherits,
			// Sent only by someone who can see APIs; left out, the role keeps
			// the grants it has.
			...(apis ? { api_scopes: granted } : {})
		};
	}

	/** A role changes what its users hold and how the users table reads, so
	    both lists and every open role mapping are refilled after a write. */
	const save = createMutation(() => ({
		mutationFn: () => (role ? rolesApi.update(role.id, payload()) : rolesApi.create(payload())),
		onSuccess: async () => {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.roles.all }),
				queryClient.invalidateQueries({ queryKey: keys.users.all })
			]);
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

		if (!editable || saving || !ready) return;

		error = '';
		saving = true;
		save.mutate();
	}
</script>

<Drawer
	bind:open
	title={editing ? (editable ? 'Edit role' : 'Role') : 'New role'}
	meta={role ? `${role.user_count} ${role.user_count === 1 ? 'user' : 'users'}` : undefined}
	onsubmit={submit}
>
	{#if error}
		<div class="error"><Alert>{error}</Alert></div>
	{/if}

	<fieldset disabled={!editable}>
		<FormSection title="Role" description="What applications check for, and where it applies.">
			<Select
				label="Applies to"
				bind:value={target}
				options={targetOptions}
				placeholder="Choose where it applies"
				readOnly={editing}
				hint={editing
					? 'A role keeps where it applies once it is made.'
					: scope === null
						? 'A global role is the same in every application.'
						: 'An application role only means something in that application.'}
				onChange={keepAllowed}
				required
			/>

			<Input
				label="Name"
				bind:value={name}
				autocomplete="off"
				autocapitalize="none"
				spellcheck={false}
				placeholder={scope === null ? 'employee' : 'editor'}
				hint="Lower case letters, numbers, dashes and underscores."
				required
			/>

			<Textarea
				label="Description"
				bind:value={description}
				rows={2}
				placeholder="Who should hold this role, and what it is for"
			/>

			{#if editing}
				<Input label="Role ID" value={role?.id ?? ''} readOnly copyable />
			{/if}

			<SwitchField
				label="Default role"
				description="Give this role to every new user automatically."
				bind:checked={isDefault}
			/>
		</FormSection>

		<FormSection
			title="Included roles"
			description={scope === null
				? 'Anyone with this role also gets the roles it includes. A global role can include any role.'
				: 'Anyone with this role also gets the roles it includes: global roles, or roles of the same application.'}
		>
			{#snippet action()}
				{#if editable && !adding}
					<Button
						size="sm"
						variant="subtle"
						onclick={() => (adding = true)}
						disabled={scope === undefined}
					>
						<Icon icon={RiAddLine} />
						Add roles
					</Button>
				{/if}
			{/snippet}

			{#if adding}
				<div class="adding">
					<RolePicker
						roles={candidates}
						{applications}
						bind:picked
						reason={(other) =>
							wouldCycle(role?.id, other.id, roles)
								? `Already includes ${role?.name}, so this would go round in a circle.`
								: undefined}
						empty="There are no other roles this role can include."
					/>

					<div class="adding-foot">
						<Button
							size="sm"
							variant="subtle"
							onclick={() => {
								adding = false;
								picked = [];
							}}
						>
							Cancel
						</Button>
						<Button size="sm" onclick={addPicked} disabled={picked.length === 0}>
							Add{picked.length > 0 ? ` ${picked.length}` : ''}
						</Button>
					</div>
				</div>
			{/if}

			{#if direct.length > 0}
				<ul class="included">
					{#each direct as other (other.id)}
						<li>
							<Icon
								icon={other.application_id === null ? RiGlobalLine : RiShieldUserLine}
								size="0.9375rem"
							/>
							<span class="name">{label(other)}</span>
							<span class="description">{other.description}</span>
							{#if editable}
								<IconButton
									icon={RiCloseLine}
									label="Stop including {other.name}"
									size="sm"
									placement="left"
									onclick={() => (inherits = inherits.filter((id) => id !== other.id))}
								/>
							{/if}
						</li>
					{/each}
				</ul>
			{:else if !adding}
				<p class="none">This role does not include other roles.</p>
			{/if}

			{#if effective.length > direct.length}
				<div class="effective">
					<span>Counting what those include, users also get</span>
					<div class="badges">
						{#each effective as other (other.id)}
							<Badge>{label(other)}</Badge>
						{/each}
					</div>
				</div>
			{/if}
		</FormSection>
		{#if apis}
			<FormSection
				title="API permissions"
				description="Scopes this role grants users. A token carries one only for its own API, and only when the application is allowed it too."
			>
				{#snippet action()}
					{#if editable && !addingScopes}
						<Button size="sm" variant="subtle" onclick={() => (addingScopes = true)}>
							<Icon icon={RiAddLine} />
							Add scopes
						</Button>
					{/if}
				{/snippet}

				{#if addingScopes}
					<div class="adding">
						<SearchInput
							label="Search API scopes"
							placeholder="Search by scope or API…"
							size="sm"
							bind:value={scopeSearch}
							onkeydown={(event) => event.key === 'Enter' && event.preventDefault()}
						/>

						<div class="scope-list">
							{#each scopeCandidates as group (group.api.id)}
								<div class="scope-group">
									<span class="api-name">
										{group.api.name}
										<code>{group.api.identifier}</code>
									</span>
									{#each group.scopes as scope (scope.id)}
										<Choice
											name={scope.name}
											description={scope.description}
											checked={pickedScopes.includes(scope.id)}
											onChange={(on) =>
												(pickedScopes = on
													? [...pickedScopes, scope.id]
													: pickedScopes.filter((id) => id !== scope.id))}
										/>
									{/each}
								</div>
							{:else}
								<p class="none">
									{(apis ?? []).length === 0
										? 'No APIs are registered yet.'
										: 'No more scopes to add, or none match this.'}
								</p>
							{/each}
						</div>

						<div class="adding-foot">
							<Button
								size="sm"
								variant="subtle"
								onclick={() => {
									addingScopes = false;
									pickedScopes = [];
								}}
							>
								Cancel
							</Button>
							<Button
								size="sm"
								onclick={() => {
									granted = [...granted, ...pickedScopes];
									pickedScopes = [];
									addingScopes = false;
								}}
								disabled={pickedScopes.length === 0}
							>
								Add{pickedScopes.length > 0 ? ` ${pickedScopes.length}` : ''}
							</Button>
						</div>
					</div>
				{/if}

				{#if grantedByApi.length > 0}
					<div class="grants">
						{#each grantedByApi as group (group.api.id)}
							<div class="grant-row">
								<span class="grant-api" title={group.api.identifier}>{group.api.name}</span>
								<ul class="scope-chips">
									{#each group.scopes as scope (scope.id)}
										<li class="scope-chip" title={scope.description}>
											{scope.name}
											{#if editable}
												<button
													type="button"
													aria-label="Stop granting {scope.name}"
													onclick={() => (granted = granted.filter((id) => id !== scope.id))}
												>
													<Icon icon={RiCloseLine} size="0.8125rem" />
												</button>
											{/if}
										</li>
									{/each}
								</ul>
							</div>
						{/each}
					</div>
				{:else if !addingScopes}
					<p class="none">This role grants no API scopes.</p>
				{/if}
			</FormSection>
		{/if}
	</fieldset>

	{#snippet footer()}
		<span class="spacer"></span>

		<Button variant="subtle" onclick={() => (open = false)} disabled={saving}>
			{editable ? 'Cancel' : 'Close'}
		</Button>

		{#if editable}
			<Button type="submit" loading={saving} disabled={saving || !ready}>
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

	.adding {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		padding: var(--space-3);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		background: var(--color-surface);
	}

	.adding-foot {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-2);
	}

	.included {
		margin: 0;
		padding: 0;
		overflow: hidden;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		list-style: none;
	}

	.included li {
		display: grid;
		grid-template-columns: auto auto 1fr auto;
		align-items: center;
		gap: var(--space-2);
		min-height: 44px;
		padding: 0 var(--space-2) 0 var(--space-3);
		color: var(--color-text-hint);
	}

	.included li + li {
		border-top: 1px solid var(--color-border);
	}

	.name {
		color: var(--color-text);
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		font-weight: 600;
		white-space: nowrap;
	}

	.description {
		overflow: hidden;
		font-size: var(--text-sm);
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.none {
		margin: 0;
		padding: var(--space-3);
		border: 1px dashed var(--color-border);
		border-radius: var(--radius-md);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		text-align: center;
	}

	.effective {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.badges {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-1);
	}

	.scope-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		max-height: 20rem;
		overflow-y: auto;
	}

	.scope-group {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.api-name {
		display: flex;
		flex-wrap: wrap;
		align-items: baseline;
		gap: var(--space-2);
		padding: 0 var(--space-2);
		font-size: var(--text-sm);
		font-weight: 600;
	}

	.api-name code {
		color: var(--color-text-hint);
		font-family: var(--font-mono);
		font-size: var(--text-xs);
		font-weight: 400;
	}

	.grants {
		overflow: hidden;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
	}

	.grant-row {
		display: grid;
		grid-template-columns: 8rem 1fr;
		align-items: start;
		gap: var(--space-3);
		padding: 10px var(--space-3);
	}

	.grant-row + .grant-row {
		border-top: 1px solid var(--color-border);
	}

	.grant-api {
		overflow: hidden;
		line-height: 26px;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		font-weight: 600;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.scope-chips {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	.scope-chip {
		display: inline-flex;
		align-items: center;
		gap: 2px;
		height: 26px;
		padding: 0 3px 0 10px;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-pill);
		background: var(--color-secondary-alt);
		font-family: var(--font-mono);
		font-size: var(--text-sm);
	}

	.scope-chip button {
		display: grid;
		place-items: center;
		width: 20px;
		height: 20px;
		padding: 0;
		border: none;
		border-radius: var(--radius-pill);
		background: transparent;
		color: var(--color-text-hint);
		cursor: pointer;
	}

	.scope-chip button:hover {
		background: var(--surface-danger);
		color: var(--color-danger);
	}

	@media (max-width: 30rem) {
		.grant-row {
			grid-template-columns: 1fr;
			gap: var(--space-2);
		}
	}

	.spacer {
		flex: 1;
	}

	@media (max-width: 30rem) {
		.included li {
			grid-template-columns: auto 1fr auto;
		}

		.description {
			display: none;
		}
	}
</style>
