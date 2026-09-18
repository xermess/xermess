<script lang="ts">
	import { resolve } from '$app/paths';
	import type { ComponentType } from 'svelte';
	import { Icon, Tooltip } from '$lib/components/ui';
	import type { Section } from './sections';
	import { useTranslator } from '$lib/i18n';

	type Props = {
		route: Section;
		label: string;
		icon: ComponentType;
		status?: 'preview' | 'soon';
		current: boolean;
		/** Folded to its icon: the name moves into a tooltip. */
		collapsed: boolean;
	};

	let { route, label, icon, status, current, collapsed }: Props = $props();

	const tooltip = $derived(
		status === 'soon'
			? `${label} · coming soon`
			: status === 'preview'
				? `${label} · placeholder data`
				: label
	);

	const t = useTranslator();
</script>

<!-- Unfolded, the tooltip would only repeat the name, so it is off. -->
<Tooltip label={tooltip} placement="right" disabled={!collapsed}>
	{#snippet children(trigger)}
		<a
			{...trigger()}
			href={resolve(route)}
			class="link"
			class:current
			class:collapsed
			aria-current={current ? 'page' : undefined}
		>
			<span class="icon"><Icon {icon} /></span>
			<span class="label">{label}</span>

			{#if status === 'soon'}
				<span class="soon">{t('shell.soon')}</span>
			{:else if status === 'preview'}
				<span class="dot" aria-hidden="true"></span>
			{/if}
		</a>
	{/snippet}
</Tooltip>

<style>
	/* The icon keeps its place whether the column is wide or folded — 12px in
	   from the link, 20px from the screen's edge, under the header's mark — so
	   folding never moves it; the words after it just run out of room. The
	   link clips them, and nothing beside the icon can squeeze it. */
	.link {
		position: relative;
		display: flex;
		align-items: center;
		gap: 10px;
		height: 32px;
		padding: 0 12px;
		overflow: hidden;
		border-radius: var(--radius-sm);
		color: var(--color-text);
		font-size: var(--text-base);
		white-space: nowrap;
		text-decoration: none;
		transition:
			background-color var(--speed-fast),
			color var(--speed-fast);
	}

	.link:hover {
		background: var(--color-secondary);
	}

	.link:focus-visible {
		outline: 2px solid var(--color-accent);
		outline-offset: -2px;
	}

	.link.current {
		background: var(--color-secondary-alt);
		font-weight: 600;
	}

	.icon {
		display: inline-flex;
		flex: none;
		width: 16px;
		height: 16px;
		align-items: center;
		justify-content: center;
		color: var(--color-text-hint);
		transition: color var(--speed-fast);
	}

	.link:hover .icon,
	.link.current .icon {
		color: var(--color-text);
	}

	.label {
		flex: 1 1 0;
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		transition: opacity var(--speed);
	}

	/* Not built yet: a word, since "soon" is worth reading. */
	.soon {
		flex: none;
		padding: 1px 5px;
		border-radius: 4px;
		background: var(--surface-info);
		color: color-mix(in srgb, var(--color-info) 80%, var(--color-text));
		font-size: 10px;
		font-weight: 600;
		line-height: 14px;
		letter-spacing: 0.03em;
		text-transform: uppercase;
		transition: opacity var(--speed);
	}

	/* Placeholder data: a quiet dot, so it marks the section without
	   competing with its name. */
	.dot {
		flex: none;
		width: 5px;
		height: 5px;
		margin-right: 2px;
		border-radius: var(--radius-pill);
		background: var(--color-text-disabled);
	}

	/* Folded: the name and the word fade out, and the marks leave the row to
	   become a dot on the icon's corner, so the icon has the row to itself. */
	.collapsed .label,
	.collapsed .soon {
		opacity: 0;
	}

	.collapsed .soon,
	.collapsed .dot {
		position: absolute;
		top: 6px;
		left: 25px;
		width: 6px;
		height: 6px;
		margin: 0;
		padding: 0;
		overflow: hidden;
		border: 1px solid var(--color-surface);
		border-radius: var(--radius-pill);
		font-size: 0;
		box-sizing: content-box;
	}

	.collapsed .soon {
		background: var(--color-info);
		opacity: 1;
	}
</style>
