<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import {
		RiBuildingLine,
		RiCustomerService2Line,
		RiFileTextLine,
		RiImageLine,
		RiMailLine,
		RiPhoneLine,
		RiShieldCheckLine,
		RiTimeLine
	} from 'svelte-remixicon';
	import {
		ApiError,
		organizationApi,
		type Organization,
		type OrganizationSettings
	} from '$lib/api';
	import {
		Alert,
		Button,
		FieldGrid,
		Input,
		List,
		ListItem,
		Panel,
		Select,
		Tag,
		Thumb
	} from '$lib/components/ui';
	import { ADMIN_DEPENDENCY } from '$lib/constants';
	import { keys, organizationOptions } from '$lib/query';
	import { formatDate, initials } from '$lib/utils/format';

	type Props = {
		/** What the server rendered, to seed the query with. */
		initial: Organization;
		/** Whether the administrator may change any of this. */
		editable: boolean;
	};

	let { initial, editable }: Props = $props();

	const queryClient = useQueryClient();

	const settings = createQuery(() => organizationOptions({ organization: initial }));

	const organization = $derived(settings.data.organization);

	/** The stored record, without the bookkeeping the form cannot change.
	    Taking it as "the rest of the record" means a setting added to the API
	    reaches the form without this line naming it. */
	function settingsOf({ created_at, ...settings }: Organization): OrganizationSettings {
		void created_at;
		return settings;
	}

	// Seeded once, when the panel is first drawn: a refetch in the background
	// leaves what is being typed alone, and a save fills it in again.
	// svelte-ignore state_referenced_locally
	let form = $state(settingsOf(organization));

	let error = $state('');
	let saved = $state(false);
	let saving = $state(false);

	/** Puts the form back to what is stored, and takes down whatever the last
	    attempt said: neither the error nor the confirmation is about the form
	    once it holds the stored values again. */
	function discard() {
		error = '';
		saved = false;
		form = settingsOf(organization);
	}

	/** What the form holds, as the API takes it. A short name is compared
	    rather than read, so it is sent lower case. */
	const input = $derived<OrganizationSettings>({
		name: form.name.trim(),
		slug: form.slug.trim().toLowerCase(),
		logo_url: form.logo_url.trim(),
		support_email: form.support_email.trim(),
		support_phone: form.support_phone.trim(),
		terms_url: form.terms_url.trim(),
		privacy_url: form.privacy_url.trim(),
		timezone: form.timezone
	});

	/** Every zone this browser knows, which is the IANA database. UTC is not
	    always among them — engines list places, and UTC is not one — and a
	    stored zone this browser calls by another name is kept, so the select
	    never shows a saved value as nothing chosen. */
	const timezones = $derived.by(() => {
		const zones = ['UTC', organization.timezone, ...Intl.supportedValuesOf('timeZone')];
		return zones
			.filter((zone, index) => zone !== '' && zones.indexOf(zone) === index)
			.map((zone) => ({ value: zone, label: zone.replaceAll('_', ' ') }));
	});

	const dirty = $derived(
		Object.entries(input).some(
			([key, value]) => value !== organization[key as keyof OrganizationSettings]
		)
	);

	/** What has to be filled in before there is anything to save. */
	const complete = $derived(input.name !== '' && input.slug !== '');

	/** Two letters standing in for a logo there is none of, or none that
	    loads: the initials of a name in several words, or the start of a name
	    in one. */
	const monogram = $derived(initials(input.name));

	/** Shown as the logo while the address looks like one; Thumb falls back
	    to the monogram if it does not load. */
	const logo = $derived(/^https?:\/\/\S+$/.test(input.logo_url) ? input.logo_url : '');

	/** What the sign-in pages have nothing to show for, so the page says what
	    is still missing rather than looking finished. */
	const missing = $derived(
		[
			['Logo', input.logo_url],
			['Support email', input.support_email],
			['Support number', input.support_phone],
			['Terms', input.terms_url],
			['Privacy policy', input.privacy_url]
		]
			.filter(([, value]) => value === '')
			.map(([label]) => label)
	);

	const save = createMutation(() => ({
		mutationFn: () => organizationApi.update(input),
		onSuccess: async (result) => {
			queryClient.setQueryData(keys.organization.settings, result);
			form = settingsOf(result.organization);
			saved = true;
			// The log gains an entry for the change, and the header shows the
			// name and logo it was loaded with.
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.admin.overview }),
				invalidate(ADMIN_DEPENDENCY)
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not save these settings';
		},
		onSettled: () => {
			saving = false;
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (!editable || saving || !dirty || !complete) return;

		error = '';
		saved = false;
		saving = true;
		save.mutate();
	}
</script>

<form onsubmit={submit}>
	<!-- The organisation as it stands, and as it is being typed: the card
	     follows the fields, so a logo or a name is seen before it is saved. -->
	<List bordered label="Organization">
		<ListItem
			title={input.name || 'Unnamed organisation'}
			description={input.slug || organization.slug}
		>
			{#snippet lead()}
				<Thumb src={logo} text={monogram} size="md" />
			{/snippet}

			{#snippet end()}
				<Tag small title="When this organisation's record was created">
					Since {formatDate(organization.created_at)}
				</Tag>
			{/snippet}
		</ListItem>
	</List>

	{#if error}
		<Alert>{error}</Alert>
	{:else if saved && !dirty}
		<Alert tone="success"
			>Settings saved. The sign-in pages and the discovery document follow.</Alert
		>
	{/if}

	<fieldset disabled={!editable}>
		<Panel title="Organization" icon={RiBuildingLine}>
			<FieldGrid spacing="comfortable">
				<Input
					label="Name"
					bind:value={form.name}
					required
					maxlength={100}
					hint="Shown on the sign-in pages."
				/>

				<Input
					label="Identifier"
					bind:value={form.slug}
					required
					maxlength={64}
					hint="The slug: lower case letters, numbers and dashes."
				/>

				<div class="full">
					<Input
						label="Logo URL"
						icon={RiImageLine}
						bind:value={form.logo_url}
						maxlength={512}
						placeholder="https://example.com/logo.svg"
						hint="A full address, shown above. An application with a logo of its own keeps it."
					/>
				</div>

				<div class="full">
					<Select
						label="Timezone"
						icon={RiTimeLine}
						bind:value={form.timezone}
						options={timezones}
						required
						disabled={!editable}
						hint="The dates on a user's account pages, such as when the account was made."
					/>
				</div>
			</FieldGrid>
		</Panel>

		<Panel title="Contact" icon={RiCustomerService2Line}>
			{#snippet meta()}
				<Tag small>Shown when signing in</Tag>
			{/snippet}

			<FieldGrid spacing="comfortable">
				<Input
					label="Support email"
					icon={RiMailLine}
					bind:value={form.support_email}
					type="email"
					maxlength={255}
					placeholder="support@example.com"
					hint="Where someone who cannot get in writes."
				/>

				<Input
					label="Support number"
					icon={RiPhoneLine}
					bind:value={form.support_phone}
					maxlength={32}
					placeholder="+996 555 123456"
					hint="In the form you want it dialled."
				/>
			</FieldGrid>
		</Panel>

		<Panel title="Agreements" icon={RiFileTextLine}>
			{#snippet meta()}
				<Tag small>In the discovery document</Tag>
			{/snippet}

			<FieldGrid spacing="comfortable">
				<div class="full">
					<Input
						label="Terms of service"
						icon={RiFileTextLine}
						bind:value={form.terms_url}
						maxlength={512}
						placeholder="https://example.com/terms"
						hint="Published as op_tos_uri, and linked under the sign-in card."
					/>
				</div>

				<div class="full">
					<Input
						label="Privacy policy"
						icon={RiShieldCheckLine}
						bind:value={form.privacy_url}
						maxlength={512}
						placeholder="https://example.com/privacy"
						hint="Published as op_policy_uri, and agreed to on registration."
					/>
				</div>
			</FieldGrid>

			<p class="note">
				An application with links of its own shows those instead; these are what users see for every
				application that has none.
			</p>
		</Panel>
	</fieldset>

	{#if missing.length > 0}
		<p class="missing">
			Not filled in yet: {missing.join(', ')}. The sign-in pages leave out what is empty.
		</p>
	{/if}

	{#if editable && dirty}
		<!-- The bar appears only once something has been typed, and then stays
		     at the foot of the window, so the buttons are within reach of
		     whichever field is being edited. -->
		<div class="actions">
			<span class="pending">Unsaved changes</span>

			<Button variant="subtle" onclick={discard} disabled={saving}>Discard</Button>
			<Button type="submit" loading={saving} disabled={saving || !complete}>
				{saving ? 'Saving…' : 'Save changes'}
			</Button>
		</div>
	{:else if !editable}
		<p class="note">Your roles let you see these settings but not change them.</p>
	{/if}
</form>

<style>
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}

	fieldset {
		display: flex;
		flex-direction: column;
		min-width: 0;
		gap: var(--space-4);
		margin: 0;
		padding: 0;
		border: none;
	}

	/* Two fields to a row, each as wide as the other, and a `wide` one across
	   both: an address or a URL is read in full rather than in half. */

	/* Opaque, and the width of the column: stuck to the foot of the window it
	   passes over the panels, which it may not show through. */
	.actions {
		position: sticky;
		bottom: 0;
		z-index: 1;
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: var(--space-2);
		margin-inline: calc(var(--space-4) * -1);
		padding: var(--space-3) var(--space-4);
		border-top: 1px solid var(--color-secondary-alt);
		background: var(--color-surface);
		box-shadow: var(--shadow-panel);
	}

	.pending {
		margin-right: auto;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.note,
	.missing {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		line-height: 1.5;
	}
</style>
