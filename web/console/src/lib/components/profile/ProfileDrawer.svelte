<script lang="ts">
	import {
		RiComputerLine,
		RiLockPasswordLine,
		RiMailLine,
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
		Tag,
		Thumb
	} from '$lib/components/ui';
	import { formatDateTime, formatRelative } from '$lib/utils/format';

	/** Which part of the drawer to open on. The account menu has a row for
	    each, so a reader looking for two-factor lands on it rather than on
	    the top of a panel they then have to read through. */
	export type ProfileSection = 'account' | 'security' | 'sessions';

	type Props = {
		admin: Admin;
		open?: boolean;
		section?: ProfileSection;
	};

	let { admin, open = $bindable(false), section = 'account' }: Props = $props();

	/** What the panel is doing: reading the account, or one of the two things
	    that take over the whole body while they are under way. A second
	    factor is set up in steps, and each step needs the room. */
	type Mode = 'reading' | 'enrolling' | 'replacing' | 'codes';

	let mode = $state<Mode>('reading');

	let sessions = $state<AdminSession[]>([]);
	let mfa = $state<MfaStatus | null>(null);
	let loading = $state(false);
	let error = $state('');

	/** Turning a factor off, and asking for new recovery codes, both take a
	    code from the authenticator: the field is shown in place of the row
	    that asked for it rather than in a dialog over a dialog. */
	let asking = $state<'disable' | 'codes' | null>(null);
	let code = $state('');
	let busy = $state(false);
	let newCodes = $state<string[]>([]);

	const initials = $derived(
		(admin.full_name || admin.username)
			.split(/\s+/)
			.filter(Boolean)
			.slice(0, 2)
			.map((part) => part[0])
			.join('')
			.toUpperCase() || admin.email.slice(0, 2).toUpperCase()
	);

	/** What this account is, in one line: the standing that outranks every
	    role, or else the roles themselves. */
	const standing = $derived(
		admin.is_super_admin
			? 'Super administrator'
			: admin.roles.filter((role) => role !== 'super_admin').join(' · ')
	);

	const activeSessions = $derived(sessions.filter((session) => session.active).length);

	/** Read out of the status here rather than inside the rows: a snippet is
	    a closure of its own, so narrowing `mfa` above it does not reach in. */
	const codesLeft = $derived(mfa?.recovery_codes_left ?? 0);

	/** Read when the panel opens, and again after anything here changes one
	    of them: the drawer is built fresh each time it opens, so there is no
	    stale copy to invalidate. */
	async function load() {
		loading = true;
		error = '';

		try {
			const [own, status] = await Promise.all([adminApi.profileSessions(), mfaApi.status()]);
			sessions = own.sessions;
			mfa = status.mfa;
		} catch (err) {
			error = messageOf(err, 'Could not read your account');
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		if (open) void load();
	});

	/** The panel to scroll to, once the drawer has been built. The drawer
	    mounts when it opens, so the element does not exist before this. */
	let body = $state<HTMLElement | null>(null);

	$effect(() => {
		if (!open || section === 'account' || !body) return;

		body
			.querySelector(`[data-section='${section}']`)
			?.scrollIntoView({ block: 'start', behavior: 'smooth' });
	});

	function backToReading() {
		mode = 'reading';
		asking = null;
		code = '';
		newCodes = [];
		void load();
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

<Drawer
	bind:open
	title="Your account"
	description="Who you are here, how you sign in, and where you are signed in."
	meta={admin.username}
	width="34rem"
>
	<div class="sections" bind:this={body}>
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
			{#if error}<Alert>{error}</Alert>{/if}

			<!-- Who is signed in, the way every other record opens: the thumb, the
		     name, and what the account is — not a card of its own. -->
			<List bordered label="Account">
				<ListItem title={admin.full_name || admin.username} description={admin.email}>
					{#snippet lead()}
						<Thumb text={initials} size="md" />
					{/snippet}

					{#snippet end()}
						<Tag tone={admin.status === 'active' ? 'success' : 'neutral'} dot strong>
							{admin.status}
						</Tag>
					{/snippet}
				</ListItem>
			</List>

			<div data-section="account">
				<Panel title="Account" icon={RiUserLine} flush>
					{#snippet meta()}
						{#if standing}<Tag small>{standing}</Tag>{/if}
					{/snippet}

					<List label="Account details">
						<ListItem title="Username" description="What you sign in with.">
							{#snippet end()}<span class="value">{admin.username}</span>{/snippet}
						</ListItem>

						<ListItem title="Email" description="Where anything about this account is sent.">
							{#snippet lead()}<Thumb icon={RiMailLine} />{/snippet}
							{#snippet end()}<span class="value">{admin.email}</span>{/snippet}
						</ListItem>

						<ListItem title="Password" description="Changing it signs out every other session.">
							{#snippet lead()}<Thumb icon={RiLockPasswordLine} />{/snippet}
							{#snippet end()}
								<Button size="sm" variant="subtle" disabled>Change</Button>
							{/snippet}
						</ListItem>

						<ListItem title="Last signed in">
							{#snippet end()}
								<span class="value">
									{#if admin.last_login_at}
										<span title={formatDateTime(admin.last_login_at)}>
											{formatRelative(admin.last_login_at)}
										</span>
									{:else}
										Never
									{/if}
								</span>
							{/snippet}
						</ListItem>
					</List>
				</Panel>
			</div>

			<div data-section="security">
				<Panel title="Two-factor sign-in" icon={RiShieldKeyholeLine} flush>
					{#snippet meta()}
						{#if mfa?.enabled}
							<Tag tone="success" dot strong>On</Tag>
						{:else if mfa?.required}
							<Tag tone="danger" dot strong>Required</Tag>
						{:else}
							<Tag dot>Off</Tag>
						{/if}
					{/snippet}

					{#if !mfa}
						<p class="quiet">{loading ? 'Reading your account…' : 'Not available.'}</p>
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
			</div>

			<div data-section="sessions">
				<Panel title="Sessions" icon={RiComputerLine} flush>
					{#snippet meta()}
						<Tag tone={activeSessions > 0 ? 'success' : 'neutral'} dot small>
							{activeSessions} active
						</Tag>
					{/snippet}

					{#if loading && sessions.length === 0}
						<p class="quiet padded">Reading your sessions…</p>
					{:else}
						<SessionList {sessions} />
					{/if}
				</Panel>
			</div>
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

	/* A value read off the row rather than typed into it: the same size as
	   the title, and never wider than the room left beside it. */
	.value {
		max-width: 16rem;
		overflow: hidden;
		color: var(--color-text);
		font-size: var(--text-sm);
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.quiet {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.padded {
		padding: var(--space-4);
	}

	.prose {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: var(--space-3);
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
