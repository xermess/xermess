<script lang="ts">
	import type { SocialProvider } from '$lib/api';
	import Icon, { type IconName } from './Icon.svelte';

	type Props = {
		/** The providers to offer, in the order the panel put them. */
		providers: SocialProvider[];
		/** The sign-in under way, so the provider can send the person back to
		    the application that asked for it. */
		request?: string | null;
		/** Where to land when there is no application waiting. */
		next?: string | null;
		/** What the buttons sit under: "or continue with" on a sign-in page. */
		label?: string;
		disabled?: boolean;
	};

	let {
		providers,
		request = null,
		next = null,
		label = 'Or continue with',
		disabled = false
	}: Props = $props();

	/** The mark each kind is known by, where this app has one. */
	const marks: Record<string, IconName> = {
		google: 'google',
		apple: 'apple',
		facebook: 'facebook',
		vk: 'vk'
	};

	/** Where the button goes: the server's own address, which sends the
	    browser on to the provider. It is a link rather than a form because
	    that is what it is — leaving this site for another. */
	function href(slug: string): string {
		const query = [
			request ? `request=${encodeURIComponent(request)}` : '',
			next ? `next=${encodeURIComponent(next)}` : ''
		]
			.filter(Boolean)
			.join('&');

		return `/oauth2/social/${encodeURIComponent(slug)}/start${query ? `?${query}` : ''}`;
	}
</script>

{#if providers.length > 0}
	<div class="social">
		<div class="divider"><span>{label}</span></div>

		<div class="buttons" class:single={providers.length === 1}>
			{#each providers as provider (provider.slug)}
				{@const to = disabled ? undefined : href(provider.slug)}
				<!-- The server's own path, which redirects to the provider: a full
				     navigation off this app, not a route of it. -->
				<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
				<a class="provider" href={to} aria-disabled={disabled} data-sveltekit-reload>
					{#if marks[provider.kind]}
						<Icon name={marks[provider.kind]} size="1.125rem" />
					{/if}
					<span>{provider.name}</span>
				</a>
			{/each}
		</div>
	</div>
{/if}

<style>
	.social {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
		margin-top: var(--space-5);
	}

	/* A rule with the label sitting in it. */
	.divider {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.divider::before,
	.divider::after {
		flex: 1;
		height: 1px;
		background: var(--color-border);
		content: '';
	}

	/* Two to a row, and one across the width when it is on its own. */
	.buttons {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-2);
	}

	.buttons.single {
		grid-template-columns: 1fr;
	}

	.provider {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-2);
		min-height: 44px;
		padding: 0 var(--space-3);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		background: var(--color-surface);
		color: var(--color-text);
		font-size: var(--text-base);
		font-weight: 600;
		text-decoration: none;
	}

	.provider:hover {
		background: var(--color-surface-alt);
	}

	.provider span {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.provider[aria-disabled='true'] {
		opacity: 0.6;
		pointer-events: none;
	}
</style>
