<script lang="ts">
	import { untrack } from 'svelte';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { RiMailLine, RiShieldUserLine, RiUserLine } from 'svelte-remixicon';
	import {
		ApiError,
		usersApi,
		type Admin,
		type Application,
		type Role,
		type UserField,
		type UserRecord
	} from '$lib/api';
	import {
		Alert,
		Button,
		Drawer,
		FormSection,
		Input,
		List,
		ListItem,
		PasswordInput,
		SwitchField,
		Tabs,
		Tag,
		Thumb
	} from '$lib/components/ui';
	import { keys } from '$lib/query';
	import { markFor as providerMark } from '$lib/components/social/providers';
	import { formatDate, formatRelative } from '$lib/utils/format';
	import RoleMappings from '$lib/components/roles/RoleMappings.svelte';
	import { additional } from './fields';
	import FieldInput from './FieldInput.svelte';

	type Props = {
		/** The user being edited, or null to create one. */
		user: UserRecord | null;
		fields: UserField[];
		/** The applications the administrator can see, to name role scopes. */
		applications: Application[];
		/** Every role the administrator can see, global and application roles. */
		roles: Role[];
		/** The signed-in administrator, whose roles say which roles they may
		    give. */
		admin: Admin;
		open: boolean;
		/** False for an administrator who may look but not change users. Roles
		    are allowed separately, by each role's scope. */
		editable?: boolean;
	};

	let {
		user,
		fields,
		applications,
		roles,
		admin,
		open = $bindable(false),
		editable = true
	}: Props = $props();

	const queryClient = useQueryClient();

	/** The user as last saved here: a new user, once created, carries on as an
	    edit of what was just made, so its roles can be given straight away. */
	let current = $state<UserRecord | null>(null);

	/** The tab on show: the record, or its role mapping. */
	let tab = $state('user');

	/** Said once, after a user is created. */
	let created = $state(false);

	// The built-in fields are the record's own columns, so they are named
	// here; the additional ones are whatever this organisation added.
	let email = $state('');
	let emailVerified = $state(false);
	let firstName = $state('');
	let lastName = $state('');
	let isActive = $state(true);

	// The password is never sent back, so these start empty every time. On
	// an edit, leaving them empty keeps the password the user has.
	let password = $state('');
	let confirmPassword = $state('');
	let isTemporaryPassword = $state(false);

	let values = $state<Record<string, string | boolean>>({});
	let error = $state('');

	/** True while the record is being written. Ours rather than the
	    mutation's own isPending, so a form cannot be left saying "Saving…". */
	let saving = $state(false);

	const editing = $derived(current !== null);

	/** The connection being disconnected, so only its own button spins. */
	let disconnecting = $state('');

	async function disconnect(identity: string) {
		if (!current) return;

		error = '';
		disconnecting = identity;

		try {
			await usersApi.disconnect(current.id, identity);
			await queryClient.invalidateQueries({ queryKey: keys.users.all });

			// The drawer is looking at the record it was given, so the
			// provider goes from it here too rather than after a reopen.
			current = {
				...current,
				social_accounts: current.social_accounts.filter((account) => account.id !== identity)
			};
		} catch (err) {
			error = err instanceof ApiError ? err.message : 'Could not disconnect this provider';
		} finally {
			disconnecting = '';
		}
	}

	/** How many roles the user holds, for the tab's count. The role mapping
	    shares this query, so opening the tab asks for nothing more. */
	const mappings = createQuery(() => ({
		queryKey: keys.users.mappings(current?.id ?? ''),
		queryFn: async () => (await usersApi.roleMappings(current!.id)).roles,
		enabled: open && current !== null
	}));

	/** Shown once the second box has something in it, not while it is still
	    being typed into for the first time. */
	const mismatch = $derived(confirmPassword !== '' && password !== confirmPassword);

	/** A password can only be temporary if there is one. A new user always
	    gets one, so the switch is only ever off-limits on an edit. */
	const hasPassword = $derived(!editing || password !== '' || (current?.has_password ?? false));

	const canSubmit = $derived(
		!saving &&
			editable &&
			email.trim() !== '' &&
			password === confirmPassword &&
			(editing || password !== '')
	);

	/** The added fields, split the way they are filled in: values first, then
	    the yes-or-no answers. */
	const extras = $derived(additional(fields));
	const details = $derived(extras.filter((field) => field.type !== 'bool'));
	const flags = $derived(extras.filter((field) => field.type === 'bool'));

	/** Fill the form whenever the drawer is opened for a different user.
	    Dates are stored as timestamps and edited as days. */
	$effect(() => {
		// Closing puts the panel back on its first tab, so the next user opens
		// on their record rather than on whatever tab was left showing.
		if (!open) {
			tab = 'user';
			return;
		}

		current = user;
		tab = 'user';
		created = false;
		// Filled without tracking what the form reads, so the fields being
		// refilled in the background cannot wipe out what is being typed.
		untrack(() => fill(user));
	});

	function fill(record: UserRecord | null) {
		email = record?.email ?? '';
		emailVerified = record?.email_verified ?? false;
		firstName = record?.first_name ?? '';
		lastName = record?.last_name ?? '';
		isActive = record?.is_active ?? true;
		password = '';
		confirmPassword = '';
		// A password an administrator makes up for someone is one they should
		// replace, so a new user's starts out temporary.
		isTemporaryPassword = record ? record.is_temporary_password : true;
		error = '';

		values = Object.fromEntries(
			extras.map((field) => {
				const stored = record?.data?.[field.name];

				if (field.type === 'bool') return [field.name, stored === true];
				if (field.type === 'date') return [field.name, stored ? String(stored).slice(0, 10) : ''];

				return [field.name, stored == null ? '' : String(stored)];
			})
		);
	}

	/** Empty text is left out entirely, which is how "not set" is stored. */
	function payload() {
		const data: Record<string, unknown> = {};

		for (const field of extras) {
			const value = values[field.name];

			if (field.type === 'bool') {
				data[field.name] = value === true;
			} else if (String(value ?? '').trim() !== '') {
				data[field.name] = value;
			}
		}

		return {
			email: email.trim(),
			email_verified: emailVerified,
			first_name: firstName.trim(),
			last_name: lastName.trim(),
			is_active: isActive,
			is_temporary_password: isTemporaryPassword && hasPassword,
			...(password !== '' ? { password, confirm_password: confirmPassword } : {}),
			data
		};
	}

	/** Writing the record. An edit closes the panel; a new user stays open on
	    its role mapping, since giving roles is usually the next thing to do.
	    Roles count their users, so they are refilled too. */
	const save = createMutation(() => ({
		mutationFn: () =>
			current ? usersApi.update(current.id, payload()) : usersApi.create(payload()),
		onSuccess: async (result: { user: UserRecord }) => {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.users.all }),
				queryClient.invalidateQueries({ queryKey: keys.roles.all })
			]);

			if (current) {
				open = false;
				return;
			}

			current = result.user;
			fill(result.user);
			created = true;
			tab = 'roles';
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not save this user';
		},
		onSettled: () => {
			saving = false;
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		// Enter and a double click both submit; one save at a time is enough,
		// and two would race each other to write the same record. The role
		// mapping saves as it goes, so the form only ever writes the record.
		if (tab !== 'user' || !canSubmit) return;

		error = '';
		saving = true;
		save.mutate();
	}

	const tabs = $derived([
		{ value: 'user', label: 'User', icon: RiUserLine },
		{
			value: 'roles',
			label: 'Role mapping',
			icon: RiShieldUserLine,
			count: mappings.data?.filter((it) => it.assigned).length,
			disabled: !editing
		}
	]);
</script>

<Drawer
	bind:open
	title={editing ? (editable ? 'Edit user' : 'User') : 'New user'}
	meta={current?.email}
	width="44rem"
	onsubmit={submit}
>
	{#if !editing}
		<p class="note">Create the user first; roles are given on the Role mapping tab afterwards.</p>
	{:else if created}
		<p class="note success">
			User created, with the default roles. Assign any others here, or close the panel.
		</p>
	{/if}

	<Tabs {tabs} bind:value={tab} label="User sections">
		{#snippet panel(value)}
			{#if value === 'user'}
				{#if error}
					<div class="error"><Alert>{error}</Alert></div>
				{/if}

				<!-- A disabled fieldset disables every control inside it at once,
				     which is how the tab becomes read-only for someone who may
				     only look. -->
				<fieldset disabled={!editable}>
					<FormSection
						title="Account"
						description="Who the user is, and the address they sign in with."
					>
						<Input
							label="Email"
							icon={RiMailLine}
							bind:value={email}
							type="email"
							autocomplete="off"
							placeholder="user@example.com"
							required
						/>

						<div class="pair">
							<Input label="First name" bind:value={firstName} />
							<Input label="Last name" bind:value={lastName} />
						</div>

						{#if editing}
							<!-- What every other system refers to this record by: shown and
							     easy to copy, never edited. -->
							<Input label="User ID" value={current?.id ?? ''} readOnly copyable />
						{/if}
					</FormSection>

					<FormSection
						title="Password"
						description={editing
							? current?.has_password
								? 'Leave both fields empty to keep the current password.'
								: 'This user has no password yet.'
							: 'At least 8 characters.'}
					>
						<div class="pair">
							<PasswordInput
								label={editing ? 'New password' : 'Password'}
								bind:value={password}
								autocomplete="new-password"
								required={!editing}
							/>
							<PasswordInput
								label="Confirm password"
								bind:value={confirmPassword}
								autocomplete="new-password"
								required={!editing || password !== ''}
							/>
						</div>

						{#if mismatch}
							<p class="mismatch">The passwords do not match.</p>
						{/if}

						<SwitchField
							label="Temporary password"
							description="The user has to choose a new password the next time they sign in."
							bind:checked={isTemporaryPassword}
							disabled={!hasPassword}
						/>
					</FormSection>

					{#if editing && (current?.social_accounts ?? []).length > 0}
						<FormSection
							title="Signs in with"
							description="Accounts elsewhere that reach this one. Disconnecting one leaves the account itself alone."
						>
							<List bordered label="Connected providers">
								{#each current?.social_accounts ?? [] as account (account.id)}
									<ListItem
										title={account.provider}
										description={account.email || `Connected ${formatDate(account.connected_at)}`}
									>
										{#snippet lead()}
											<Thumb icon={providerMark(account.kind)} />
										{/snippet}

										{#snippet end()}
											{#if account.last_login_at}
												<Tag small>used {formatRelative(account.last_login_at)}</Tag>
											{/if}

											{#if editable}
												<Button
													size="sm"
													variant="subtle"
													colorPalette="danger"
													loading={disconnecting === account.id}
													disabled={disconnecting !== ''}
													onclick={() => disconnect(account.id)}
												>
													Disconnect
												</Button>
											{/if}
										{/snippet}
									</ListItem>
								{/each}
							</List>
						</FormSection>
					{/if}

					<FormSection title="Status">
						<div class="switches">
							<SwitchField
								label="Active"
								description="An inactive user cannot sign in to any application."
								bind:checked={isActive}
							/>
							<SwitchField
								label="Email verified"
								description="Whether the user has confirmed they own this address."
								bind:checked={emailVerified}
							/>
						</div>
					</FormSection>

					{#if extras.length > 0}
						<FormSection
							title="Additional fields"
							description="The fields this organisation keeps about its users."
						>
							{#each details as field (field.id)}
								<FieldInput {field} bind:value={values[field.name]} />
							{/each}

							{#if flags.length > 0}
								<div class="flags">
									{#each flags as field (field.id)}
										<FieldInput {field} bind:value={values[field.name]} />
									{/each}
								</div>
							{/if}
						</FormSection>
					{/if}
				</fieldset>
			{:else if current}
				<RoleMappings userId={current.id} {roles} {applications} {admin} />
			{/if}
		{/snippet}
	</Tabs>

	{#snippet footer()}
		<span class="spacer"></span>

		<Button variant="subtle" onclick={() => (open = false)} disabled={saving}>
			{editable && tab === 'user' && !created ? 'Cancel' : 'Close'}
		</Button>

		{#if editable && tab === 'user'}
			<Button type="submit" loading={saving} disabled={!canSubmit}>
				{saving ? 'Saving…' : editing ? 'Save changes' : 'Create user'}
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
		padding: var(--space-2) var(--space-3);
		border-radius: var(--radius-sm);
		background: var(--surface-info);
		font-size: var(--text-sm);
	}

	.note.success {
		background: var(--surface-success);
	}

	.pair {
		display: grid;
		grid-template-columns: 1fr 1fr;
		align-items: start;
		gap: var(--space-3);
	}

	.mismatch {
		margin: calc(var(--space-2) * -1) 0 0;
		color: var(--color-danger);
		font-size: var(--text-sm);
	}

	.switches {
		display: flex;
		flex-direction: column;
	}

	/* The organisation's yes-or-no fields are short, so they sit two to a
	   row where there is room. */
	.flags {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(14rem, 1fr));
		gap: var(--space-2) var(--space-4);
	}

	.spacer {
		flex: 1;
	}

	@media (max-width: 34rem) {
		.pair {
			grid-template-columns: 1fr;
		}
	}
</style>
