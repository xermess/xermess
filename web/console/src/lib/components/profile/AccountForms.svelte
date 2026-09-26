<script lang="ts">
	import { invalidate } from '$app/navigation';
	import { RiLockPasswordLine, RiUserLine } from 'svelte-remixicon';
	import { adminApi, messageOf, type Admin } from '$lib/api';
	import { Alert, Button, FieldGrid, Input, Panel, PasswordInput } from '$lib/components/ui';
	import { ADMIN_DEPENDENCY, MIN_ADMIN_PASSWORD } from '$lib/constants';

	type Props = { admin: Admin };

	let { admin }: Props = $props();

	/* The forms start from the account as it stands. This component is
	   mounted each time the account dialog opens and taken down when it
	   closes, so there is nothing to reset: an opening is a fresh copy.
	   Saving refreshes `admin`, which leaves what was typed alone. */

	/* ---- Profile: name and address ------------------------------------ */

	// svelte-ignore state_referenced_locally
	let firstName = $state(admin.first_name);
	// svelte-ignore state_referenced_locally
	let lastName = $state(admin.last_name);
	// svelte-ignore state_referenced_locally
	let email = $state(admin.email);
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
			// is read again — not the page open behind the dialog.
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
		} catch (err) {
			passwordError = messageOf(err, 'Could not change your password');
		} finally {
			changingPassword = false;
		}
	}
</script>

<div class="sections">
	<Panel title="Profile" icon={RiUserLine}>
		<form class="form" onsubmit={saveProfile}>
			<FieldGrid>
				<Input label="First name" bind:value={firstName} autocomplete="given-name" required />
				<Input label="Last name" bind:value={lastName} autocomplete="family-name" />
			</FieldGrid>
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
			<FieldGrid>
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
			</FieldGrid>
			<p class="quiet">
				{`At least ${MIN_ADMIN_PASSWORD} characters. Changing it signs you out everywhere but here.`}
			</p>

			{#if passwordError}<Alert>{passwordError}</Alert>{/if}
			{#if passwordChanged}
				<Alert tone="success">Your password is changed. Your other sessions are signed out.</Alert>
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
</div>

<style>
	.sections {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}

	.form {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}

	.quiet {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.actions {
		display: flex;
		justify-content: flex-end;
	}
</style>
