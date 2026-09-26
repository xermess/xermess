<script lang="ts">
	import type { ComponentType } from 'svelte';
	import { Menu } from '@ark-ui/svelte/menu';
	import { RiExternalLinkLine } from 'svelte-remixicon';
	import Icon from './Icon.svelte';

	type Props = {
		/** Identifies the row within its menu. */
		value: string;
		icon: ComponentType;
		label: string;
		/** A quieter line under the label, saying what the row leads to. */
		description?: string;
		/** Somewhere off this site. The row becomes a link that opens a tab of
		    its own, and says so with a mark at its end. */
		href?: string;
		/** A row that cannot be taken back, such as signing out. */
		danger?: boolean;
		onSelect?: () => void;
	};

	let { value, icon, label, description, href, danger = false, onSelect }: Props = $props();
</script>

{#snippet content()}
	<Icon {icon} />
	<span class="text">
		{label}
		{#if description}<small>{description}</small>{/if}
	</span>
	{#if href}<span class="external"><Icon icon={RiExternalLinkLine} size="0.875rem" /></span>{/if}
{/snippet}

{#if href}
	<Menu.Item {value}>
		{#snippet asChild(item)}
			<!-- Somewhere else entirely, so there is no route for resolve() to
			     make of it. -->
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
			<a {...item()} {href} target="_blank" rel="noreferrer noopener">{@render content()}</a>
		{/snippet}
	</Menu.Item>
{:else}
	<Menu.Item {value} data-palette={danger ? 'danger' : undefined} {onSelect}>
		{@render content()}
	</Menu.Item>
{/if}

<style>
	.text {
		display: flex;
		flex: 1;
		flex-direction: column;
		min-width: 0;
		line-height: 1.3;
	}

	small {
		overflow: hidden;
		color: var(--color-text-hint);
		font-size: var(--text-xs);
		font-weight: 400;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.external {
		display: inline-flex;
		margin-left: auto;
		color: var(--color-text-hint);
	}
</style>
