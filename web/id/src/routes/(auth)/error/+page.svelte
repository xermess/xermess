<script lang="ts">
	import { useTranslator } from '$lib/i18n';
	import { page } from '$app/state';
	import { Alert, AuthCard } from '$lib/components';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const t = useTranslator();

	const code = $derived(page.url.searchParams.get('error') ?? 'server_error');
	const description = $derived(page.url.searchParams.get('error_description') ?? '');

	/** A failed sign-in elsewhere names its reason with the same code the API
	    answers with, so it is said the same way, in the reader's language. */
	const reason = $derived(page.url.searchParams.get('reason') ?? '');
	const said = $derived(reason !== '' && t.has(`error.${reason}`));

	/** Otherwise the code is the authorization endpoint's own, meant for
	    developers: a sentence for the ones a person can do something about,
	    and what the server said kept underneath for whoever has to debug it. */
	const explanation = $derived.by(() => {
		if (said) return t(`error.${reason}`);

		const key = `error_page.reason.${code}`;
		return t.has(key) ? t(key) : t('error_page.reason.other');
	});
</script>

<svelte:head>
	<title>{t('error_page.head')}</title>
</svelte:head>

<AuthCard organization={data.organization} title={t('error_page.title')} subtitle={explanation}>
	{#if description && !said}
		<Alert tone="warning"><code>{code}</code> {description}</Alert>
	{/if}
	<p class="muted">
		{data.organization?.support_email ? t('error_page.body_support') : t('error_page.body')}
	</p>
</AuthCard>

<style>
	code {
		margin-right: var(--space-1);
		font-family: var(--font-mono);
		font-weight: 700;
	}

	.muted {
		margin-top: var(--space-4);
		color: var(--color-text-hint);
		text-align: center;
	}
</style>
