<script lang="ts">
	import { useQueryClient } from '@tanstack/svelte-query';
	import { RiBuildingLine, RiRefreshLine, RiTranslate2 } from 'svelte-remixicon';
	import { BRAND } from '$lib/brand';
	import { IconButton, PageContainer, PageHeader, Tabs } from '$lib/components/ui';
	import LanguageSettings from '$lib/components/languages/LanguageSettings.svelte';
	import OrganizationSettings from '$lib/components/organization/OrganizationSettings.svelte';
	import { can } from '$lib/permissions';
	import { keys } from '$lib/query';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	/** Only the tabs the administrator's roles let them read: the server left
	    out whatever they may not. */
	const tabs = $derived([
		...(data.organization
			? [{ value: 'organization', label: 'Organization', icon: RiBuildingLine }]
			: []),
		...(data.languages ? [{ value: 'languages', label: 'Languages', icon: RiTranslate2 }] : [])
	]);

	// svelte-ignore state_referenced_locally
	let tab = $state(tabs[0].value);

	function reload() {
		queryClient.invalidateQueries({ queryKey: keys.organization.settings });
		queryClient.invalidateQueries({ queryKey: keys.languages.all });
	}
</script>

<svelte:head><title>Settings · {BRAND.name}</title></svelte:head>

<PageContainer size="lg">
	<div class="page">
		<PageHeader crumbs={['Dashboard', 'Settings']}>
			{#snippet secondary()}
				<IconButton icon={RiRefreshLine} label="Reload these settings" onclick={reload} />
			{/snippet}
		</PageHeader>

		<p class="lead">Who this installation belongs to, and the languages its sign-in pages speak.</p>

		<Tabs {tabs} bind:value={tab} label="Settings">
			{#snippet panel(value)}
				{#if value === 'organization' && data.organization}
					<OrganizationSettings
						initial={data.organization}
						editable={can(data.admin, 'organization.write')}
					/>
				{:else if value === 'languages' && data.languages}
					<LanguageSettings
						initial={data.languages}
						canWrite={can(data.admin, 'languages.write')}
					/>
				{/if}
			{/snippet}
		</Tabs>
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
