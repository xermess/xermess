<script lang="ts">
	import type { ComponentType } from 'svelte';
	import Icon from './Icon.svelte';

	type Props = { label: string; icon?: ComponentType; required?: boolean };

	let { label, icon, required = false }: Props = $props();
</script>

<!-- What a field's label says: an icon, the words, and a mark when it is
     required. Every kind of field — Field, Select, PasswordInput — draws its
     label with this, so they all say it the same way. -->
{#if icon}<Icon {icon} size="1em" />{/if}{label}{#if required}<span
		data-part="required-indicator"
		aria-hidden="true">&thinsp;*</span
	>{/if}
