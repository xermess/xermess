<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { Application, Organization } from '$lib/api';
	import { legalLinks } from '$lib/utils/legal';
	import AppMark from './AppMark.svelte';

	type Props = {
		/** The application being signed in to. Without one the page speaks for
		    the account itself: a reset link opened on its own, an error. */
		application?: Application | null;
		/** The organisation the server signs users in for, whose name stands
		    in where there is no application and whose agreements are linked
		    where the application publishes none. */
		organization?: Organization | null;
		title: string;
		subtitle?: string;
		children: Snippet;
		/** A line under the card, such as "Don't have an account?". */
		below?: Snippet;
	};

	let {
		application = null,
		organization = null,
		title,
		subtitle,
		children,
		below
	}: Props = $props();

	const legal = $derived(legalLinks(application, organization));
</script>

<section class="card" aria-labelledby="auth-title">
	<header>
		<AppMark
			name={application?.name ?? organization?.name}
			logo={application?.logo_uri || organization?.logo_url}
		/>
		<h1 id="auth-title">{title}</h1>
		{#if subtitle}<p class="subtitle">{subtitle}</p>{/if}
	</header>

	{@render children()}
</section>

{#if below}
	<p class="below">{@render below()}</p>
{/if}

{#if legal.length > 0}
	<nav class="legal" aria-label="Legal">
		{#each legal as link (link.href)}
			<!-- The application's own pages, or the organisation's, on their
			     own sites. -->
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a href={link.href} target="_blank" rel="noopener noreferrer">{link.label}</a>
		{/each}
	</nav>
{/if}

<style>
	.card {
		width: 100%;
		padding: var(--space-6);
		border: 1px solid var(--color-border);
		border-radius: 18px;
		background: var(--color-surface);
		box-shadow: var(--shadow-card);
	}

	header {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-2);
		margin-bottom: var(--space-5);
		text-align: center;
	}

	h1 {
		margin-top: var(--space-2);
		font-size: 1.5rem;
		font-weight: 700;
		line-height: 1.25;
		letter-spacing: -0.01em;
	}

	.subtitle {
		color: var(--color-text-hint);
		font-size: var(--text-lg);
	}

	.below {
		margin-top: var(--space-5);
		text-align: center;
		color: var(--color-text-hint);
	}

	.below :global(a) {
		font-weight: 600;
	}

	.legal {
		display: flex;
		justify-content: center;
		flex-wrap: wrap;
		gap: var(--space-1) var(--space-5);
		margin-top: var(--space-3);
		font-size: var(--text-sm);
	}

	.legal a {
		color: var(--color-text-hint);
	}

	.legal a:hover {
		color: var(--color-text);
	}

	/* On a phone the card is the page: no frame, the form gets the width. */
	@media (max-width: 30rem) {
		.card {
			padding: var(--space-2) 0 0;
			border: 0;
			border-radius: 0;
			background: transparent;
			box-shadow: none;
		}
	}
</style>
