<script lang="ts">
	import type { ComponentType } from 'svelte';
	import type { HTMLAnchorAttributes } from 'svelte/elements';
	import type { ControlProps, Size } from './control';
	import Icon from './Icon.svelte';
	import Tooltip from './Tooltip.svelte';

	type Props = HTMLAnchorAttributes &
		Omit<ControlProps, 'icon' | 'loading' | 'disabled'> & {
			href: string;
			icon: ComponentType;
			/** Where it leads. It is both the tooltip and the name the link is
			    read out by, so it is never left out — an icon on its own says
			    nothing to a screen reader. */
			label: string;
			placement?: 'top' | 'bottom' | 'left' | 'right';
		};

	let {
		href,
		icon,
		label,
		size = 'md',
		variant = 'ghost',
		colorPalette = 'neutral',
		placement = 'bottom',
		...rest
	}: Props = $props();

	/** The glyph is a share of the control, so every size stays in proportion.
	    The same steps as IconButton, which this is the anchor of. */
	const glyph: Record<Size, string> = { sm: '1rem', md: '1.125rem', lg: '1.25rem' };
</script>

<!--
	An IconButton that navigates: an anchor with the same shape and tooltip. The caller supplies
	the href, usually external, so no resolve() applies.
-->
<!-- eslint-disable svelte/no-navigation-without-resolve -->
<Tooltip {label} {placement}>
	{#snippet children(trigger)}
		<a
			{...trigger()}
			{href}
			class="control"
			data-icon="true"
			data-size={size}
			data-variant={variant}
			data-palette={colorPalette}
			aria-label={label}
			{...rest}
		>
			<Icon {icon} size={glyph[size]} />
		</a>
	{/snippet}
</Tooltip>

<!-- eslint-enable svelte/no-navigation-without-resolve -->
