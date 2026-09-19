<script lang="ts">
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import {
		RiCheckLine,
		RiDeleteBinLine,
		RiDownload2Line,
		RiFlaskLine,
		RiGroupLine,
		RiLinksLine,
		RiRefreshLine,
		RiServerLine,
		RiSettings3Line
	} from 'svelte-remixicon';
	import {
		ssoApi,
		type Role,
		type SSOConnection,
		type SSOConnectionInput,
		type SSOProtocol,
		type SSORoleMapping,
		type SSOTestResult
	} from '$lib/api';
	import {
		Alert,
		Button,
		Drawer,
		FormSection,
		Icon,
		Input,
		PasswordInput,
		Select,
		SwitchField,
		Tabs,
		Textarea
	} from '$lib/components/ui';
	import { messageOf, useTranslator } from '$lib/i18n';
	import { keys } from '$lib/query';
	import DomainsField from './DomainsField.svelte';
	import RoleMappings from './RoleMappings.svelte';

	type Props = {
		open: boolean;
		/** The connection being changed, or null to add one. */
		connection?: SSOConnection | null;
		roles: Role[];
		/** Whether the administrator may read roles at all — without it,
		    `roles` is empty for a reason other than there being none. */
		canReadRoles: boolean;
		canWrite: boolean;
	};

	let {
		open = $bindable(false),
		connection = null,
		roles,
		canReadRoles,
		canWrite
	}: Props = $props();

	const queryClient = useQueryClient();
	const t = useTranslator();

	/** The connection as it stands on the server: the one opened, or the one
	    just made — which the drawer stays open on, since what comes next is
	    giving its provider the addresses on the Service provider tab. */
	let current = $state<SSOConnection | null>(null);

	let tab = $state('connection');
	let protocol = $state<SSOProtocol>('oidc');
	let name = $state('');
	let slug = $state('');
	let enabled = $state(false);
	let issuer = $state('');
	let clientId = $state('');
	let clientSecret = $state('');
	let scopes = $state('');
	let metadataUrl = $state('');
	let metadata = $state('');
	let nameIdFormat = $state<SSOConnection['name_id_format']>('email');
	let signRequests = $state(false);
	let domains = $state<string[]>([]);
	let enforceDomains = $state(false);
	let showOnLogin = $state(false);
	let matching = $state<SSOConnection['matching']>('link');
	let createUsers = $state(true);
	let syncProfile = $state(true);
	let emailAttribute = $state('');
	let firstNameAttribute = $state('');
	let lastNameAttribute = $state('');
	let groupsAttribute = $state('');
	let mappings = $state<SSORoleMapping[]>([]);
	let syncRoles = $state(false);

	let error = $state('');
	let saving = $state(false);
	let confirmingDelete = $state(false);
	let tested = $state<SSOTestResult | null>(null);
	let testError = $state('');
	let testing = $state(false);

	$effect(() => {
		if (!open) return;

		current = connection;
		fill(connection);
		tab = 'connection';
		error = '';
		confirmingDelete = false;
		tested = null;
		testError = '';
	});

	/** The form, from a connection — or as a new one starts. */
	function fill(from: SSOConnection | null) {
		protocol = from?.protocol ?? 'oidc';
		name = from?.name ?? '';
		slug = from?.slug ?? '';
		enabled = from?.enabled ?? false;
		issuer = from?.issuer ?? '';
		clientId = from?.client_id ?? '';
		clientSecret = '';
		scopes = (from?.scopes ?? []).join(' ');
		metadataUrl = from?.metadata_url ?? '';
		metadata = from?.metadata_url ? '' : (from?.metadata ?? '');
		nameIdFormat = from?.name_id_format ?? 'email';
		signRequests = from?.sign_requests ?? false;
		domains = [...(from?.domains ?? [])];
		enforceDomains = from?.enforce_domains ?? false;
		// A new connection starts with its button, which is what lets it be
		// made before its domains are known.
		showOnLogin = from?.show_on_login ?? true;
		matching = from?.matching ?? 'link';
		createUsers = from?.create_users ?? true;
		syncProfile = from?.sync_profile ?? true;
		emailAttribute = from?.email_attribute ?? '';
		firstNameAttribute = from?.first_name_attribute ?? '';
		lastNameAttribute = from?.last_name_attribute ?? '';
		groupsAttribute = from?.groups_attribute ?? '';
		mappings = (from?.role_mappings ?? []).map((mapping) => ({ ...mapping }));
		syncRoles = from?.sync_roles ?? false;
	}

	const editing = $derived(current !== null);
	const readOnly = $derived(!canWrite);

	/** SSO is required of the addresses at a connection's domains, so without
	    any there is nobody to require it of. */
	const enforced = $derived(domains.length > 0 && enforceDomains);

	/** What still stands between the form and saving, and the tab it is on —
	    said beside the button, which takes you there, rather than leaving the
	    button greyed out for no reason anybody can see. */
	const missing = $derived.by((): { text: string; tab: string } | null => {
		if (name.trim() === '') return { text: t('sso.need_name'), tab: 'connection' };
		if (protocol === 'oidc' && (issuer.trim() === '' || clientId.trim() === ''))
			return { text: t('sso.need_oidc'), tab: 'connection' };
		if (protocol === 'saml' && metadataUrl.trim() === '' && metadata.trim() === '')
			return { text: t('sso.need_metadata'), tab: 'connection' };
		if (domains.length === 0 && !showOnLogin) return { text: t('sso.need_way_in'), tab: 'signin' };
		return null;
	});

	function input(): SSOConnectionInput {
		const body: SSOConnectionInput = {
			name: name.trim(),
			enabled,
			domains,
			enforce_domains: enforced,
			show_on_login: showOnLogin,
			matching,
			create_users: createUsers,
			sync_profile: syncProfile,
			email_attribute: emailAttribute.trim(),
			first_name_attribute: firstNameAttribute.trim(),
			last_name_attribute: lastNameAttribute.trim(),
			groups_attribute: groupsAttribute.trim(),
			role_mappings: mappings.filter((mapping) => mapping.group.trim() !== ''),
			sync_roles: syncRoles
		};

		if (!editing) {
			body.protocol = protocol;
			if (slug.trim()) body.slug = slug.trim();
		}

		if (protocol === 'oidc') {
			body.issuer = issuer.trim();
			body.client_id = clientId.trim();
			body.scopes = scopes.split(/\s+/).filter(Boolean);
			if (clientSecret.trim()) body.client_secret = clientSecret.trim();
		} else {
			body.metadata_url = metadataUrl.trim();
			// Metadata from an address is fetched by the server; pasted
			// metadata is sent as it is.
			if (!metadataUrl.trim()) body.metadata = metadata.trim();
			body.name_id_format = nameIdFormat;
			body.sign_requests = signRequests;
		}

		return body;
	}

	const save = createMutation(() => ({
		mutationFn: () => (current ? ssoApi.update(current.id, input()) : ssoApi.create(input())),
		onSuccess: async ({ connection: saved }) => {
			const created = !editing;
			await queryClient.invalidateQueries({ queryKey: keys.sso.all });

			if (created) {
				current = saved;
				fill(saved);
				tab = 'provider';
				return;
			}

			open = false;
		},
		onError: (err: unknown) => {
			error = messageOf(err, t, t('sso.save_failed'));
		},
		onSettled: () => {
			saving = false;
		}
	}));

	const remove = createMutation(() => ({
		mutationFn: () => ssoApi.remove(current?.id ?? ''),
		onSuccess: async () => {
			open = false;
			await queryClient.invalidateQueries({ queryKey: keys.sso.all });
		},
		onError: (err: unknown) => {
			error = messageOf(err, t, t('sso.remove_failed'));
		}
	}));

	const refresh = createMutation(() => ({
		mutationFn: () => ssoApi.refreshMetadata(current?.id ?? ''),
		onSuccess: async ({ connection: saved }) => {
			current = saved;
			await queryClient.invalidateQueries({ queryKey: keys.sso.all });
		},
		onError: (err: unknown) => {
			error = messageOf(err, t, t('sso.save_failed'));
		}
	}));

	async function test() {
		testing = true;
		testError = '';
		tested = null;

		try {
			tested = await ssoApi.test({
				protocol,
				issuer: issuer.trim(),
				scopes: scopes.split(/\s+/).filter(Boolean),
				metadata_url: metadataUrl.trim(),
				metadata: metadataUrl.trim() ? undefined : metadata.trim()
			});
		} catch (err) {
			testError = messageOf(err, t);
		} finally {
			testing = false;
		}
	}

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (missing || saving || readOnly) return;

		error = '';
		saving = true;
		save.mutate();
	}

	const tabs = $derived([
		{ value: 'connection', label: t('sso.tab_connection'), icon: RiSettings3Line },
		{ value: 'provider', label: t('sso.tab_provider'), icon: RiServerLine },
		{ value: 'signin', label: t('sso.tab_signin'), icon: RiLinksLine },
		{ value: 'users', label: t('sso.tab_users'), icon: RiGroupLine }
	]);

	const protocols = $derived([
		{ value: 'oidc', label: t('sso.protocol_oidc'), description: t('sso.protocol_oidc_hint') },
		{ value: 'saml', label: t('sso.protocol_saml'), description: t('sso.protocol_saml_hint') }
	]);

	const nameIdFormats = $derived([
		{ value: 'email', label: t('sso.name_id_email') },
		{ value: 'persistent', label: t('sso.name_id_persistent') },
		{ value: 'unspecified', label: t('sso.name_id_unspecified') }
	]);

	const matchings = $derived([
		{ value: 'link', label: t('sso.matching_link') },
		{ value: 'deny', label: t('sso.matching_deny') }
	]);

	/** A provider certificate that runs out within a month is worth saying:
	    it is the usual way a working connection stops working. */
	const expiring = $derived.by(() => {
		const expires = current?.identity_provider?.certificate_expires;
		if (!expires) return null;

		const when = new Date(expires);
		return when.getTime() - Date.now() < 30 * 24 * 3600 * 1000 ? when : null;
	});

	function download(filename: string, text: string, type: string) {
		const url = URL.createObjectURL(new Blob([text], { type }));
		const link = document.createElement('a');
		link.href = url;
		link.download = filename;
		link.click();
		URL.revokeObjectURL(url);
	}
</script>

<Drawer
	bind:open
	title={current ? current.name : t('sso.new_title')}
	description={current ? t('sso.edit_description') : t('sso.new_description')}
	meta={current?.slug}
	width="52rem"
	onsubmit={submit}
>
	{#if error}
		<div class="message"><Alert>{error}</Alert></div>
	{/if}

	<Tabs {tabs} bind:value={tab} label={t('sso.title')}>
		{#snippet panel(value)}
			<div class="panel">
				{#if value === 'connection'}
					<FormSection title={t('sso.section_connection')}>
						{#if !editing}
							<Select label={t('sso.protocol')} bind:value={protocol} options={protocols} />
						{/if}

						<div class="pair">
							<Input
								label={t('sso.name')}
								bind:value={name}
								hint={t('sso.name_hint', { name: name.trim() || 'Acme' })}
								required
								{readOnly}
							/>
							<Input
								label={t('sso.slug')}
								bind:value={slug}
								hint={t('sso.slug_hint')}
								placeholder="acme"
								readOnly={editing || readOnly}
							/>
						</div>

						<SwitchField
							label={t('sso.enabled')}
							description={t('sso.enabled_hint')}
							bind:checked={enabled}
							disabled={readOnly}
						/>
					</FormSection>

					{#if protocol === 'oidc'}
						<FormSection title={t('sso.section_oidc')} description={t('sso.oidc_description')}>
							<Input
								label={t('sso.issuer')}
								bind:value={issuer}
								hint={t('sso.issuer_hint')}
								placeholder="https://acme.okta.com"
								required
								{readOnly}
							/>
							<div class="pair">
								<Input label={t('sso.client_id')} bind:value={clientId} required {readOnly} />
								<div class="secret">
									<PasswordInput
										label={t('sso.client_secret')}
										bind:value={clientSecret}
										autocomplete="new-password"
										disabled={readOnly}
									/>
									{#if current?.has_client_secret}
										<p class="note small">{t('sso.client_secret_stored')}</p>
									{/if}
								</div>
							</div>
							<Input
								label={t('sso.scopes')}
								bind:value={scopes}
								hint={t('sso.scopes_hint')}
								placeholder="email profile groups"
								{readOnly}
							/>
						</FormSection>
					{:else}
						<FormSection title={t('sso.section_saml')} description={t('sso.saml_description')}>
							<Input
								label={t('sso.metadata_url')}
								bind:value={metadataUrl}
								hint={t('sso.metadata_url_hint')}
								placeholder="https://login.microsoftonline.com/…/federationmetadata.xml"
								{readOnly}
							/>
							{#if !metadataUrl.trim()}
								<Textarea
									label={t('sso.metadata')}
									bind:value={metadata}
									rows={5}
									hint={t('sso.metadata_hint')}
								/>
							{/if}
							<div class="pair">
								<Select
									label={t('sso.name_id')}
									bind:value={nameIdFormat}
									options={nameIdFormats}
									{readOnly}
								/>
							</div>
							<SwitchField
								label={t('sso.sign_requests')}
								description={t('sso.sign_requests_hint')}
								bind:checked={signRequests}
								disabled={readOnly}
							/>
						</FormSection>
					{/if}

					{#if canWrite}
						<div class="test">
							<Button variant="subtle" onclick={test} loading={testing} disabled={testing}>
								<Icon icon={RiFlaskLine} />
								{t('sso.test')}
							</Button>

							{#if testError}
								<Alert>{testError}</Alert>
							{:else if tested?.authorization_endpoint}
								<Alert tone="success">
									{t('sso.test_ok_oidc', { url: tested.authorization_endpoint })}
								</Alert>
								{#if tested.unsupported_scopes?.length}
									<Alert tone="warning">
										{t('sso.test_unsupported_scopes', {
											scopes: tested.unsupported_scopes.join(' ')
										})}
									</Alert>
								{/if}
							{:else if tested?.identity_provider}
								<Alert tone="success">
									{t('sso.test_ok_saml', {
										entity: tested.identity_provider.entity_id,
										url: tested.identity_provider.sso_url
									})}
								</Alert>
							{/if}
						</div>
					{/if}
				{:else if value === 'provider'}
					{#if !current}
						<p class="note">{t('sso.save_first')}</p>
					{:else if current.protocol === 'oidc'}
						<FormSection title={t('sso.section_sp')} description={t('sso.sp_oidc_description')}>
							<Input
								label={t('sso.callback_url')}
								value={current.service_provider.callback_url ?? ''}
								readOnly
								copyable
							/>
						</FormSection>
					{:else}
						<FormSection title={t('sso.section_sp')} description={t('sso.sp_saml_description')}>
							<Input
								label={t('sso.acs_url')}
								value={current.service_provider.acs_url ?? ''}
								readOnly
								copyable
							/>
							<Input
								label={t('sso.entity_id')}
								value={current.service_provider.entity_id ?? ''}
								readOnly
								copyable
							/>
							<Input
								label={t('sso.metadata_link')}
								value={current.service_provider.metadata_url ?? ''}
								readOnly
								copyable
							/>

							<div class="files">
								<Button
									variant="subtle"
									size="sm"
									onclick={() =>
										download(
											`${current!.slug}-certificate.pem`,
											current!.service_provider.certificate ?? '',
											'application/x-pem-file'
										)}
								>
									<Icon icon={RiDownload2Line} />
									{t('sso.certificate_download')}
								</Button>
							</div>
						</FormSection>

						{#if current.identity_provider}
							<FormSection title={t('sso.section_idp')}>
								<dl class="facts">
									<dt>{t('sso.idp_entity')}</dt>
									<dd><code>{current.identity_provider.entity_id}</code></dd>
									<dt>{t('sso.idp_sso')}</dt>
									<dd><code>{current.identity_provider.sso_url}</code></dd>
									{#if current.identity_provider.certificate_expires}
										<dt>{t('sso.idp_certificate')}</dt>
										<dd>
											{new Date(current.identity_provider.certificate_expires).toLocaleDateString()}
										</dd>
									{/if}
								</dl>

								{#if expiring}
									<Alert tone="warning"
										>{t('sso.idp_expiring', { when: expiring.toLocaleDateString() })}</Alert
									>
								{/if}

								{#if canWrite && current.metadata_url}
									<div>
										<Button
											variant="subtle"
											size="sm"
											onclick={() => refresh.mutate()}
											loading={refresh.isPending}
										>
											<Icon icon={RiRefreshLine} />
											{t('sso.refresh_metadata')}
										</Button>
									</div>
								{/if}
							</FormSection>
						{/if}
					{/if}
				{:else if value === 'signin'}
					<FormSection title={t('sso.section_domains')}>
						<DomainsField bind:domains {readOnly} />
					</FormSection>

					<FormSection title={t('sso.section_signin')}>
						<SwitchField
							label={t('sso.enforce')}
							description={domains.length > 0
								? t('sso.enforce_hint')
								: t('sso.enforce_needs_domain')}
							bind:checked={() => enforced, (value) => (enforceDomains = value)}
							disabled={readOnly || domains.length === 0}
						/>
						<SwitchField
							label={t('sso.show_on_login')}
							description={t('sso.show_on_login_hint', { name: name || '…' })}
							bind:checked={showOnLogin}
							disabled={readOnly}
						/>
					</FormSection>
				{:else}
					<FormSection title={t('sso.section_provisioning')}>
						<Select
							label={t('sso.matching')}
							bind:value={matching}
							options={matchings}
							{readOnly}
						/>
						<SwitchField
							label={t('sso.create_users')}
							description={t('sso.create_users_hint')}
							bind:checked={createUsers}
							disabled={readOnly}
						/>
						<SwitchField
							label={t('sso.sync_profile')}
							description={t('sso.sync_profile_hint')}
							bind:checked={syncProfile}
							disabled={readOnly}
						/>
					</FormSection>

					<FormSection title={t('sso.section_attributes')} description={t('sso.attributes_hint')}>
						<div class="pair">
							<Input
								label={t('sso.attr_email')}
								bind:value={emailAttribute}
								placeholder={protocol === 'oidc' ? 'email' : 'mail'}
								{readOnly}
							/>
							<Input
								label={t('sso.attr_groups')}
								bind:value={groupsAttribute}
								placeholder={protocol === 'oidc' ? 'groups' : 'memberOf'}
								{readOnly}
							/>
							<Input
								label={t('sso.attr_first')}
								bind:value={firstNameAttribute}
								placeholder={protocol === 'oidc' ? 'given_name' : 'givenName'}
								{readOnly}
							/>
							<Input
								label={t('sso.attr_last')}
								bind:value={lastNameAttribute}
								placeholder={protocol === 'oidc' ? 'family_name' : 'surname'}
								{readOnly}
							/>
						</div>
					</FormSection>

					<FormSection title={t('sso.section_mappings')} description={t('sso.mappings_hint')}>
						<RoleMappings bind:mappings {roles} canRead={canReadRoles} {readOnly} />
						<SwitchField
							label={t('sso.sync_roles')}
							description={t('sso.sync_roles_hint')}
							bind:checked={syncRoles}
							disabled={readOnly}
						/>
					</FormSection>
				{/if}
			</div>
		{/snippet}
	</Tabs>

	{#snippet footer()}
		{#if canWrite && current}
			{#if confirmingDelete}
				<div class="confirm">
					<span>{t('sso.remove_confirm', { count: current.users })}</span>
					<Button variant="subtle" size="sm" onclick={() => (confirmingDelete = false)}
						>{t('sso.keep')}</Button
					>
					<Button colorPalette="danger" size="sm" onclick={() => remove.mutate()}
						>{t('sso.remove')}</Button
					>
				</div>
			{:else}
				<Button
					colorPalette="danger"
					variant="subtle"
					size="sm"
					onclick={() => (confirmingDelete = true)}
				>
					<Icon icon={RiDeleteBinLine} />
					{t('sso.remove')}
				</Button>
			{/if}
		{/if}

		<div class="actions">
			{#if canWrite && missing}
				<button type="button" class="missing" onclick={() => (tab = missing.tab)}>
					{missing.text}
				</button>
			{/if}
			<Button variant="subtle" onclick={() => (open = false)}>{t('action.cancel')}</Button>
			{#if canWrite}
				<Button type="submit" loading={saving} disabled={missing !== null || saving}>
					<Icon icon={RiCheckLine} />
					{editing ? t('action.save_changes') : t('sso.create')}
				</Button>
			{/if}
		</div>
	{/snippet}
</Drawer>

<style>
	.message {
		margin-bottom: var(--space-4);
	}

	.panel {
		padding-top: var(--space-4);
	}

	.pair {
		display: grid;
		grid-template-columns: 1fr 1fr;
		align-items: start;
		gap: var(--space-3);
	}

	.test,
	.files {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: var(--space-2);
	}

	.note {
		margin: var(--space-4) 0;
		color: var(--color-text-hint);
	}

	.note.small {
		margin: var(--space-1) 0 0;
		font-size: var(--text-sm);
	}

	.facts {
		display: grid;
		grid-template-columns: max-content 1fr;
		gap: var(--space-1) var(--space-3);
		margin: 0;
		font-size: var(--text-sm);
	}

	.facts dt {
		color: var(--color-text-hint);
	}

	.facts dd {
		margin: 0;
		overflow-wrap: anywhere;
	}

	.facts code {
		font-family: var(--font-mono);
	}

	.confirm {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin-right: auto;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.missing {
		all: unset;
		cursor: pointer;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		text-align: right;
	}

	.missing:hover {
		color: var(--color-text);
		text-decoration: underline;
	}

	.missing:focus-visible {
		outline: 2px solid var(--color-info);
		outline-offset: 2px;
		border-radius: var(--radius-sm);
	}

	.actions {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		margin-left: auto;
	}

	@media (max-width: 40rem) {
		.pair {
			grid-template-columns: 1fr;
		}
	}
</style>
