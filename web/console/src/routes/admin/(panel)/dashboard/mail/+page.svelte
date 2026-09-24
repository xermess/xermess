<script lang="ts">
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { RiFileTextLine, RiRefreshLine, RiServerLine } from 'svelte-remixicon';
	import { IconButton, PageContainer, PageHeader, Tabs } from '$lib/components/ui';
	import MailContent from '$lib/components/mail/MailContent.svelte';
	import MailServer from '$lib/components/mail/MailServer.svelte';
	import { keys, mailContentOptions, mailOptions } from '$lib/query';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	const settings = createQuery(() => mailOptions(data.mail));
	const content = createQuery(() => mailContentOptions(data.content));

	/** The server on one tab and the words on the other: they are two jobs,
	    and a page that asks for a host name and a translation at once is a
	    page nobody finishes. */
	let tab = $state('server');

	const tabs = [
		{ value: 'server', label: 'Mail server', icon: RiServerLine },
		{ value: 'content', label: 'Email content', icon: RiFileTextLine }
	];

	function reload() {
		queryClient.invalidateQueries({ queryKey: keys.mail.all });
	}
</script>

<svelte:head><title>Mail · xermess admin</title></svelte:head>

<PageContainer>
	<div class="page">
		<PageHeader crumbs={['Dashboard', 'Mail']}>
			{#snippet secondary()}
				<IconButton icon={RiRefreshLine} label="Reload these settings" onclick={reload} />
			{/snippet}
		</PageHeader>

		<p class="lead">
			How this installation sends email — the server it hands a message to, and what each message
			says. Password resets, address confirmations and one-time codes all go this way.
		</p>

		<Tabs {tabs} bind:value={tab} label="Mail settings">
			{#snippet panel(value)}
				{#if value === 'server'}
					<MailServer settings={settings.data} />
				{:else}
					<MailContent content={content.data} />
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
