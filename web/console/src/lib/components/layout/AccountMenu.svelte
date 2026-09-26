<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { useQueryClient } from '@tanstack/svelte-query';
	import {
		RiLogoutBoxRLine,
		RiMailLine,
		RiShieldUserLine,
		RiUserSettingsLine
	} from 'svelte-remixicon';
	import { adminApi, type Admin } from '$lib/api';
	import {
		Button,
		DropdownMenu,
		MenuGroup,
		MenuInfo,
		MenuItem,
		MenuSeparator,
		Modal,
		Thumb
	} from '$lib/components/ui';
	import { initials } from '$lib/utils/format';
	import ProfileDialog from '$lib/components/profile/ProfileDialog.svelte';

	type Props = { admin: Admin };

	let { admin }: Props = $props();

	const queryClient = useQueryClient();

	let signingOut = $state(false);
	let accountOpen = $state(false);
	let confirmingSignOut = $state(false);

	/** The name the account goes by — the server's first and last name —
	    or the address when neither was given. */
	const name = $derived(admin.full_name.trim() || admin.email);

	const monogram = $derived(initials(name));

	/** What the administrator may do, in words: the standing first, since it
	    outranks any role, then the roles the API lists. */
	const role = $derived(
		[
			...(admin.is_super_admin ? ['Super administrator'] : []),
			...admin.roles.filter((name) => name !== 'super_admin')
		].join(', ') || 'No roles'
	);

	async function signOut() {
		signingOut = true;

		try {
			await adminApi.logout();
		} finally {
			// However the server answered, this browser is done with the
			// session: drop everything that was loaded with it — the pages
			// and the cache behind them — and go to the sign-in page. What
			// one administrator saw is not for whoever signs in next.
			queryClient.clear();
			await invalidateAll();
			await goto(resolve('/admin/login'), { replaceState: true });
		}
	}
</script>

<!-- Who is signed in, as the avatar at the end of the header. The menu under
     it is laid out as the help menu is — rows with an icon, a label and a
     line under it — and holds the four things about an account worth
     reaching from anywhere: its address, what it may do, its settings, and
     leaving. -->
<DropdownMenu label="Account: {name}" shape="avatar" width="18rem">
	{#snippet trigger()}
		<Thumb text={monogram} size="xs" shape="circle" tone="accent" />
	{/snippet}

	<MenuGroup label="Signed in as {name}">
		<MenuInfo icon={RiMailLine} label={admin.email} description="Email" />
		<MenuInfo icon={RiShieldUserLine} label={role} description="Role" />
	</MenuGroup>

	<MenuSeparator />

	<MenuGroup>
		<MenuItem
			value="account"
			icon={RiUserSettingsLine}
			label="Account"
			description="Profile, security and sessions"
			onSelect={() => (accountOpen = true)}
		/>
	</MenuGroup>

	<MenuSeparator />

	<MenuItem
		value="sign-out"
		icon={RiLogoutBoxRLine}
		label="Sign out"
		danger
		onSelect={() => (confirmingSignOut = true)}
	/>
</DropdownMenu>

<ProfileDialog {admin} bind:open={accountOpen} />

<Modal
	bind:open={confirmingSignOut}
	title="Sign out?"
	description="You will need to sign in again to manage the console."
	size="sm"
	closable={!signingOut}
>
	{#snippet footer()}
		<Button variant="subtle" disabled={signingOut} onclick={() => (confirmingSignOut = false)}>
			Cancel
		</Button>
		<Button colorPalette="danger" loading={signingOut} onclick={signOut}>Sign out</Button>
	{/snippet}
</Modal>
