<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { useQueryClient } from '@tanstack/svelte-query';
	import {
		RiArrowDownSLine,
		RiLogoutBoxRLine,
		RiShieldUserLine,
		RiUserSettingsLine
	} from 'svelte-remixicon';
	import { adminApi, type Admin } from '$lib/api';
	import {
		ConfirmDialog,
		DropdownMenu,
		Icon,
		MenuGroup,
		MenuInfo,
		MenuItem,
		MenuSeparator,
		Thumb
	} from '$lib/components/ui';
	import { initials } from '$lib/utils/format';
	import ProfileDrawer from '$lib/components/profile/ProfileDrawer.svelte';

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
     it is laid out as the help menu is — a heading over rows with an icon, a
     label and a line under it — and holds what is worth reaching from
     anywhere: who the account is, what it may do, its settings, and leaving. -->
<DropdownMenu label="Account: {name}" shape="labelled">
	{#snippet trigger()}
		<Thumb src={admin.avatar_url} text={monogram} size="xs" shape="circle" tone="accent" />
		<span class="name">{name}</span>
		<span class="chevron" aria-hidden="true"><Icon icon={RiArrowDownSLine} size="1rem" /></span>
	{/snippet}

	<MenuGroup label="Signed in as">
		<MenuInfo label={name} description={admin.email}>
			{#snippet lead()}
				<Thumb src={admin.avatar_url} text={monogram} size="xs" shape="circle" tone="accent" />
			{/snippet}
		</MenuInfo>
		<MenuInfo icon={RiShieldUserLine} label={role} description="Role" />
	</MenuGroup>

	<MenuSeparator />

	<MenuGroup label="Account">
		<MenuItem
			value="account"
			icon={RiUserSettingsLine}
			label="Account settings"
			description="Profile, security and sessions"
			onSelect={() => (accountOpen = true)}
		/>
	</MenuGroup>

	<MenuSeparator />

	<MenuItem
		value="sign-out"
		icon={RiLogoutBoxRLine}
		label="Sign out"
		description="End this session"
		danger
		onSelect={() => (confirmingSignOut = true)}
	/>
</DropdownMenu>

<ProfileDrawer {admin} bind:open={accountOpen} />

<ConfirmDialog
	bind:open={confirmingSignOut}
	title="Sign out?"
	description="You will need to sign in again to manage the console."
	tone="danger"
	icon={RiLogoutBoxRLine}
	confirmLabel="Sign out"
	busy={signingOut}
	onConfirm={signOut}
/>

<style>
	/* A long name is cut rather than pushing the bar about; the menu's
	   address says who it is in full. */
	.name {
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.chevron {
		display: inline-flex;
		flex: none;
		color: var(--color-text-hint);
		transition: transform var(--speed);
	}

	:global(.dropdown-trigger[data-state='open']) .chevron {
		transform: rotate(180deg);
	}

	/* Where the bar has no room for a name, the button is the avatar alone
	   (DropdownMenu makes it a circle to match). */
	@media (max-width: 64rem) {
		.name,
		.chevron {
			display: none;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.chevron {
			transition: none;
		}
	}
</style>
