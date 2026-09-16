<script lang="ts">
	import { page } from '$app/state';
	import { Alert, AuthCard } from '$lib/components';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	/** A sentence for the errors a person can do something about; what the
	    server said is kept underneath for whoever has to debug it. */
	const explanations: Record<string, string> = {
		invalid_request:
			'The application sent an incomplete or incorrect sign-in request. This is a problem with the application, not your account.',
		unauthorized_client: 'This application is not allowed to sign users in right now.',
		access_denied: 'You do not have access to this application.',
		server_error: 'Something went wrong on our side. Try again in a moment.'
	};

	const code = $derived(page.url.searchParams.get('error') ?? 'server_error');
	const description = $derived(page.url.searchParams.get('error_description') ?? '');
	const explanation = $derived(explanations[code] ?? 'The sign-in could not be completed.');
</script>

<svelte:head>
	<title>Sign-in error</title>
</svelte:head>

<AuthCard organization={data.organization} title="We couldn’t sign you in" subtitle={explanation}>
	{#if description}
		<Alert tone="warning"><code>{code}</code> {description}</Alert>
	{/if}
	<p class="muted">
		Go back to the application and try again. If it keeps happening, contact its support{data
			.organization?.support_email
			? ', or ours below'
			: ''}.
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
