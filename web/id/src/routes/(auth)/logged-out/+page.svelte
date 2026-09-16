<script lang="ts">
	import { resolve } from '$app/paths';
	import { Alert, AuthCard } from '$lib/components';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const app = $derived(data.application);
</script>

<svelte:head>
	<title>Signed out</title>
</svelte:head>

<AuthCard organization={data.organization} application={app} title="You’re signed out">
	<Alert tone="success">You have been signed out on this device.</Alert>

	{#if app?.client_uri}
		<div class="action">
			<!-- The application's own site. -->
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a class="primary" href={app.client_uri}>Return to {app.name}</a>
		</div>
	{/if}

	{#snippet below()}
		<a href={resolve('/login')}>Sign in to your account</a>
	{/snippet}
</AuthCard>

<style>
	.action {
		display: flex;
		margin-top: var(--space-5);
	}

	.primary {
		display: flex;
		flex: 1;
		align-items: center;
		justify-content: center;
		height: var(--control-height);
		border-radius: var(--radius-md);
		background: var(--color-primary);
		color: var(--color-primary-text);
		font-weight: 600;
		text-decoration: none;
	}

	.primary:hover {
		background: var(--color-primary-hover);
		text-decoration: none;
	}
</style>
