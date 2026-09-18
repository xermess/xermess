<script lang="ts">
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { RiDeleteBinLine, RiExternalLinkLine } from 'svelte-remixicon';
	import {
		ApiError,
		socialApi,
		type SocialKind,
		type SocialProvider,
		type SocialProviderInput,
		type SocialSpec,
		type SocialTokenAuth
	} from '$lib/api';
	import {
		Alert,
		Button,
		Drawer,
		FormSection,
		Icon,
		Input,
		Select,
		SwitchField
	} from '$lib/components/ui';
	import { keys } from '$lib/query';
	import { kindOptions, tokenAuthOptions } from './providers';
	import SecretField from './SecretField.svelte';

	type Props = {
		open: boolean;
		/** The provider being changed, or null to register one. */
		provider?: SocialProvider | null;
		/** The kinds this server knows, which fill the form in. */
		kinds: SocialSpec[];
	};

	let { open = $bindable(false), provider = null, kinds }: Props = $props();

	const queryClient = useQueryClient();

	const editing = $derived(provider !== null);

	let kind = $state<SocialKind>('google');
	let slug = $state('');
	let name = $state('');
	let clientID = $state('');
	let clientSecret = $state('');
	let teamID = $state('');
	let keyID = $state('');
	let privateKey = $state('');
	let scopes = $state('');
	let tokenAuth = $state<SocialTokenAuth | ''>('');
	let authorizeURL = $state('');
	let tokenURL = $state('');
	let userInfoURL = $state('');
	let enabled = $state(true);
	let linkVerifiedEmails = $state(true);
	let allowRegistration = $state(true);

	let error = $state('');
	let saving = $state(false);
	let confirmingDelete = $state(false);

	const spec = $derived(kinds.find((one) => one.kind === kind));

	/** Apple signs its secret with a key instead of holding one. */
	const signs = $derived(spec?.signed_secret ?? false);
	/** The two custom kinds are the ones whose addresses have to be typed in. */
	const custom = $derived(spec?.custom ?? false);

	/** Everything that has to be filled in before there is anything to save.
	    A stored secret counts, since it cannot be shown to be retyped. */
	const ready = $derived(
		name.trim() !== '' &&
			clientID.trim() !== '' &&
			(signs
				? (privateKey.trim() !== '' || (provider?.has_private_key ?? false)) &&
					teamID.trim() !== '' &&
					keyID.trim() !== ''
				: clientSecret.trim() !== '' || (provider?.has_client_secret ?? false)) &&
			(!custom || (authorizeURL.trim() !== '' && tokenURL.trim() !== ''))
	);

	// The form is filled in each time the panel opens: with the provider
	// being changed, or with what the chosen kind already knows.
	$effect(() => {
		if (!open) return;

		error = '';
		confirmingDelete = false;
		clientSecret = '';
		privateKey = '';

		if (provider) {
			kind = provider.kind;
			slug = provider.slug;
			name = provider.name;
			clientID = provider.client_id;
			teamID = provider.team_id;
			keyID = provider.key_id;
			scopes = provider.scopes.join(' ');
			tokenAuth = provider.token_auth;
			authorizeURL = provider.authorize_url;
			tokenURL = provider.token_url;
			userInfoURL = provider.userinfo_url;
			enabled = provider.enabled;
			linkVerifiedEmails = provider.link_verified_emails;
			allowRegistration = provider.allow_registration;
			return;
		}

		fillFromKind(kinds[0]?.kind ?? 'google');
		enabled = true;
		linkVerifiedEmails = true;
		allowRegistration = true;
	});

	/** Choosing a kind fills in everything that kind already decides: its
	    name, the scopes it is usually asked for, and its addresses. */
	function fillFromKind(chosen: SocialKind) {
		kind = chosen;

		const known = kinds.find((one) => one.kind === chosen);
		slug = chosen;
		name = known?.label ?? chosen;
		scopes = (known?.scopes ?? []).join(' ');
		tokenAuth = '';
		authorizeURL = known?.custom ? '' : (known?.authorize_url ?? '');
		tokenURL = known?.custom ? '' : (known?.token_url ?? '');
		userInfoURL = known?.custom ? '' : (known?.userinfo_url ?? '');
		clientID = '';
		clientSecret = '';
		teamID = '';
		keyID = '';
		privateKey = '';
	}

	function input(): SocialProviderInput {
		const body: SocialProviderInput = {
			name: name.trim(),
			client_id: clientID.trim(),
			scopes: scopes.split(/\s+/).filter(Boolean),
			enabled,
			link_verified_emails: linkVerifiedEmails,
			allow_registration: allowRegistration
		};

		// A secret is only ever sent when one was typed: what is stored
		// cannot be read back, and an empty field means "leave it".
		if (clientSecret.trim() !== '') body.client_secret = clientSecret.trim();
		if (privateKey.trim() !== '') body.private_key = privateKey.trim();

		if (signs) {
			body.team_id = teamID.trim();
			body.key_id = keyID.trim();
		}

		if (custom) {
			body.authorize_url = authorizeURL.trim();
			body.token_url = tokenURL.trim();
			body.userinfo_url = userInfoURL.trim();
			body.token_auth = tokenAuth;
		}

		if (!editing) {
			body.kind = kind;
			body.slug = slug.trim().toLowerCase();
		}

		return body;
	}

	const save = createMutation(() => ({
		mutationFn: () =>
			provider ? socialApi.update(provider.id, input()) : socialApi.create(input()),
		onSuccess: async () => {
			open = false;
			await queryClient.invalidateQueries({ queryKey: keys.social.all });
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not save this provider';
		},
		onSettled: () => {
			saving = false;
		}
	}));

	const remove = createMutation(() => ({
		mutationFn: () => socialApi.remove(provider?.id ?? ''),
		onSuccess: async () => {
			open = false;
			await queryClient.invalidateQueries({ queryKey: keys.social.all });
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not remove this provider';
			saving = false;
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (!ready || saving) return;

		error = '';
		saving = true;
		save.mutate();
	}
</script>

<Drawer
	bind:open
	title={editing ? provider!.name : 'Add a provider'}
	description={editing
		? 'What this server sends to the provider, and what it may do here.'
		: 'Register this server with a provider, then paste what it gave you here.'}
	meta={editing ? provider!.slug : undefined}
	width="34rem"
	onsubmit={submit}
>
	{#if error}
		<div class="message"><Alert>{error}</Alert></div>
	{/if}

	<FormSection
		title="Provider"
		description={editing
			? 'The kind and the identifier are in the address registered with the provider, so they stay as they are.'
			: 'Choosing one fills in everything about it this server already knows.'}
	>
		{#if editing}
			<Input label="Kind" value={spec?.label ?? kind} readOnly />
		{:else}
			<Select
				label="Kind"
				value={kind}
				options={kindOptions(kinds)}
				onChange={(chosen) => fillFromKind(chosen)}
			/>
		{/if}

		<div class="pair">
			<Input label="Name" bind:value={name} required hint="What the button says." />
			<Input
				label="Identifier"
				bind:value={slug}
				readOnly={editing}
				required
				hint="In the callback address."
			/>
		</div>

		{#if spec?.docs}
			<p class="note">
				<!-- The provider's own console, on its own site. -->
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
				<a href={spec.docs} target="_blank" rel="noopener noreferrer">
					Register this server with {spec.label}
					<Icon icon={RiExternalLinkLine} size="0.8rem" />
				</a>
			</p>
		{/if}
	</FormSection>

	{#if editing}
		<FormSection
			title="Redirect URI"
			description="What the provider has to be told to send people back to. It has to match exactly."
		>
			<Input label="Callback URL" value={provider!.callback_url} readOnly copyable />
		</FormSection>
	{/if}

	<FormSection title="Credentials" description="What the provider gave you when you registered.">
		<Input label="Client ID" bind:value={clientID} required copyable={editing} />

		{#if signs}
			<div class="pair">
				<Input label="Team ID" bind:value={teamID} required copyable={editing} />
				<Input label="Key ID" bind:value={keyID} required copyable={editing} />
			</div>

			<SecretField
				label="Signing key"
				providerId={provider?.id}
				stored={provider?.has_private_key ?? false}
				bind:value={privateKey}
				multiline
				hint="The .p8 file from the developer account."
			/>
		{:else}
			<SecretField
				label="Client secret"
				providerId={provider?.id}
				stored={provider?.has_client_secret ?? false}
				bind:value={clientSecret}
			/>
		{/if}
	</FormSection>

	{#if custom}
		<FormSection
			title="Endpoints"
			description="Where this provider signs people in. Its discovery document lists all three."
		>
			<Input label="Authorization URL" bind:value={authorizeURL} required />
			<Input label="Token URL" bind:value={tokenURL} required />
			<Input
				label="Userinfo URL"
				bind:value={userInfoURL}
				hint="OpenID Connect providers may leave this out; the id_token carries the same claims."
			/>
			<Select label="Client authentication" bind:value={tokenAuth} options={tokenAuthOptions} />
		</FormSection>
	{/if}

	<FormSection title="Scopes" description="What to ask the provider for, separated by spaces.">
		<Input label="Scopes" bind:value={scopes} placeholder="openid email profile" />
	</FormSection>

	<FormSection title="What it may do here">
		<SwitchField
			label="Offer this provider"
			description="Show its button on the sign-in pages."
			bind:checked={enabled}
		/>

		<SwitchField
			label="Link to existing accounts"
			description="Sign someone in to the account that already has their address, when the provider says it has verified it. Off, they have to sign in and connect it themselves."
			bind:checked={linkVerifiedEmails}
		/>

		<SwitchField
			label="Create accounts"
			description="Let somebody with no account here make one by signing in with this provider."
			bind:checked={allowRegistration}
		/>
	</FormSection>

	{#snippet footer()}
		{#if editing}
			{#if confirmingDelete}
				<div class="confirm">
					<span>
						{provider!.identities > 0
							? `${provider!.identities} ${provider!.identities === 1 ? 'person signs' : 'people sign'} in with this. Their accounts stay; this way in goes.`
							: 'Nobody signs in with this yet.'}
					</span>
					<Button variant="subtle" size="sm" onclick={() => (confirmingDelete = false)}>
						Keep it
					</Button>
					<Button colorPalette="danger" size="sm" onclick={() => remove.mutate()}>Remove</Button>
				</div>
			{:else}
				<Button
					colorPalette="danger"
					variant="subtle"
					size="sm"
					onclick={() => (confirmingDelete = true)}
				>
					<Icon icon={RiDeleteBinLine} />
					Remove
				</Button>
			{/if}
		{/if}

		<div class="actions">
			<Button variant="subtle" onclick={() => (open = false)}>Cancel</Button>
			<Button type="submit" loading={saving} disabled={!ready || saving}>
				{editing ? 'Save changes' : 'Add provider'}
			</Button>
		</div>
	{/snippet}
</Drawer>

<style>
	.message {
		margin-bottom: var(--space-4);
	}

	.pair {
		display: grid;
		grid-template-columns: 1fr 1fr;
		align-items: start;
		gap: var(--space-3);
	}

	.note {
		margin: 0;
		font-size: var(--text-sm);
	}

	.note a {
		display: inline-flex;
		align-items: center;
		gap: var(--space-1);
		color: var(--color-text-hint);
	}

	.note a:hover {
		color: var(--color-text);
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

	.actions {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		margin-left: auto;
	}

	@media (max-width: 34rem) {
		.pair {
			grid-template-columns: 1fr;
		}
	}
</style>
