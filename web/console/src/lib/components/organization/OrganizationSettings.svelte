<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { createMutation, createQuery, useQueryClient } from '@tanstack/svelte-query';
	import {
		RiBuildingLine,
		RiCustomerService2Line,
		RiFileTextLine,
		RiGlobalLine,
		RiImageLine,
		RiMailLine,
		RiPhoneLine,
		RiShieldCheckLine,
		RiTimeLine
	} from 'svelte-remixicon';
	import { organizationApi, type Organization, type OrganizationSettings } from '$lib/api';
	import {
		Alert,
		Button,
		Code,
		FieldGrid,
		Input,
		Select,
		Tag,
		Thumb,
		notify
	} from '$lib/components/ui';
	import { ADMIN_DEPENDENCY } from '$lib/constants';
	import { keys, organizationOptions } from '$lib/query';
	import { formatDate, initials } from '$lib/utils/format';
	import SettingsSection from './SettingsSection.svelte';

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

	let saving = $state(false);

	/** Puts the form back to what is stored. */
	function discard() {
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

	/** What the sign-in pages show when it is filled in, so the summary can
	    say how much of it is still missing rather than the page looking
	    finished. */
	const details = $derived([
		{ label: 'Logo', filled: input.logo_url !== '' },
		{ label: 'Support email', filled: input.support_email !== '' },
		{ label: 'Support number', filled: input.support_phone !== '' },
		{ label: 'Terms', filled: input.terms_url !== '' },
		{ label: 'Privacy policy', filled: input.privacy_url !== '' }
	]);

	const filled = $derived(details.filter((detail) => detail.filled).length);

	const missing = $derived(details.filter((detail) => !detail.filled));

	/** The zone's distance from UTC, as "GMT+06:00": the same on the server
	    and in the browser, which the time of day would not be. */
	const offset = $derived.by(() => {
		try {
			return (
				new Intl.DateTimeFormat('en-US', { timeZone: input.timezone, timeZoneName: 'longOffset' })
					.formatToParts(new Date())
					.find((part) => part.type === 'timeZoneName')?.value ?? ''
			);
		} catch {
			return '';
		}
	});

	const save = createMutation(() => ({
		mutationFn: () => organizationApi.update(input),
		onSuccess: async (result) => {
			queryClient.setQueryData(keys.organization.settings, result);
			form = settingsOf(result.organization);
			notify.success('Settings saved', 'The sign-in pages and the discovery document follow.');
			// The log gains an entry for the change, and the header shows the
			// name and logo it was loaded with.
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.admin.overview }),
				invalidate(ADMIN_DEPENDENCY)
			]);
		},
		onError: (err: unknown) => {
			notify.error(err, 'Could not save these settings');
		},
		onSettled: () => {
			saving = false;
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (!editable || saving || !dirty || !complete) return;

		saving = true;
		save.mutate();
	}
</script>

<form onsubmit={submit}>
	<!-- The organisation as it stands, and as it is being typed: the card
	     follows the fields, so a logo or a name is seen before it is saved,
	     and it says how much of what the sign-in pages can show is there. -->
	<section class="summary" aria-label="Organization">
		<div class="identity">
			<Thumb src={logo} text={monogram} size="md" shape="circle" tone="accent" />

			<div class="names">
				<strong title={input.name}>{input.name || 'Unnamed organization'}</strong>
				<span class="line">
					<Code tone="quiet" truncate>{input.slug || organization.slug}</Code>
					<span aria-hidden="true">·</span>
					<span title="When this organization's record was created">
						Since {formatDate(organization.created_at)}
					</span>
				</span>
			</div>
		</div>

		<div class="progress">
			<div class="count">
				<span>Public details</span>
				<strong>{filled} of {details.length}</strong>
			</div>
			<div
				class="meter"
				role="meter"
				aria-label="Public details filled in"
				aria-valuemin={0}
				aria-valuemax={details.length}
				aria-valuenow={filled}
			>
				<span style:width="{(filled / details.length) * 100}%"></span>
			</div>
			{#if missing.length > 0}
				<div class="missing">
					{#each missing as detail (detail.label)}<Tag small>{detail.label}</Tag>{/each}
				</div>
				<small>The sign-in pages leave these out until they are filled in.</small>
			{:else}
				<small>Everything the sign-in pages can show is filled in.</small>
			{/if}
		</div>
	</section>

	{#if !editable}
		<Alert tone="info">Your roles let you see these settings but not change them.</Alert>
	{/if}

	<fieldset disabled={!editable}>
		<SettingsSection
			title="Identity"
			description="What the organization is called, and the mark the sign-in pages and this panel's header show."
			icon={RiBuildingLine}
		>
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
					hint="Lower case letters, numbers and dashes."
				/>

				<div class="full">
					<Input
						label="Logo URL"
						icon={RiImageLine}
						bind:value={form.logo_url}
						maxlength={512}
						placeholder="https://example.com/logo.svg"
						hint="A full address, previewed above. An application with a logo of its own keeps it."
					/>
				</div>
			</FieldGrid>
		</SettingsSection>

		<SettingsSection
			title="Region"
			description="The zone dates are written in on a user's account pages, such as when the account was made."
			icon={RiGlobalLine}
		>
			{#snippet meta()}
				{#if offset}<Tag small>{offset}</Tag>{/if}
			{/snippet}

			<Select
				label="Timezone"
				icon={RiTimeLine}
				bind:value={form.timezone}
				options={timezones}
				required
				disabled={!editable}
			/>
		</SettingsSection>

		<SettingsSection
			title="Support"
			description="Where someone who cannot get in can ask for help."
			icon={RiCustomerService2Line}
		>
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
		</SettingsSection>

		<SettingsSection
			title="Agreements"
			description="What users agree to. An application with links of its own shows those instead."
			icon={RiFileTextLine}
		>
			{#snippet meta()}
				<Tag small>Sign-in pages</Tag>
				<Tag small>Discovery document</Tag>
			{/snippet}

			<Input
				label="Terms of service"
				icon={RiFileTextLine}
				bind:value={form.terms_url}
				maxlength={512}
				placeholder="https://example.com/terms"
				hint="Published as op_tos_uri, and linked under the sign-in card."
			/>

			<Input
				label="Privacy policy"
				icon={RiShieldCheckLine}
				bind:value={form.privacy_url}
				maxlength={512}
				placeholder="https://example.com/privacy"
				hint="Published as op_policy_uri, and agreed to on registration."
			/>
		</SettingsSection>
	</fieldset>

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
	{/if}
</form>

<style>
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-5);
	}

	fieldset {
		display: flex;
		flex-direction: column;
		min-width: 0;
		gap: var(--space-5);
		margin: 0;
		padding: 0;
		border: none;
	}

	/* The organisation at a glance: who it is on the left, how much of it
	   the sign-in pages can show on the right. Measured on its own width, so
	   it stacks wherever it is narrow. */
	.summary {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-4);
		padding: var(--space-4);
		border: 1px solid var(--color-secondary-alt);
		border-radius: var(--radius-sm);
		background: var(--color-surface);
		box-shadow: var(--shadow-panel);
	}

	.identity {
		display: flex;
		flex: 1 1 18rem;
		align-items: center;
		gap: var(--space-3);
		min-width: 0;
	}

	.names {
		display: flex;
		flex-direction: column;
		gap: 4px;
		min-width: 0;
	}

	.names strong {
		overflow: hidden;
		font-size: var(--text-xl);
		font-weight: 600;
		line-height: 1.25;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.line {
		display: flex;
		align-items: center;
		gap: var(--space-1);
		min-width: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		white-space: nowrap;
	}

	.progress {
		display: flex;
		flex: 0 1 20rem;
		flex-direction: column;
		gap: var(--space-1);
		min-width: 0;
	}

	.count {
		display: flex;
		justify-content: space-between;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.count strong {
		color: var(--color-text);
		font-weight: 600;
	}

	.meter {
		height: 6px;
		overflow: hidden;
		border-radius: var(--radius-pill);
		background: var(--color-secondary);
	}

	.meter span {
		display: block;
		height: 100%;
		border-radius: inherit;
		background: var(--color-brand);
		transition: width var(--speed);
	}

	.missing {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		margin-top: 2px;
	}

	.progress small {
		color: var(--color-text-hint);
		font-size: var(--text-xs);
	}

	/* Opaque, and the width of the column: stuck to the foot of the window it
	   passes over the sections, which it may not show through. */
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

	@media (prefers-reduced-motion: reduce) {
		.meter span {
			transition: none;
		}
	}
</style>
