<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { useQueryClient } from '@tanstack/svelte-query';
	import { Dialog } from '@ark-ui/svelte/dialog';
	import { Menu } from '@ark-ui/svelte/menu';
	import { Portal } from '@ark-ui/svelte/portal';
	import {
		RiArrowDownSLine,
		RiBookOpenLine,
		RiExternalLinkLine,
		RiGithubFill,
		RiLogoutBoxRLine,
		RiSettings3Line,
		RiUserLine
	} from 'svelte-remixicon';
	import { adminApi, type Admin } from '$lib/api';
	import { Button, Icon, Tag, type Size } from '$lib/components/ui';
	import { DOCS_URL, GITHUB_URL } from '$lib/constants';
	import { initials } from '$lib/utils/format';
	import ProfileDrawer, { type ProfileView } from '$lib/components/profile/ProfileDrawer.svelte';

	type Props = { admin: Admin; size?: Size };

	let { admin, size = 'md' }: Props = $props();

	const queryClient = useQueryClient();

	let signingOut = $state(false);
	let drawerOpen = $state(false);
	/** Which half of the account drawer the chosen row opens. */
	let drawerView = $state<ProfileView>('account');

	function openDrawer(view: ProfileView) {
		drawerView = view;
		drawerOpen = true;
	}
	let confirmingSignOut = $state(false);

	/** The name the account goes by — the server's first and last name —
	    or the address when neither was given. */
	const name = $derived(admin.full_name.trim() || admin.email);

	const monogram = $derived(initials(name));

	/** The roles shown under the address. The super administrator badge is
	    separate because it is a standing, not one of the roles the API lists. */
	const roles = $derived(admin.roles.filter((role) => role !== 'super_admin'));

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

<!-- The account, named in the bar: who is signed in is the one thing the
     header answers without being asked. The menu under it says the rest —
     the address and roles — and holds the two things done to an account from
     here: its settings, and leaving. -->
<Menu.Root positioning={{ placement: 'bottom-end', gutter: 6 }}>
	<Menu.Trigger
		class="control account-trigger"
		data-size={size}
		data-variant="ghost"
		data-palette="neutral"
		aria-label="Account: {name}"
	>
		<span class="avatar" aria-hidden="true">{monogram}</span>
		<span class="name">{name}</span>
		<span class="chevron" aria-hidden="true"><Icon icon={RiArrowDownSLine} size="1rem" /></span>
	</Menu.Trigger>

	<Portal>
		<Menu.Positioner>
			<Menu.Content class="account-menu">
				<!-- Who is signed in, by the address they sign in with and what they
				     may do: the name is on the button this menu opens from. -->
				<div class="identity">
					<span class="caption">Signed in as</span>
					<span class="email" title={admin.email}>{admin.email}</span>
					{#if admin.is_super_admin || roles.length > 0}
						<span class="roles">
							{#if admin.is_super_admin}
								<Tag small tone="info">Super administrator</Tag>
							{/if}
							{#each roles as role (role)}
								<Tag small>{role}</Tag>
							{/each}
						</span>
					{/if}
				</div>

				<Menu.Separator />

				<Menu.ItemGroup>
					<Menu.Item value="account" onSelect={() => openDrawer('account')}>
						<Icon icon={RiUserLine} />
						Account
					</Menu.Item>
					<Menu.Item value="settings" onSelect={() => openDrawer('settings')}>
						<Icon icon={RiSettings3Line} />
						Settings
					</Menu.Item>
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
					data-palette="danger"
					onSelect={() => (confirmingSignOut = true)}
				>
					<Icon icon={RiLogoutBoxRLine} />
					Sign out
				</Menu.Item>
			</Menu.Content>
		</Menu.Positioner>
	</Portal>
</Menu.Root>

<ProfileDrawer {admin} view={drawerView} bind:open={drawerOpen} />

<Dialog.Root
	open={confirmingSignOut}
	onOpenChange={(details) => {
		if (!signingOut) confirmingSignOut = details.open;
	}}
	lazyMount
	unmountOnExit
>
	<Portal>
		<Dialog.Backdrop />
		<Dialog.Positioner class="signout-positioner">
			<Dialog.Content class="signout-dialog">
				<Dialog.Title>Sign out?</Dialog.Title>
				<Dialog.Description
					>You will need to sign in again to manage the console.</Dialog.Description
				>

				<div class="signout-actions">
					<Button
						variant="subtle"
						disabled={signingOut}
						onclick={() => (confirmingSignOut = false)}
					>
						Cancel
					</Button>
					<Button colorPalette="danger" loading={signingOut} onclick={signOut}>Sign out</Button>
				</div>
			</Dialog.Content>
		</Dialog.Positioner>
	</Portal>
</Dialog.Root>

<style>
	/* The avatar, the name and a chevron saying it opens. The name is held
	   to a width so a long one cannot push the bar about; it is cut, and the
	   menu has it in full. */
	:global(.account-trigger.control) {
		gap: var(--space-2);
		max-width: 16rem;
		padding: 0 var(--space-1) 0 3px;
		color: var(--color-text);
		font-weight: 500;
	}

	:global(.account-trigger.control[data-state='open']) {
		background: var(--palette-subtle);
	}

	/* The initials in a circle, beside the name in the bar. */
	.avatar {
		--avatar-size: 26px;

		display: grid;
		flex: none;
		place-items: center;
		width: var(--avatar-size);
		height: var(--avatar-size);
		border-radius: var(--radius-pill);
		background: var(--color-accent);
		color: var(--color-accent-text);
		font-size: var(--text-xs);
		font-weight: 700;
	}

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

	:global(.account-trigger.control[data-state='open']) .chevron {
		transform: rotate(180deg);
	}

	/* One width, whoever is signed in: a menu that changes shape with the
	   longest name or role list is harder to scan. Long names and addresses
	   are cut; roles wrap onto another line instead. */
	:global(.account-menu) {
		width: min(20rem, calc(100vw - var(--space-4)));
	}

	/* The head of the menu: a quiet caption, the address, and the roles. */
	.identity {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
		padding: 10px 10px 12px;
	}

	.caption {
		color: var(--color-text-hint);
		font-size: var(--text-xs);
		font-weight: 500;
	}

	.email {
		overflow: hidden;
		color: var(--color-text);
		font-size: var(--text-sm);
		font-weight: 600;
		white-space: nowrap;
		text-overflow: ellipsis;
	}

	/* The standing, then each role, wrapping onto another line rather than
	   widening the menu. */
	.roles {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
		margin-top: 6px;
	}

	/* The links' own mark, that they open somewhere else, at the far end. */
	:global(.account-menu a[data-part='item'] svg:last-child) {
		margin-left: auto;
		color: var(--color-text-hint);
	}

	/* Leaving is the one account action that needs asking about first. The
	   panel is centred rather than a drawer, and keeps the same dialog colours
	   and flat dark-theme treatment as every other dialog. */
	:global(.signout-positioner[data-scope='dialog'][data-part='positioner']) {
		justify-content: center;
		align-items: center;
		padding: var(--space-4);
		overflow-y: auto;
	}

	:global(.signout-dialog[data-scope='dialog'][data-part='content']) {
		width: min(26rem, 100%);
		height: auto;
		max-height: calc(100dvh - var(--space-6));
		gap: var(--space-5);
		padding: var(--space-5);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-md);
		overflow-y: auto;
	}

	:global(.signout-dialog[data-scope='dialog'][data-part='content'][data-state='open']) {
		animation: signout-in var(--speed) ease-out;
	}

	:global(.signout-dialog[data-scope='dialog'][data-part='content'][data-state='closed']) {
		animation: signout-out var(--speed) ease-out forwards;
	}

	@keyframes signout-in {
		from {
			opacity: 0;
			transform: scale(0.98) translateY(-4px);
		}
	}

	@keyframes signout-out {
		to {
			opacity: 0;
			transform: scale(0.98) translateY(-4px);
		}
	}

	.signout-actions {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-2);
	}

	/* The header shows these two as icons and drops them at the same width
	   the sidebar's column goes, so the menu carries them from exactly there
	   — the one breakpoint written in both files, for one reason. On a
	   narrow screen the bar has no room for a name either, so the trigger
	   is the avatar alone. */
	:global(.account-menu .when-narrow) {
		display: none;
	}

	@media (prefers-reduced-motion: reduce) {
		.chevron {
			transition: none;
		}

		:global(.signout-dialog[data-scope='dialog'][data-part='content'][data-state='open']),
		:global(.signout-dialog[data-scope='dialog'][data-part='content'][data-state='closed']) {
			animation: none;
		}
	}

	@media (max-width: 55rem) {
		:global(.account-menu .when-narrow) {
			display: block;
		}

		.name,
		.chevron {
			display: none;
		}

		:global(.account-trigger.control) {
			padding: 0 3px;
		}
	}
</style>
