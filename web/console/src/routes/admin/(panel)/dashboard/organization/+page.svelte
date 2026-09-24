<script lang="ts">
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { RiRefreshLine } from 'svelte-remixicon';
	import { IconButton, PageContainer, PageHeader } from '$lib/components/ui';
	import OrganizationSettings from '$lib/components/organization/OrganizationSettings.svelte';
	import { can } from '$lib/permissions';
	import { keys, organizationOptions } from '$lib/query';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	const settings = createQuery(() => organizationOptions({ organization: data.organization }));

	const canWrite = $derived(can(data.admin, 'organization.write'));
</script>

<svelte:head><title>Organization · xermess admin</title></svelte:head>

<PageContainer>
	<div class="page">
		<PageHeader crumbs={['Dashboard', 'Organization']}>
			{#snippet secondary()}
				<IconButton
					icon={RiRefreshLine}
					label="Reload these settings"
					onclick={() => queryClient.invalidateQueries({ queryKey: keys.organization.settings })}
				/>
			{/snippet}
		</PageHeader>

		<p class="lead">
			Who this installation belongs to: what the sign-in pages show, whom users ask for help, and
			the agreements they accept.
		</p>

		<OrganizationSettings organization={settings.data.organization} editable={canWrite} />
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
