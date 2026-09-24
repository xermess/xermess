<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { useQueryClient } from '@tanstack/svelte-query';
	import { Menu } from '@ark-ui/svelte/menu';
	import { Portal } from '@ark-ui/svelte/portal';
	import {
		RiBookOpenLine,
		RiBuildingLine,
		RiComputerLine,
		RiExternalLinkLine,
		RiGithubFill,
		RiLogoutBoxRLine,
		RiShieldKeyholeLine,
		RiUserLine
	} from 'svelte-remixicon';
	import { adminApi, type Admin } from '$lib/api';
	import { Icon, Tag, type Size } from '$lib/components/ui';
	import { DOCS_URL, GITHUB_URL } from '$lib/constants';
	import { can } from '$lib/permissions';
	import ProfileDrawer, { type ProfileSection } from '$lib/components/profile/ProfileDrawer.svelte';

	type Props = { admin: Admin; size?: Size };

	let { admin, size = 'md' }: Props = $props();

	const queryClient = useQueryClient();

	let signingOut = $state(false);
	let profileOpen = $state(false);
	/** Which part of the profile panel the chosen row asked for. */
	let profileSection = $state<ProfileSection>('account');

	function openProfile(section: ProfileSection) {
		profileSection = section;
		profileOpen = true;
	}

	/** The first letter of the name, which is enough to tell accounts apart. */
	const monogram = $derived((admin.full_name || admin.username).charAt(0).toUpperCase());

	/** The panel's own settings pages, only the ones this administrator may
	    open: a row that leads to a refusal is worse than no row. */
	const maySeeOrganization = $derived(can(admin, 'organization.read'));

	/** What this account is, in one line under the address: the standing that
	    outranks every role, or else the roles themselves. A super
	    administrator is said once rather than again as the role behind it. */
	const standing = $derived(
		admin.is_super_admin
			? 'Super administrator'
			: admin.roles.filter((role) => role !== 'super_admin').join(' · ')
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

<Menu.Root positioning={{ placement: 'bottom-end', gutter: 6 }}>
	<Menu.Trigger
		class="control trigger"
		data-size={size}
		data-variant="ghost"
		data-palette="neutral"
		aria-label="Profile"
	>
		<span class="monogram" aria-hidden="true">{monogram}</span>
	</Menu.Trigger>

	<Portal>
		<Menu.Positioner>
			<Menu.Content class="account-menu">
				<div class="identity">
					<span class="face" aria-hidden="true">{monogram}</span>
					<span class="who">
						<strong>{admin.full_name || admin.username}</strong>
						<span class="hint">{admin.email}</span>
					</span>
				</div>

				{#if standing}
					<div class="standing">
						<Tag small tone={admin.is_super_admin ? 'info' : 'neutral'}>
							{standing}
						</Tag>
					</div>
				{/if}

				<Menu.Separator />

				<!-- Three rows that lead somewhere, and nothing that is set
				     here. The theme used to be a submenu off this one; it is a
				     toggle in the bar and a pair of buttons in the profile
				     panel, and a third place to change it was one too many.
				     The two that open the same panel say which part of it they
				     open, so the menu answers "where do I turn two-factor on"
				     without being a settings screen itself. -->
				<Menu.ItemGroup>
					<Menu.Item value="profile" onSelect={() => openProfile('account')}>
						<Icon icon={RiUserLine} />
						Your account
					</Menu.Item>

					<Menu.Item value="security" onSelect={() => openProfile('security')}>
						<Icon icon={RiShieldKeyholeLine} />
						Two-factor sign-in
					</Menu.Item>

					<Menu.Item value="sessions" onSelect={() => openProfile('sessions')}>
						<Icon icon={RiComputerLine} />
						Your sessions
					</Menu.Item>

					{#if maySeeOrganization}
						<Menu.Item
							value="organization"
							onSelect={() => goto(resolve('/admin/dashboard/organization'))}
						>
							<Icon icon={RiBuildingLine} />
							Organization
						</Menu.Item>
					{/if}
				</Menu.ItemGroup>

				<!-- The header carries these two as icons and drops them when
				     the bar runs out of room, so they appear here only at the
				     width where they are missing up there: never both at once,
				     and never gone. Real links, so they can be opened in a
				     tab, copied or middle-clicked the way any other link is. -->
				<Menu.Separator class="when-narrow" />

				<Menu.ItemGroup class="when-narrow">
					<Menu.Item value="documentation">
						{#snippet asChild(item)}
							<!-- Somewhere else entirely, so there is no route for
							     resolve() to make of it. -->
							<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
							<a {...item()} href={DOCS_URL} target="_blank" rel="noreferrer noopener">
								<Icon icon={RiBookOpenLine} />
								Documentation
								<Icon icon={RiExternalLinkLine} size="0.875rem" />
							</a>
						{/snippet}
					</Menu.Item>

					<Menu.Item value="github">
						{#snippet asChild(item)}
							<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
							<a {...item()} href={GITHUB_URL} target="_blank" rel="noreferrer noopener">
								<Icon icon={RiGithubFill} />
								GitHub
								<Icon icon={RiExternalLinkLine} size="0.875rem" />
							</a>
						{/snippet}
					</Menu.Item>
				</Menu.ItemGroup>

				<Menu.Separator />

				<Menu.Item
					value="sign-out"
					class="sign-out"
					closeOnSelect={false}
					disabled={signingOut}
					onSelect={signOut}
				>
					<Icon icon={RiLogoutBoxRLine} />
					{signingOut ? 'Signing out…' : 'Sign out'}
				</Menu.Item>
			</Menu.Content>
		</Menu.Positioner>
	</Portal>
</Menu.Root>

<ProfileDrawer {admin} bind:open={profileOpen} section={profileSection} />

<style>
	/* The account is intentionally an icon-only control: identity details
	   belong in the menu and dialog, leaving the header calm at every width. */
	:global(.trigger.control) {
		padding: 0 var(--space-1);
		color: var(--color-text);
		font-weight: 500;
	}

	:global(.trigger.control[data-state='open']) {
		background: var(--palette-subtle);
	}

	.monogram {
		display: grid;
		place-items: center;
		width: 24px;
		height: 24px;
		border-radius: var(--radius-pill);
		background: var(--color-accent);
		color: var(--color-accent-text);
		font-size: var(--text-sm);
		font-weight: 700;
	}

	/* One width, whoever is signed in: a menu that is as wide as the longest
	   name in it is a menu that changes shape between accounts. Long names
	   and addresses are cut instead, which the rows below already expect. */
	:global(.account-menu) {
		width: 17rem;
	}

	/* Who is signed in, drawn once at the top the way the trigger draws it,
	   so the menu opens out of the button rather than beside it. */
	.identity {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		padding: var(--space-3) var(--space-2) var(--space-2);
	}

	.face {
		display: grid;
		flex-shrink: 0;
		place-items: center;
		width: 34px;
		height: 34px;
		border-radius: var(--radius-pill);
		background: var(--color-accent);
		color: var(--color-accent-text);
		font-size: var(--text-base);
		font-weight: 700;
	}

	.who {
		display: flex;
		flex-direction: column;
		min-width: 0;
		gap: 1px;
	}

	.who strong,
	.hint,
	.standing {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.hint {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	/* What the account is, on its own line under the name: a Tag, because
	   that is how this fact is shown on the Administrators page and in the
	   profile panel, and three spellings of one thing is two too many. */
	.standing {
		padding: 0 var(--space-2) var(--space-3);
	}

	/* Menu rows are divs; this one is a link and has to be told to sit like
	   the rest of them. */
	:global(.account-menu a[data-part='item']) {
		color: var(--color-text);
		text-decoration: none;
	}

	:global(.account-menu a[data-part='item'] svg:last-child) {
		margin-left: auto;
		color: var(--color-text-hint);
	}

	/* Leaving is the one row that is not neutral, and it is last. */
	:global(.account-menu [data-part='item'].sign-out) {
		color: var(--color-danger);
	}

	:global(.account-menu [data-part='item'].sign-out[data-highlighted]) {
		background: var(--surface-danger);
	}

	/* The header shows these two as icons and drops them at the same width
	   the sidebar's column goes, so the menu carries them from exactly there
	   — the one breakpoint written in both files, for one reason. */
	:global(.account-menu .when-narrow) {
		display: none;
	}

	@media (max-width: 55rem) {
		:global(.account-menu .when-narrow) {
			display: block;
		}
	}
</style>
