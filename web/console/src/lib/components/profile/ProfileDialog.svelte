<script lang="ts">
	import { RiSettings3Line, RiUserLine } from 'svelte-remixicon';
	import type { Admin } from '$lib/api';
	import { Modal, Tabs } from '$lib/components/ui';
	import AccountForms from './AccountForms.svelte';
	import SecuritySettings from './SecuritySettings.svelte';

	type Props = {
		admin: Admin;
		open?: boolean;
	};

	let { admin, open = $bindable(false) }: Props = $props();

	/** The account itself — who you are and how you sign in — and its
	    settings — the second factor and where you are signed in. */
	const tabs = [
		{ value: 'account', label: 'Account', icon: RiUserLine },
		{ value: 'settings', label: 'Settings', icon: RiSettings3Line }
	];
</script>

<!-- The signed-in administrator's own account. The dialog's body is built
     when it opens and taken down when it closes, and so is everything in
     it: every opening starts on the Account tab with the forms as the
     account stands, and the settings are read the first time their tab is
     shown. -->
<Modal
	bind:open
	title="Your account"
	description="Who you are, how you sign in, and where you are signed in."
	size="lg"
>
	<Tabs {tabs} value="account" label="Your account">
		{#snippet panel(value)}
			{#if value === 'account'}
				<AccountForms {admin} />
			{:else}
				<SecuritySettings />
			{/if}
		{/snippet}
	</Tabs>
</Modal>
