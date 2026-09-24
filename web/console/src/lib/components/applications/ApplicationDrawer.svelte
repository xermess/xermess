<script lang="ts">
	import { resolve } from '$app/paths';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import {
		RiArrowRightLine,
		RiCodeBoxLine,
		RiRefreshLine,
		RiSettings3Line,
		RiShieldCheckLine
	} from 'svelte-remixicon';
	import {
		ApiError,
		applicationsApi,
		type Admin,
		type Application,
		type ApplicationInput,
		type ApplicationType,
		type ApplicationWithSecret,
		type GrantType,
		type Scope
	} from '$lib/api';
	import {
		Alert,
		Badge,
		Button,
		Drawer,
		FormSection,
		Icon,
		Input,
		Select,
		SwitchField,
		Tabs,
		Textarea,
		type SelectOption
	} from '$lib/components/ui';
	import Choice from '$lib/components/roles/Choice.svelte';
	import { keys, loginFlowChoicesOptions } from '$lib/query';
	import { formatDateTime } from '$lib/utils/format';
	import { authMethods, blank, grants, lines, scopes, types } from './applications';
	import ApplicationApis from './ApplicationApis.svelte';
	import SecretPanel from './SecretPanel.svelte';
	import TokenPreviewPanel from './TokenPreviewPanel.svelte';

	type Props = {
		/** The application being edited, or null to register one. */
		application: Application | null;
		/** Whether the administrator may change this application (or, for a
		    new one, register applications at all). */
		editable: boolean;
		/** The signed-in administrator, for the token preview. */
		admin: Admin;
		open: boolean;
	};

	let { application, editable, admin, open = $bindable(false) }: Props = $props();

	/** The tab on show, once the application exists. */
	let tab = $state('settings');

	const queryClient = useQueryClient();

	/** The application as last saved here, so a registration that returns a
	    secret carries on as an edit of what was just made. */
	let current = $state<Application | null>(null);

	/** The secret the server just made, shown until the panel closes. */
	let secret = $state('');

	let form = $state<ApplicationInput>(blank('web'));
	let redirects = $state('');
	let logoutRedirects = $state('');

	/** Lifetimes are edited in the units people think in. */
	let accessMinutes = $state('60');
	let idMinutes = $state('60');
	let refreshDays = $state('30');

	let error = $state('');
	let saving = $state(false);
	let confirmingRotate = $state(false);

	const editing = $derived(current !== null);
	const kind = $derived(types[form.type]);
	const machine = $derived(form.type === 'm2m');

	/** The flows this application can be pointed at: the default one, and any
	    that is offered. A flow that has been turned off is still listed when
	    this application is the one holding it, so the picker shows what is
	    stored rather than silently reading as the default.

	    An administrator who may not read flows gets none, and the picker is
	    left out: the application keeps whichever flow it has. */
	const flows = createQuery(() => loginFlowChoicesOptions());

	/** What the picker calls "no flow of its own". The request says that with
	    an empty string, which a select reads as nothing chosen — and a flow id
	    is always a uuid, so this cannot be one. */
	const DEFAULT_FLOW = 'default';

	const flowOptions = $derived.by<SelectOption<string>[]>(() => {
		const all = flows.data?.flows ?? [];

		const options: SelectOption<string>[] = [
			{
				value: DEFAULT_FLOW,
				label: 'The default flow',
				description: all.find((flow) => flow.is_default)?.name ?? 'Whichever flow is the default'
			}
		];

		for (const flow of all) {
			if (flow.is_default) continue;
			if (!flow.enabled && flow.id !== form.login_flow_id) continue;

			options.push({
				value: flow.id,
				label: flow.name,
				description: flow.enabled ? flow.description : `${flow.description} · turned off`
			});
		}

		return options;
	});

	const typeOptions: SelectOption<ApplicationType>[] = (
		Object.entries(types) as [ApplicationType, (typeof types)[ApplicationType]][]
	).map(([value, meta]) => ({
		value,
		label: meta.label,
		description: meta.description,
		icon: meta.icon
	}));

	$effect(() => {
		// Closing puts the panel back on its settings, for the next one opened.
		if (!open) {
			tab = 'settings';
			return;
		}

		current = application;
		secret = '';
		error = '';
		confirmingRotate = false;
		fill(application ? asInput(application) : blank('web'));
	});

	/** An application as the form edits it. The flow is the one field the two
	    shapes spell differently: an application that names none has null, and
	    the form — like the request — says so with an empty string. */
	function asInput(app: Application): ApplicationInput {
		return { ...app, login_flow_id: app.login_flow_id ?? '' };
	}

	function fill(input: ApplicationInput) {
		form = {
			...input,
			grant_types: [...input.grant_types],
			scopes: [...input.scopes],
			redirect_uris: [...input.redirect_uris],
			post_logout_redirect_uris: [...input.post_logout_redirect_uris]
		};
		redirects = input.redirect_uris.join('\n');
		logoutRedirects = input.post_logout_redirect_uris.join('\n');
		accessMinutes = String(Math.round(input.access_token_lifetime / 60));
		idMinutes = String(Math.round(input.id_token_lifetime / 60));
		refreshDays = String(Math.round(input.refresh_token_lifetime / 86400));
	}

	/** Picking a type on a new application starts its settings over from what
	    that type should have, keeping only what was typed about it. */
	function changeType(type: ApplicationType) {
		const { name, description, logo_uri, client_uri, policy_uri, tos_uri } = form;
		fill({ ...blank(type), name, description, logo_uri, client_uri, policy_uri, tos_uri });
	}

	function toggle<T extends string>(list: T[], value: T, on: boolean): T[] {
		return on ? [...list, value] : list.filter((it) => it !== value);
	}

	function payload(): ApplicationInput {
		return {
			...form,
			name: form.name.trim(),
			description: form.description.trim(),
			redirect_uris: lines(redirects),
			post_logout_redirect_uris: lines(logoutRedirects),
			access_token_lifetime: Number(accessMinutes) * 60,
			id_token_lifetime: Number(idMinutes) * 60,
			refresh_token_lifetime: Number(refreshDays) * 86400
		};
	}

	/** Whatever was saved, the list and the pickers elsewhere are refilled,
	    and a secret that came back is shown. */
	async function saved(result: ApplicationWithSecret) {
		await queryClient.invalidateQueries({ queryKey: keys.applications.all });

		current = result.application;
		fill(asInput(result.application));

		if (result.client_secret) {
			secret = result.client_secret;
		} else {
			open = false;
		}
	}

	const save = createMutation(() => ({
		mutationFn: () =>
			current ? applicationsApi.update(current.id, payload()) : applicationsApi.create(payload()),
		onSuccess: saved,
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not save this application';
		},
		onSettled: () => {
			saving = false;
		}
	}));

	const rotate = createMutation(() => ({
		mutationFn: () => applicationsApi.rotateSecret(current!.id),
		onSuccess: async (result: ApplicationWithSecret) => {
			confirmingRotate = false;
			await saved(result);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not rotate the secret';
		},
		onSettled: () => {
			saving = false;
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (!editable || tab !== 'settings' || saving || form.name.trim() === '') return;

		error = '';
		saving = true;
		save.mutate();
	}

	function rotateNow() {
		if (!current || saving) return;

		error = '';
		saving = true;
		rotate.mutate();
	}

	/** The roles page, showing this application's roles. */
	const rolesHref = $derived(
		current
			? `${resolve('/admin/(panel)/dashboard/roles')}?tab=application&application=${current.id}`
			: ''
	);

	/** The grants a client of this kind may be offered: a public client has
	    no secret to act as itself with. */
	const offeredGrants = $derived(
		grants.filter((grant) => grant.value !== 'client_credentials' || !kind.public)
	);

	/** The secret as it is shown once made: only its last characters are
	    kept, so the rest is dots. */
	const maskedSecret = $derived(
		current?.has_secret ? `••••••••••••••••${current.secret_hint}` : 'No secret yet'
	);
</script>

<!-- The settings form, shown on its own for a new application and on the
     Settings tab once it exists. -->
{#snippet settings()}
	<fieldset disabled={!editable}>
		<FormSection title="General" description="How the application is named and presented.">
			<Select
				label="Application type"
				value={form.type}
				options={typeOptions}
				readOnly={editing}
				hint={editing
					? 'The type cannot change: copies of the app already in use depend on it.'
					: kind.description}
				onChange={changeType}
			/>

			<Input label="Name" bind:value={form.name} placeholder="Shop" required />

			<Textarea
				label="Description"
				bind:value={form.description}
				rows={2}
				placeholder="What the application is, for other administrators"
			/>

			<div class="pair">
				<Input
					label="Homepage URL"
					bind:value={form.client_uri}
					type="url"
					placeholder="https://shop.example.com"
				/>
				<Input
					label="Logo URL"
					bind:value={form.logo_uri}
					type="url"
					placeholder="https://shop.example.com/logo.png"
					hint="https only."
				/>
			</div>

			<SwitchField
				label="Enabled"
				description="A disabled application cannot sign anyone in."
				bind:checked={form.enabled}
			/>
		</FormSection>

		<FormSection
			title="Credentials"
			description={kind.public
				? 'A public client cannot keep a secret, so it has none and must use PKCE.'
				: 'What the application proves who it is with at the token endpoint.'}
		>
			{#if current}
				<Input label="Client ID" value={current.client_id} readOnly copyable />
			{:else}
				<p class="note">
					The client ID{kind.public ? '' : ' and secret'} are made when you create the application.
				</p>
			{/if}

			{#if !kind.public}
				<Select
					label="Client authentication"
					bind:value={form.token_endpoint_auth_method}
					options={authMethods}
				/>

				{#if current}
					<div class="secret">
						<Input
							label="Client secret"
							value={maskedSecret}
							readOnly
							hint={current.secret_created_at
								? `Created ${formatDateTime(current.secret_created_at)}. Only its last characters are kept.`
								: undefined}
						/>

						{#if editable}
							<div class="rotate">
								{#if confirmingRotate}
									<p>The current secret stops working at once.</p>
									<div class="rotate-actions">
										<Button variant="subtle" size="sm" onclick={() => (confirmingRotate = false)}>
											Keep it
										</Button>
										<Button colorPalette="danger" size="sm" onclick={rotateNow} disabled={saving}>
											Rotate now
										</Button>
									</div>
								{:else}
									<Button
										size="sm"
										variant="subtle"
										onclick={() => (confirmingRotate = true)}
										disabled={saving}
									>
										<Icon icon={RiRefreshLine} />
										Rotate secret
									</Button>
								{/if}
							</div>
						{/if}
					</div>
				{/if}
			{/if}
		</FormSection>

		{#if !machine}
			<FormSection
				title="Redirects"
				description="Where users may be sent back to. Each URI is matched exactly, with no wildcards."
			>
				<Textarea
					label="Redirect URIs"
					bind:value={redirects}
					rows={3}
					placeholder="https://shop.example.com/auth/callback"
					hint={form.type === 'native'
						? 'One per line. https, a loopback address, or a private-use scheme such as com.example.app:/callback.'
						: 'One per line. https, or http on localhost.'}
				/>

				<Textarea
					label="Post-logout redirect URIs"
					bind:value={logoutRedirects}
					rows={2}
					placeholder="https://shop.example.com/"
					hint="One per line. Where the app may send users after they sign out."
				/>
			</FormSection>

			<FormSection
				title="Grants and scopes"
				description="How the application signs users in, and what it may ask for."
			>
				<div class="options">
					<span class="options-label">Grant types</span>
					<div class="grid">
						{#each offeredGrants as grant (grant.value)}
							<Choice
								name={grant.value}
								description={grant.description}
								checked={form.grant_types.includes(grant.value)}
								onChange={(on) =>
									(form.grant_types = toggle<GrantType>(form.grant_types, grant.value, on))}
							/>
						{/each}
					</div>
				</div>

				<div class="options">
					<span class="options-label">Scopes</span>
					<div class="grid">
						{#each scopes as scope (scope.value)}
							<Choice
								name={scope.value}
								description={scope.description}
								checked={form.scopes.includes(scope.value)}
								onChange={(on) => (form.scopes = toggle<Scope>(form.scopes, scope.value, on))}
							/>
						{/each}
					</div>
				</div>

				<SwitchField
					label="Require PKCE"
					description={kind.public
						? 'Always on for a public client.'
						: 'Recommended for every client: it stops a stolen code from being used.'}
					bind:checked={form.require_pkce}
					disabled={kind.public}
				/>
			</FormSection>
		{:else}
			<FormSection
				title="Grant"
				description="A machine-to-machine application uses client_credentials only: it has no user to sign in, so nothing to redirect to."
			>
				<Choice
					name="client_credentials"
					description="Act as the application itself, with no user"
					checked={true}
					onChange={() => {}}
					disabled
				/>
			</FormSection>
		{/if}

		<FormSection
			title="Token lifetimes"
			description="How long each token the application is given stays valid."
		>
			<div class="triple">
				<Input
					label="Access token"
					bind:value={accessMinutes}
					type="number"
					min="1"
					suffix="minutes"
				/>
				{#if !machine}
					<Input label="ID token" bind:value={idMinutes} type="number" min="1" suffix="minutes" />
					<Input
						label="Refresh token"
						bind:value={refreshDays}
						type="number"
						min="1"
						suffix="days"
					/>
				{/if}
			</div>
		</FormSection>

		{#if !machine}
			<FormSection
				title="Sign-in page"
				description="What users see when this application sends them to sign in."
			>
				<div class="pair">
					<Input
						label="Terms of service URL"
						bind:value={form.tos_uri}
						type="url"
						placeholder="https://shop.example.com/terms"
					/>
					<Input
						label="Privacy policy URL"
						bind:value={form.policy_uri}
						type="url"
						placeholder="https://shop.example.com/privacy"
					/>
				</div>
				<p class="note">
					Linked at the bottom of every sign-in page. When either is set, new users have to accept
					them to create an account.
				</p>

				<SwitchField
					label="Allow registration"
					description="Offer “Create an account” on the sign-in page. New users get the default roles."
					bind:checked={form.allow_registration}
				/>

				{#if flowOptions.length > 1}
					<Select
						label="Login flow"
						value={form.login_flow_id || DEFAULT_FLOW}
						onChange={(chosen) => (form.login_flow_id = chosen === DEFAULT_FLOW ? '' : chosen)}
						options={flowOptions}
						hint="What this application's users are taken through when they sign in. Manage the flows themselves on the Login flows page."
					/>
				{/if}
			</FormSection>

			<FormSection title="Roles" description="How the application's roles reach its tokens.">
				{#snippet action()}
					{#if current}
						<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
						<a class="link" href={rolesHref}>
							Manage {current.role_count}
							{current.role_count === 1 ? 'role' : 'roles'}
							<Icon icon={RiArrowRightLine} size="0.875rem" />
						</a>
					{/if}
				{/snippet}

				<div class="switches">
					<SwitchField
						label="Include roles in tokens"
						description="Put the user's roles in this application into its tokens and userinfo."
						bind:checked={form.assert_roles}
					/>
					<SwitchField
						label="Require a role to sign in"
						description="Only let users who hold at least one of its roles sign in to it."
						bind:checked={form.require_role_assignment}
					/>
				</div>
			</FormSection>
		{/if}
	</fieldset>
{/snippet}

<Drawer
	bind:open
	title={editing ? (editable ? 'Edit application' : 'Application') : 'New application'}
	meta={current?.name}
	onsubmit={submit}
>
	{#if secret && current}
		<SecretPanel clientId={current.client_id} {secret} />
	{/if}

	{#if error}
		<div class="error"><Alert>{error}</Alert></div>
	{/if}

	{#if current}
		<!-- What the application is, at a glance, before its settings. -->
		<div class="summary">
			<span class="mark"><Icon icon={kind.icon} size="1.25rem" /></span>
			<div class="summary-text">
				<strong>{current.name}</strong>
				<span>{kind.label} · {kind.public ? 'Public client' : 'Confidential client'}</span>
			</div>
			<Badge tone={current.enabled ? 'success' : 'neutral'}>
				{current.enabled ? 'Enabled' : 'Disabled'}
			</Badge>
		</div>
	{/if}

	{#if current}
		<Tabs
			tabs={[
				{ value: 'settings', label: 'Settings', icon: RiSettings3Line },
				{ value: 'apis', label: 'API access', icon: RiCodeBoxLine },
				{ value: 'preview', label: 'Token preview', icon: RiShieldCheckLine }
			]}
			bind:value={tab}
			label="Application sections"
		>
			{#snippet panel(value)}
				{#if value === 'settings'}
					{@render settings()}
				{:else if value === 'apis'}
					<ApplicationApis applicationId={current!.id} {editable} />
				{:else}
					<TokenPreviewPanel application={current!} {admin} />
				{/if}
			{/snippet}
		</Tabs>
	{:else}
		{@render settings()}
	{/if}

	{#snippet footer()}
		<span class="spacer"></span>

		<Button variant="subtle" onclick={() => (open = false)} disabled={saving}>
			{secret || !editable || tab !== 'settings' ? 'Close' : 'Cancel'}
		</Button>

		{#if editable && tab === 'settings'}
			<Button type="submit" loading={saving} disabled={saving || form.name.trim() === ''}>
				{saving ? 'Saving…' : editing ? 'Save changes' : 'Create application'}
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

	.summary {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		margin-bottom: var(--space-5);
		padding: var(--space-3);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
	}

	.mark {
		display: grid;
		place-items: center;
		width: 40px;
		height: 40px;
		border-radius: var(--radius-md);
		background: var(--color-secondary-alt);
	}

	.summary-text {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}

	.summary-text span {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.note {
		margin: 0;
		padding: var(--space-3);
		border-radius: var(--radius-sm);
		background: var(--color-secondary-alt);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.pair {
		display: grid;
		grid-template-columns: 1fr 1fr;
		align-items: start;
		gap: var(--space-3);
	}

	.triple {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
		gap: var(--space-3);
	}

	.secret {
		display: grid;
		grid-template-columns: 1fr auto;
		align-items: start;
		gap: var(--space-3);
	}

	.secret :global([data-part='input']) {
		font-family: var(--font-mono);
		letter-spacing: 0.02em;
	}

	.rotate {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: var(--space-2);
		padding-top: 14px;
	}

	.rotate p {
		margin: 0;
		color: var(--color-danger);
		font-size: var(--text-sm);
		white-space: nowrap;
	}

	.rotate-actions {
		display: flex;
		gap: var(--space-2);
	}

	.options {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}

	.options-label {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		font-weight: 700;
	}

	.grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(16rem, 1fr));
		gap: 2px var(--space-3);
	}

	.switches {
		display: flex;
		flex-direction: column;
	}

	.link {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		color: var(--color-text);
		font-size: var(--text-sm);
		font-weight: 600;
		text-decoration: none;
		white-space: nowrap;
	}

	.link:hover {
		text-decoration: underline;
	}

	.spacer {
		flex: 1;
	}

	@media (max-width: 36rem) {
		.pair,
		.secret {
			grid-template-columns: 1fr;
		}

		.rotate {
			align-items: flex-start;
			padding-top: 0;
		}
	}
</style>
