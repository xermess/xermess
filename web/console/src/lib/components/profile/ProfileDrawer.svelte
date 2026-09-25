<script lang="ts">
	import { untrack } from 'svelte';
	import { invalidate } from '$app/navigation';
	import { ADMIN_DEPENDENCY, MIN_ADMIN_PASSWORD } from '$lib/constants';
	import {
		RiComputerLine,
		RiLockPasswordLine,
		RiShieldKeyholeLine,
		RiUserLine
	} from 'svelte-remixicon';
	import {
		adminApi,
		messageOf,
		mfaApi,
		type Admin,
		type AdminSession,
		type MfaStatus
	} from '$lib/api';
	import RecoveryCodes from '$lib/components/mfa/RecoveryCodes.svelte';
	import TotpSetup from '$lib/components/mfa/TotpSetup.svelte';
	import SessionList from '$lib/components/profile/SessionList.svelte';
	import {
		Alert,
		Button,
		Drawer,
		Input,
		List,
		ListItem,
		Panel,
		PasswordInput,
		Tag,
		Thumb
	} from '$lib/components/ui';
	import { formatRelative } from '$lib/utils/format';

	/** Which half of the drawer the account menu asked for: the account itself
	    — who you are and how you sign in — or its settings — the second factor
	    and where you are signed in. */
	export type ProfileView = 'account' | 'settings';

	type Props = {
		admin: Admin;
		open?: boolean;
		view?: ProfileView;
	};

	let { admin, open = $bindable(false), view = 'account' }: Props = $props();

	const heading = $derived(
		view === 'account'
			? {
					title: 'Account',
					description: 'Your name, the email you sign in with, and your password.'
				}
			: {
					title: 'Settings',
					description: 'Two-factor sign-in, and everywhere you are signed in.'
				}
	);

	/* ---- Profile: name and address ------------------------------------ */

	let firstName = $state('');
	let lastName = $state('');
	let email = $state('');
	/** Asked for only when the address changes: it is what they sign in
	    with, so a session left open is not enough to move it. */
	let emailPassword = $state('');
	let savingProfile = $state(false);
	let profileError = $state('');
	let profileSaved = $state(false);

	const emailChanged = $derived(email.trim().toLowerCase() !== admin.email.toLowerCase());
	const profileChanged = $derived(
		firstName.trim() !== admin.first_name || lastName.trim() !== admin.last_name || emailChanged
	);

	async function saveProfile(event: SubmitEvent) {
		event.preventDefault();
		savingProfile = true;
		profileError = '';
		profileSaved = false;

		try {
			await adminApi.updateProfile({
				first_name: firstName.trim(),
				last_name: lastName.trim(),
				email: email.trim(),
				current_password: emailChanged ? emailPassword : undefined
			});
			emailPassword = '';
			profileSaved = true;
			// The name is in the header and the address in the menu: both
			// come from the layout's read of the administrator, and only that
			// is read again — not the page open behind the drawer.
			await invalidate(ADMIN_DEPENDENCY);
		} catch (err) {
			profileError = messageOf(err, 'Could not save your profile');
		} finally {
			savingProfile = false;
		}
	}

	/* ---- Password ------------------------------------------------------ */

	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let changingPassword = $state(false);
	let passwordError = $state('');
	let passwordChanged = $state(false);

	async function changePassword(event: SubmitEvent) {
		event.preventDefault();
		passwordError = '';
		passwordChanged = false;

		if (newPassword.length < MIN_ADMIN_PASSWORD) {
			passwordError = `The new password must be at least ${MIN_ADMIN_PASSWORD} characters.`;
			return;
		}
		if (newPassword !== confirmPassword) {
			passwordError = 'The two new passwords are not the same.';
			return;
		}

		changingPassword = true;

		try {
			await adminApi.changePassword(currentPassword, newPassword);
			currentPassword = newPassword = confirmPassword = '';
			passwordChanged = true;
			// The settings view reads sessions when it opens; this view has no
			// session list to refresh.
		} catch (err) {
			passwordError = messageOf(err, 'Could not change your password');
		} finally {
			changingPassword = false;
		}
	}

	/** What the panel is doing: reading the account, or one of the two things
	    that take over the whole body while they are under way. A second
	    factor is set up in steps, and each step needs the room. */
	type Mode = 'reading' | 'enrolling' | 'replacing' | 'codes';

	let mode = $state<Mode>('reading');

	let sessions = $state<AdminSession[]>([]);
	let mfa = $state<MfaStatus | null>(null);
	let loading = $state(false);
	let error = $state('');
	let settingsRequest = 0;
	let wasOpen = false;
	let previousView: ProfileView | undefined;

	/** Turning a factor off, and asking for new recovery codes, both take a
	    code from the authenticator: the field is shown in place of the row
	    that asked for it rather than in a dialog over a dialog. */
	let asking = $state<'disable' | 'codes' | null>(null);
	let code = $state('');
	let busy = $state(false);
	let newCodes = $state<string[]>([]);

	const activeSessions = $derived(sessions.filter((session) => session.active).length);

	/** Read out of the status here rather than inside the rows: a snippet is
	    a closure of its own, so narrowing `mfa` above it does not reach in. */
	const codesLeft = $derived(mfa?.recovery_codes_left ?? 0);

	/** Read the settings when the panel opens, and again after anything here
	    changes one of them: the drawer is built fresh each time it opens, so
	    there is no stale copy to invalidate. */
	async function loadSettings() {
		const request = ++settingsRequest;
		loading = true;
		sessions = [];
		mfa = null;
		error = '';

		try {
			const [own, status] = await Promise.all([adminApi.profileSessions(), mfaApi.status()]);
			if (request !== settingsRequest || !open || view !== 'settings') return;
			sessions = own.sessions;
			mfa = status.mfa;
		} catch (err) {
			if (request === settingsRequest && open && view === 'settings') {
				error = messageOf(err, 'Could not read your account');
			}
		} finally {
			if (request === settingsRequest) loading = false;
		}
	}

	// Opening starts every form from the account as it stands. A view change
	// while the drawer is already open only swaps the settings pane, so an
	// account edit is not silently discarded. Saving refreshes `admin` and
	// must not start the forms over.
	$effect(() => {
		const currentView = view;
		const currentOpen = open;
		if (!currentOpen) {
			wasOpen = false;
			previousView = currentView;
			settingsRequest += 1;
			loading = false;
			return;
		}

		const opening = !wasOpen;
		const changedView = previousView !== currentView;
		wasOpen = true;
		previousView = currentView;

		untrack(() => {
			if (opening) {
				firstName = admin.first_name;
				lastName = admin.last_name;
				email = admin.email;
				emailPassword = currentPassword = newPassword = confirmPassword = '';
				profileError = passwordError = '';
				profileSaved = passwordChanged = false;
			}

			if (opening || changedView) {
				mode = 'reading';
				asking = null;
				code = '';
				newCodes = [];
				error = '';
				if (currentView === 'settings') void loadSettings();
				else {
					settingsRequest += 1;
					loading = false;
				}
			}
		});
	});

	function backToReading() {
		mode = 'reading';
		asking = null;
		code = '';
		newCodes = [];
		if (open && view === 'settings') void loadSettings();
	}

	async function disable() {
		busy = true;
		error = '';

		try {
			await mfaApi.disable(code.trim());
			backToReading();
		} catch (err) {
			error = messageOf(err, 'Could not turn two-factor sign-in off');
		} finally {
			busy = false;
		}
	}

	async function regenerate() {
		busy = true;
		error = '';

		try {
			const { recovery_codes } = await mfaApi.recoveryCodes(code.trim());
			newCodes = recovery_codes;
			mode = 'codes';
			asking = null;
			code = '';
		} catch (err) {
			error = messageOf(err, 'Could not make new recovery codes');
		} finally {
			busy = false;
		}
	}
</script>

<Drawer bind:open title={heading.title} description={heading.description}>
	<div class="sections">
		{#if mode === 'enrolling' || mode === 'replacing'}
			<Panel
				title={mode === 'replacing' ? 'Replace your authenticator' : 'Set up an authenticator'}
				icon={RiShieldKeyholeLine}
			>
				<TotpSetup
					replacing={mode === 'replacing'}
					onDone={backToReading}
					onCancel={() => (mode = 'reading')}
				/>
			</Panel>
		{:else if mode === 'codes'}
			<Panel title="New recovery codes" icon={RiShieldKeyholeLine}>
				<RecoveryCodes codes={newCodes} onDone={backToReading} />
			</Panel>
		{:else}
			{#if view === 'settings' && error}<Alert>{error}</Alert>{/if}

			{#if view === 'account'}
				<Panel title="Profile" icon={RiUserLine}>
					<form class="form" onsubmit={saveProfile}>
						<div class="pair">
							<Input label="First name" bind:value={firstName} autocomplete="given-name" required />
							<Input label="Last name" bind:value={lastName} autocomplete="family-name" />
						</div>
						<Input
							label="Email"
							type="email"
							bind:value={email}
							autocomplete="email"
							hint="You sign in with it, and anything about this account is sent to it."
							required
						/>
						{#if emailChanged}
							<PasswordInput
								label="Current password, to change your email"
								bind:value={emailPassword}
								required
							/>
						{/if}

						{#if profileError}<Alert>{profileError}</Alert>{/if}
						{#if profileSaved}<Alert tone="success">Your profile is saved.</Alert>{/if}

						<div class="actions">
							<Button
								type="submit"
								size="sm"
								loading={savingProfile}
								disabled={savingProfile ||
									!profileChanged ||
									firstName.trim() === '' ||
									(emailChanged && emailPassword === '')}
							>
								Save profile
							</Button>
						</div>
					</form>
				</Panel>

				<Panel title="Password" icon={RiLockPasswordLine}>
					<form class="form" onsubmit={changePassword}>
						<PasswordInput label="Current password" bind:value={currentPassword} required />
						<div class="pair">
							<PasswordInput
								label="New password"
								bind:value={newPassword}
								autocomplete="new-password"
								required
							/>
							<PasswordInput
								label="Repeat the new password"
								bind:value={confirmPassword}
								autocomplete="new-password"
								required
							/>
						</div>
						<p class="quiet">
							{`At least ${MIN_ADMIN_PASSWORD} characters. Changing it signs you out everywhere but here.`}
						</p>

						{#if passwordError}<Alert>{passwordError}</Alert>{/if}
						{#if passwordChanged}
							<Alert tone="success"
								>Your password is changed. Your other sessions are signed out.</Alert
							>
						{/if}

						<div class="actions">
							<Button
								type="submit"
								size="sm"
								loading={changingPassword}
								disabled={changingPassword ||
									currentPassword === '' ||
									newPassword === '' ||
									confirmPassword === ''}
							>
								Change password
							</Button>
						</div>
					</form>
				</Panel>
			{:else}
				<Panel title="Two-factor sign-in" icon={RiShieldKeyholeLine} flush>
					{#snippet meta()}
						{#if loading}
							<Tag>Loading</Tag>
						{:else if mfa?.enabled}
							<Tag tone="success" dot strong>On</Tag>
						{:else if mfa?.required}
							<Tag tone="danger" dot strong>Required</Tag>
						{:else if mfa}
							<Tag dot>Off</Tag>
						{:else}
							<Tag>Unknown</Tag>
						{/if}
					{/snippet}

					{#if !mfa}
						<p class="quiet padded">{loading ? 'Reading your account…' : 'Not available.'}</p>
					{:else if !mfa.enabled}
						<div class="prose">
							<p>
								An authenticator app asks for a six-digit code when you sign in, so a password on
								its own is not enough to get in as you.
							</p>
							<Button size="sm" onclick={() => (mode = 'enrolling')}>Set one up</Button>
						</div>
					{:else}
						<List label="Two-factor">
							<ListItem
								title="Authenticator"
								description={mfa.confirmed_at
									? `Set up ${formatRelative(mfa.confirmed_at)}`
									: 'Set up'}
							>
								{#snippet lead()}<Thumb icon={RiShieldKeyholeLine} />{/snippet}
								{#snippet end()}
									<Button size="sm" variant="subtle" onclick={() => (mode = 'replacing')}>
										Replace
									</Button>
								{/snippet}
							</ListItem>

							<ListItem
								title="Recovery codes"
								description="Each one signs you in once if you lose your phone."
							>
								{#snippet end()}
									<Tag small tone={codesLeft > 2 ? 'neutral' : 'warning'}>
										{codesLeft} left
									</Tag>
									<Button size="sm" variant="subtle" onclick={() => (asking = 'codes')}>New</Button>
								{/snippet}
							</ListItem>

							{#if !mfa.required}
								<ListItem
									title="Turn it off"
									description="A password alone will sign you in again."
								>
									{#snippet end()}
										<Button
											size="sm"
											variant="subtle"
											colorPalette="danger"
											onclick={() => (asking = 'disable')}
										>
											Turn off
										</Button>
									{/snippet}
								</ListItem>
							{/if}
						</List>

						{#if asking}
							<!-- Both of these change how this account is protected, so
					     both ask the authenticator to prove it is the reader
					     holding the phone and not the browser. -->
							<div class="asking">
								<Input
									label={asking === 'disable'
										? 'Code from your authenticator, to turn it off'
										: 'Code from your authenticator, for new codes'}
									bind:value={code}
									inputmode="numeric"
									autocomplete="one-time-code"
									maxlength={16}
									disabled={busy}
								/>

								<div class="buttons">
									<Button
										size="sm"
										variant="subtle"
										disabled={busy}
										onclick={() => {
											asking = null;
											code = '';
										}}
									>
										Cancel
									</Button>
									<Button
										size="sm"
										colorPalette={asking === 'disable' ? 'danger' : 'neutral'}
										loading={busy}
										disabled={busy || code.trim() === ''}
										onclick={asking === 'disable' ? disable : regenerate}
									>
										{asking === 'disable' ? 'Turn off' : 'Make new codes'}
									</Button>
								</div>
							</div>
						{/if}
					{/if}
				</Panel>

				<Panel title="Sessions" icon={RiComputerLine} flush>
					{#snippet meta()}
						{#if loading}
							<Tag>Loading</Tag>
						{:else}
							<Tag tone={activeSessions > 0 ? 'success' : 'neutral'} dot small>
								{activeSessions} active
							</Tag>
						{/if}
					{/snippet}

					{#if loading && sessions.length === 0}
						<p class="quiet padded">Reading your sessions…</p>
					{:else}
						<SessionList {sessions} />
					{/if}
				</Panel>
			{/if}
		{/if}
	</div>
</Drawer>

<style>
	/* The panels, in the column the drawer's body gives them. */
	.sections {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
		scroll-behavior: smooth;
	}

	.quiet {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.form {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}

	.pair {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-3);
	}

	@media (max-width: 34rem) {
		.pair {
			grid-template-columns: 1fr;
		}
	}

	.actions {
		display: flex;
		justify-content: flex-end;
	}

	.padded {
		padding: var(--space-4);
	}

	/* The panel is flush for its list rows, so text in it brings its own
	   padding back. */
	.prose {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: var(--space-3);
		padding: var(--space-4);
	}

	.prose p {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		line-height: 1.55;
	}

	/* The code asked for before a change to two-factor: inside the panel it
	   belongs to, ruled off from the rows above it. */
	.asking {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
		padding: var(--space-4);
		border-top: 1px solid var(--color-border);
		background: var(--color-surface-alt);
	}

	.buttons {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-2);
	}
</style>
