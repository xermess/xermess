<script lang="ts">
	import { page } from '$app/state';
	import type { PublicLanguage } from '$lib/api';
	import { useTranslator } from '$lib/i18n';
	import { switchLanguage } from '$lib/i18n/switch';
	import Icon from './Icon.svelte';

	type Props = {
		/** The languages this installation offers. Fewer than two and the
		    picker is not drawn: there is nothing to pick. */
		languages: PublicLanguage[];
		/** The one being shown. */
		current: string;
	};

	let { languages, current }: Props = $props();

	const t = useTranslator();

	let menu = $state<HTMLDetailsElement>();
	let switching = $state(false);

	const shown = $derived(languages.find((language) => language.code === current));

	/** Where each language leads without JavaScript: this same page, asking
	    for it. The server remembers the choice and redirects the parameter
	    away. */
	function href(code: string): string {
		const url = new URL(page.url);
		url.searchParams.set('lang', code);

		return `${url.pathname}${url.search}`;
	}

	/** With JavaScript the page is redrawn where it is instead — no reload, so
	    whatever was typed into a form is still there in the new language. A
	    click with a modifier is somebody opening a new tab, and is left to the
	    link. */
	async function choose(event: MouseEvent, code: string) {
		if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
			return;
		}

		event.preventDefault();
		close();

		if (code === current || switching) return;

		switching = true;
		try {
			await switchLanguage(code);
		} finally {
			switching = false;
		}
	}

	function close(focus = false) {
		if (!menu?.open) return;

		menu.open = false;
		if (focus) menu.querySelector('summary')?.focus();
	}

	// A <details> opens and closes itself, and works before the page has
	// loaded any JavaScript. What it does not do is close when somebody
	// clicks elsewhere or presses Escape, which a menu is expected to.
	$effect(() => {
		const outside = (event: PointerEvent) => {
			if (menu && !menu.contains(event.target as Node)) close();
		};
		const escape = (event: KeyboardEvent) => {
			if (event.key === 'Escape') close(true);
		};

		document.addEventListener('pointerdown', outside);
		document.addEventListener('keydown', escape);

		return () => {
			document.removeEventListener('pointerdown', outside);
			document.removeEventListener('keydown', escape);
		};
	});
</script>

{#if languages.length > 1}
	<details class="picker" bind:this={menu}>
		<summary
			aria-label="{t('shell.language')}: {shown?.native ?? current}"
			aria-busy={switching || undefined}
		>
			<Icon name="globe" size="1.0625rem" />
			<span class="current" lang={current}>{shown?.native ?? current}</span>
			<span class="chevron"><Icon name="chevron" size="1rem" /></span>
		</summary>

		<!-- Each link is this same page in another language: the path is the
		     one the router already resolved, with one parameter added, so there
		     is nothing left for resolve() to do to it. -->
		<!-- eslint-disable svelte/no-navigation-without-resolve -->
		<ul class="menu" aria-label={t('shell.language')}>
			{#each languages as language (language.code)}
				<li>
					<a
						href={href(language.code)}
						hreflang={language.code}
						aria-current={language.code === current ? 'true' : undefined}
						data-sveltekit-reload
						onclick={(event) => choose(event, language.code)}
					>
						<span class="names">
							<span class="native" lang={language.code}>{language.native}</span>
							{#if language.name !== language.native}
								<span class="english" lang="en">{language.name}</span>
							{/if}
						</span>
						{#if language.code === current}
							<Icon name="check" size="1rem" />
						{/if}
					</a>
				</li>
			{/each}
		</ul>
		<!-- eslint-enable svelte/no-navigation-without-resolve -->
	</details>
{/if}

<style>
	.picker {
		position: relative;
	}

	summary {
		display: flex;
		align-items: center;
		gap: var(--space-1);
		height: 38px;
		padding: 0 var(--space-2);
		border-radius: var(--radius-md);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		list-style: none;
		cursor: pointer;
		user-select: none;
		transition:
			background-color var(--speed),
			color var(--speed);
	}

	summary::-webkit-details-marker {
		display: none;
	}

	summary:hover,
	.picker[open] summary {
		background: var(--color-input);
		color: var(--color-text);
	}

	summary:focus-visible {
		outline: 2px solid var(--color-focus);
		outline-offset: 2px;
	}

	/* While the new text is on its way the control says so, rather than
	   looking as if the click did nothing. */
	summary[aria-busy='true'] {
		opacity: 0.6;
		cursor: progress;
	}

	.chevron {
		display: grid;
		transition: transform var(--speed);
	}

	.picker[open] .chevron {
		transform: rotate(180deg);
	}

	.menu {
		position: absolute;
		top: calc(100% + var(--space-1));
		right: 0;
		z-index: 20;
		min-width: 13rem;
		max-height: min(22rem, 70dvh);
		margin: 0;
		padding: var(--space-1);
		overflow-y: auto;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		background: var(--color-surface);
		box-shadow: var(--shadow-card);
		list-style: none;
		animation: drop var(--speed) ease-out;
	}

	@keyframes drop {
		from {
			opacity: 0;
			transform: translateY(-4px);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.menu {
			animation: none;
		}
	}

	a {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		padding: var(--space-2) var(--space-3);
		border-radius: var(--radius-sm);
		color: var(--color-text);
		text-decoration: none;
	}

	a:hover,
	a:focus-visible {
		background: var(--color-input);
		outline: none;
	}

	a[aria-current='true'] {
		color: var(--color-link);
	}

	.names {
		display: flex;
		flex-direction: column;
		line-height: 1.3;
	}

	.native {
		font-weight: 600;
	}

	.english {
		color: var(--color-text-hint);
		font-size: var(--text-xs);
	}

	/* On a phone the globe says enough, and the room is the page's. */
	@media (max-width: 30rem) {
		.current {
			display: none;
		}
	}
</style>
