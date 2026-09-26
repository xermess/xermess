<script lang="ts">
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { RiRefreshLine } from 'svelte-remixicon';
	import { BRAND } from '$lib/brand';
	import { IconButton, PageContainer, PageHeader } from '$lib/components/ui';
	import OtpSettings from '$lib/components/otp/OtpSettings.svelte';
	import { keys, otpOptions } from '$lib/query';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	const settings = createQuery(() => otpOptions(data.otp));
</script>

<svelte:head><title>One-time codes · {BRAND.name}</title></svelte:head>

<PageContainer>
	<div class="page">
		<PageHeader crumbs={['Dashboard', 'One-time codes']}>
			{#snippet secondary()}
				<IconButton
					icon={RiRefreshLine}
					label="Reload these settings"
					onclick={() => queryClient.invalidateQueries({ queryKey: keys.otp.settings })}
				/>
			{/snippet}
		</PageHeader>

		<p class="lead">
			The codes this server emails people as they sign in: how long one is, how long it lasts, and
			how many guesses it takes. The codes an authenticator app shows are not these — those are
			fixed by the apps that read them.
		</p>

		<OtpSettings settings={settings.data} />
	</div>
</PageContainer>

<style>
	.page {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}

	.lead {
		margin: calc(var(--space-3) * -1) 0 0;
		color: var(--color-text-hint);
		font-size: var(--text-base);
	}
</style>
