<script lang="ts">
	import { goto, invalidateAll } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { useQueryClient } from '@tanstack/svelte-query';
	import { Menu } from '@ark-ui/svelte/menu';
	import { Portal } from '@ark-ui/svelte/portal';
	import { RiArrowDownSLine, RiLogoutBoxRLine, RiUserSettingsLine } from 'svelte-remixicon';
	import { adminApi, type Admin } from '$lib/api';
	import { Icon, type Size } from '$lib/components/ui';
	import { useTranslator } from '$lib/i18n';

	type Props = { admin: Admin; size?: Size };

	let { admin, size = 'md' }: Props = $props();

	const queryClient = useQueryClient();

	let signingOut = $state(false);

	/** The first letter of the name, which is enough to tell accounts apart. */
	const monogram = $derived((admin.full_name || admin.username).charAt(0).toUpperCase());

	async function open(value: string) {
		if (value === 'profile') {
			await goto(resolve('/admin/profile'));
			return;
		}

		await signOut();
	}

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

	const t = useTranslator();
</script>

<Menu.Root
	positioning={{ placement: 'bottom-end', gutter: 6 }}
	onSelect={(details) => open(details.value)}
>
	<Menu.Trigger
		class="control trigger"
		data-size={size}
		data-variant="ghost"
		data-palette="neutral"
	>
		<span class="monogram" aria-hidden="true">{monogram}</span>
		<span class="name">{admin.username}</span>
		<Icon icon={RiArrowDownSLine} />
	</Menu.Trigger>

	<Portal>
		<Menu.Positioner>
			<Menu.Content>
				<div class="identity">
					<strong>{admin.full_name}</strong>
					<span class="hint">{admin.email}</span>
					{#if admin.roles.length > 0}
						<span class="roles">{admin.roles.join(', ')}</span>
					{/if}
				</div>

				<Menu.Separator />

				<Menu.Item value="profile">
					<Icon icon={RiUserSettingsLine} />
					{t('shell.profile')}
				</Menu.Item>

				<Menu.Item value="sign-out">
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

	.identity {
		display: flex;
		flex-direction: column;
		gap: 2px;
		padding: var(--space-2);
	}

	.hint,
	.roles {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.roles {
		font-family: var(--font-mono);
	}

	@media (max-width: 40rem) {
		.name {
			display: none;
		}
	}
</style>
