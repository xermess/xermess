<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { useQueryClient } from '@tanstack/svelte-query';
	import { Menu } from '@ark-ui/svelte/menu';
	import { Portal } from '@ark-ui/svelte/portal';
	import {
		RiArrowDownSLine,
		RiArrowRightSLine,
		RiBookOpenLine,
		RiBuildingLine,
		RiCheckLine,
		RiExternalLinkLine,
		RiGithubFill,
		RiGlobalLine,
		RiLogoutBoxRLine,
		RiPaletteLine,
		RiUserSettingsLine
	} from 'svelte-remixicon';
	import { adminApi, type Admin } from '$lib/api';
	import { Icon, type Size } from '$lib/components/ui';
	import { DOCS_URL, GITHUB_URL } from '$lib/constants';
	import { useTranslator } from '$lib/i18n';
	import { can } from '$lib/permissions';
	import type { LanguageChoice } from '$lib/state/language.svelte';
	import { setLanguage } from '$lib/state/language.svelte';
	import { theme, type Theme } from '$lib/state/theme.svelte';

	type Props = { admin: Admin; size?: Size };

	let { admin, size = 'md' }: Props = $props();

	const queryClient = useQueryClient();
	const t = useTranslator();

	let signingOut = $state(false);

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
			? t('shell.super_admin')
			: admin.roles.filter((role) => role !== 'super_admin').join(' · ')
	);

	/** The language the panel is drawn in, and the ones it can be drawn in,
	    both settled by the root layout before this renders. */
	const language = $derived(page.data.language as string);
	const panelLanguages = $derived((page.data.panelLanguages ?? []) as LanguageChoice[]);

	const themes: Theme[] = ['light', 'dark'];

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
	>
		<span class="monogram" aria-hidden="true">{monogram}</span>
		<span class="name">{admin.full_name || admin.username}</span>
		<Icon icon={RiArrowDownSLine} />
	</Menu.Trigger>

	<Portal>
		<Menu.Positioner>
			<Menu.Content class="account-menu">
				<div class="identity">
					<span class="face" aria-hidden="true">{monogram}</span>
					<span class="who">
						<strong>{admin.full_name || admin.username}</strong>
						<span class="hint">{admin.email}</span>
						{#if standing}<span class="standing">{standing}</span>{/if}
					</span>
				</div>

				<Menu.Separator />

				<!-- One block, not four: where this account goes, and the two
				     preferences that are this reader's own on this machine.
				     They are few enough to read at a glance, and headings over
				     groups of two were more furniture than the menu needed. -->
				<Menu.ItemGroup>
					<Menu.Item value="profile" onSelect={() => goto(resolve('/admin/profile'))}>
						<Icon icon={RiUserSettingsLine} />
						{t('shell.profile')}
					</Menu.Item>

					{#if maySeeOrganization}
						<Menu.Item
							value="organization"
							onSelect={() => goto(resolve('/admin/dashboard/organization'))}
						>
							<Icon icon={RiBuildingLine} />
							{t('nav.organization')}
						</Menu.Item>
					{/if}

					<Menu.Root positioning={{ placement: 'left-start', gutter: 2 }}>
						<Menu.TriggerItem>
							<Icon icon={RiPaletteLine} />
							{t('shell.theme')}
							<span class="value">{t(`shell.theme_${theme.current}`)}</span>
							<Icon icon={RiArrowRightSLine} />
						</Menu.TriggerItem>

						<Portal>
							<Menu.Positioner>
								<Menu.Content class="account-menu">
									<Menu.RadioItemGroup
										value={theme.current}
										onValueChange={(details) => theme.set(details.value as Theme)}
									>
										{#each themes as option (option)}
											<Menu.RadioItem value={option}>
												<span class="tick">
													<Menu.ItemIndicator><Icon icon={RiCheckLine} /></Menu.ItemIndicator>
												</span>
												<Menu.ItemText>{t(`shell.theme_${option}`)}</Menu.ItemText>
											</Menu.RadioItem>
										{/each}
									</Menu.RadioItemGroup>
								</Menu.Content>
							</Menu.Positioner>
						</Portal>
					</Menu.Root>

					{#if panelLanguages.length > 1}
						<Menu.Root positioning={{ placement: 'left-start', gutter: 2 }}>
							<Menu.TriggerItem>
								<Icon icon={RiGlobalLine} />
								{t('shell.language')}
								<span class="value">
									{panelLanguages.find((one) => one.code === language)?.native ?? language}
								</span>
								<Icon icon={RiArrowRightSLine} />
							</Menu.TriggerItem>

							<Portal>
								<Menu.Positioner>
									<Menu.Content class="account-menu">
										<Menu.RadioItemGroup
											value={language}
											onValueChange={(details) => setLanguage(details.value)}
										>
											{#each panelLanguages as choice (choice.code)}
												<Menu.RadioItem value={choice.code}>
													<span class="tick">
														<Menu.ItemIndicator><Icon icon={RiCheckLine} /></Menu.ItemIndicator>
													</span>
													<Menu.ItemText>{choice.native}</Menu.ItemText>
												</Menu.RadioItem>
											{/each}
										</Menu.RadioItemGroup>
									</Menu.Content>
								</Menu.Positioner>
							</Portal>
						</Menu.Root>
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
								{t('shell.documentation')}
								<Icon icon={RiExternalLinkLine} size="0.875rem" />
							</a>
						{/snippet}
					</Menu.Item>

					<Menu.Item value="github">
						{#snippet asChild(item)}
							<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
							<a {...item()} href={GITHUB_URL} target="_blank" rel="noreferrer noopener">
								<Icon icon={RiGithubFill} />
								{t('shell.github')}
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
					{signingOut ? t('shell.signing_out') : t('shell.sign_out')}
				</Menu.Item>
			</Menu.Content>
		</Menu.Positioner>
	</Portal>
</Menu.Root>

<style>
	/* The shape is the shared control; what belongs to this one is the name
	   beside the monogram, which reads as text rather than as a label on a
	   quiet button. */
	:global(.trigger.control) {
		gap: var(--space-2);
		padding: 0 var(--space-2);
		color: var(--color-text);
		font-weight: 500;
	}

	/* A name as long as somebody cares to have is still one line: the bar's
	   layout does not move because of who signed in. */
	.name {
		max-width: 12rem;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
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
		padding: var(--space-3) var(--space-2);
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

	/* What the account is, said in words rather than worn as a badge: a
	   coloured pill here competed with the rows under it for the first look,
	   and this is a fact about the account, not a warning. */
	.standing {
		margin-top: 2px;
		color: var(--color-text-hint);
		font-size: var(--text-xs);
		letter-spacing: 0.02em;
		text-transform: uppercase;
	}

	/* What a preference is set to now, pushed to the right of its row so the
	   submenu can be read without being opened. */
	.value {
		margin-left: auto;
		padding-left: var(--space-2);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	/* The tick column: kept even while nothing is ticked, so the rows of a
	   radio group do not shift sideways as the choice moves. */
	.tick {
		display: grid;
		flex-shrink: 0;
		place-items: center;
		width: 1rem;
		color: var(--color-accent);
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

	@media (max-width: 40rem) {
		.name {
			display: none;
		}
	}
</style>
