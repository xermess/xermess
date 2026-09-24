<script lang="ts">
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import {
		RiAtLine,
		RiCheckLine,
		RiLockLine,
		RiMailLine,
		RiMailSendLine,
		RiServerLine,
		RiUserLine
	} from 'svelte-remixicon';
	import { ApiError, mailApi, type MailEncryption, type MailResponse } from '$lib/api';
	import { Alert, Button, Input, Panel, Select, SwitchField, Tag } from '$lib/components/ui';
	import { keys } from '$lib/query';

	type Props = {
		settings: MailResponse;
	};

	let { settings }: Props = $props();

	const queryClient = useQueryClient();

	const stored = $derived(settings.mail);

	/** The stored record as the form holds it: without what the form cannot
	    change, and with the port as text, since that is what a field binds. */
	function formOf({ has_password, port, ...rest }: MailResponse['mail']) {
		void has_password;
		return { ...rest, port: String(port) };
	}

	// Seeded once, when the panel is first drawn: a refetch in the background
	// leaves what is being typed alone, and a save fills it in again.
	// svelte-ignore state_referenced_locally
	let form = $state(formOf(settings.mail));

	/** The password is never read back, so the field starts empty whatever is
	    stored: typing one replaces it, and leaving it alone keeps it. */
	let password = $state('');
	/** Asking for the stored password to go, for a server that takes none. */
	let removePassword = $state(false);

	let error = $state('');
	let saved = $state(false);
	let saving = $state(false);

	function discard() {
		error = '';
		saved = false;
		password = '';
		removePassword = false;
		form = formOf(stored);
	}

	/** What the form holds, as the API takes it. A host name and an address
	    are dialled and compared rather than read, so they go lower case.

	    Its shape is inferred rather than declared: every key here is one of
	    the stored record's, which is what lets `dirty` below compare the two
	    without naming a field. */
	const input = $derived({
		enabled: form.enabled,
		host: form.host.trim().toLowerCase(),
		port: Number(form.port),
		encryption: form.encryption,
		username: form.username.trim(),
		from_address: form.from_address.trim().toLowerCase(),
		from_name: form.from_name.trim()
	});

	/** The password as the API takes it: left out unless one was typed, or
	    an empty string to clear the one that is stored. */
	const passwordInput = $derived(
		password !== '' ? password : removePassword && stored.has_password ? '' : undefined
	);

	const dirty = $derived(
		passwordInput !== undefined ||
			(Object.keys(input) as (keyof typeof input)[]).some((key) => input[key] !== stored[key])
	);

	/** What has to be filled in before a message could go anywhere. Sending
	    switched off asks for nothing but a port: a half-filled form is work in
	    progress, which is what the server says too. */
	const complete = $derived(
		Number.isInteger(input.port) &&
			input.port >= 1 &&
			input.port <= 65535 &&
			(!input.enabled || (input.host !== '' && input.from_address !== ''))
	);

	const encryptions = $derived(
		settings.encryptions.map((value) => ({
			value,
			label: labelOf(value),
			description: describe(value)
		}))
	);

	function labelOf(encryption: MailEncryption): string {
		return { starttls: 'STARTTLS', tls: 'TLS', none: 'None' }[encryption];
	}

	function describe(encryption: MailEncryption): string {
		return {
			starttls: 'Connect in the clear and upgrade. What port 587 expects.',
			tls: 'Encrypted from the first byte. What port 465 expects.',
			none: 'Nothing encrypted. Only for a relay on this machine.'
		}[encryption];
	}

	const save = createMutation(() => ({
		mutationFn: () => mailApi.update({ ...input, password: passwordInput }),
		onSuccess: async (result) => {
			queryClient.setQueryData(keys.mail.settings, result);
			form = formOf(result.mail);
			password = '';
			removePassword = false;
			saved = true;
			// The log gains an entry for the change.
			await queryClient.invalidateQueries({ queryKey: keys.admin.overview });
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

		if (saving || !dirty || !complete) return;

		error = '';
		saved = false;
		saving = true;
		save.mutate();
	}

	// ---- Sending a test message -------------------------------------------

	let testTo = $state('');
	let testError = $state('');
	let testSent = $state('');

	const test = createMutation(() => ({
		mutationFn: () => mailApi.test({ ...input, password: passwordInput, to: testTo.trim() }),
		onSuccess: async (result) => {
			testSent = result.to;
			await queryClient.invalidateQueries({ queryKey: keys.admin.overview });
		},
		onError: (err: unknown) => {
			testError = err instanceof ApiError ? err.message : 'Could not send a test message';
		}
	}));

	function sendTest() {
		if (test.isPending || testTo.trim() === '') return;

		testError = '';
		testSent = '';
		test.mutate();
	}
</script>

<form onsubmit={submit}>
	{#if error}
		<Alert>{error}</Alert>
	{:else if saved && !dirty}
		<Alert tone="success">
			Settings saved. The next password reset, confirmation and one-time code goes through this
			server.
		</Alert>
	{/if}

	<Panel title="Sending" icon={RiMailSendLine}>
		{#snippet meta()}
			{#if form.enabled}
				<Tag tone="success" dot small>Sending</Tag>
			{:else}
				<Tag tone="warning" dot small>Logged, not sent</Tag>
			{/if}
		{/snippet}

		<SwitchField
			label="Send email through an SMTP server"
			description="Off, every message is written to the server's log instead — enough to follow a reset link while developing, and nothing anybody receives."
			bind:checked={form.enabled}
		/>
	</Panel>

	<Panel title="Mail server" icon={RiServerLine}>
		<div class="grid">
			<div class="wide">
				<Input
					label="Host"
					icon={RiServerLine}
					bind:value={form.host}
					maxlength={255}
					placeholder="smtp.example.com"
					hint="The host name on its own, with no scheme and no port."
				/>
			</div>

			<Input
				label="Port"
				type="number"
				bind:value={form.port}
				min={1}
				max={65535}
				hint="587 for STARTTLS, 465 for TLS, 25 for a local relay."
			/>

			<Select
				label="Encryption"
				bind:value={form.encryption}
				options={encryptions}
				hint="How the connection is protected."
			/>

			<Input
				label="Username"
				icon={RiUserLine}
				bind:value={form.username}
				maxlength={255}
				hint="Leave empty for a server that takes no credentials."
			/>

			<Input
				label="Password"
				icon={RiLockLine}
				bind:value={password}
				type="password"
				autocomplete="off"
				maxlength={255}
				disabled={removePassword}
				placeholder={stored.has_password ? 'Stored — type to replace it' : ''}
				hint="Sealed with the server's secret key, and never read back."
			/>

			{#if stored.has_password}
				<div class="wide">
					<SwitchField
						label="Forget the stored password"
						description="For a server that takes no credentials. The password is removed when you save."
						bind:checked={removePassword}
						onChange={() => (password = '')}
					/>
				</div>
			{/if}
		</div>
	</Panel>

	<Panel title="From" icon={RiMailLine}>
		{#snippet meta()}
			<Tag small>On every message</Tag>
		{/snippet}

		<div class="grid">
			<Input
				label="From address"
				icon={RiAtLine}
				bind:value={form.from_address}
				type="email"
				maxlength={255}
				placeholder="no-reply@example.com"
				hint="Most providers refuse an address they do not host."
			/>

			<Input
				label="From name"
				bind:value={form.from_name}
				maxlength={100}
				placeholder="Acme"
				hint="What a mail client shows instead of the address."
			/>
		</div>
	</Panel>

	<Panel title="Try it" icon={RiMailSendLine}>
		{#snippet meta()}
			<Tag small>Sent with what is on this form</Tag>
		{/snippet}

		<p class="note">
			Sends one message with the settings above, whether or not they have been saved — so a server
			can be tried before anybody's sign-in depends on it. A password left blank is the stored one.
		</p>

		{#if testError}
			<Alert>{testError}</Alert>
		{:else if testSent}
			<Alert tone="success">
				<RiCheckLine size="16" /> Sent to {testSent}. If it does not arrive, look in the spam folder
				before changing anything.
			</Alert>
		{/if}

		<div class="try">
			<Input
				label="Send a test message to"
				icon={RiAtLine}
				bind:value={testTo}
				type="email"
				maxlength={255}
				placeholder="you@example.com"
			/>

			<Button
				variant="outline"
				onclick={sendTest}
				loading={test.isPending}
				disabled={test.isPending || testTo.trim() === '' || !complete}
			>
				{test.isPending ? 'Sending…' : 'Send'}
			</Button>
		</div>
	</Panel>

	{#if dirty}
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
		gap: var(--space-4);
	}

	/* Two fields to a row, each as wide as the other, and a `wide` one across
	   both: a host name is read in full rather than in half. */
	.grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		align-items: start;
		gap: var(--space-3) var(--space-4);
	}

	.wide {
		grid-column: 1 / -1;
	}

	/* The address and the button on one line, the button sitting on the
	   field's baseline rather than on its label's. */
	.try {
		display: flex;
		align-items: flex-end;
		gap: var(--space-3);
	}

	.try :global(> :first-child) {
		flex: 1;
		min-width: 0;
	}

	.note {
		margin: 0 0 var(--space-3);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		line-height: 1.5;
	}

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

	@media (max-width: 40rem) {
		.grid {
			grid-template-columns: 1fr;
		}
	}
</style>
