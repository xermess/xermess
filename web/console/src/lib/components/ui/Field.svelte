<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { ComponentType } from 'svelte';
	import { Field } from '@ark-ui/svelte/field';
	import FieldText from './FieldText.svelte';

	type Props = {
		label: string;
		/** The control itself: an input, a select, a textarea. */
		children: Snippet;
		/** Drawn before the label, to say what kind of value this is. */
		icon?: ComponentType;
		/** A line under the control, for what someone needs to know before
		    filling it in. */
		hint?: string;
		/** A line under the control, for what went wrong. It replaces the
		    hint and marks the field as invalid. */
		error?: string;
		required?: boolean;
		disabled?: boolean;
		readOnly?: boolean;
		/** The control holds a value, so the label sits raised in the top of
		    the box rather than where the value would go. */
		filled?: boolean;
	};

	let {
		label,
		children,
		icon,
		hint,
		error,
		required,
		disabled,
		readOnly,
		filled = false
	}: Props = $props();
</script>

<!-- Every field in the panel is this: an outlined box holding a floating
     label and the control, and under the box one line of help or of blame.
     The parts are Ark's, styled once in styles/fields.css, so every field
     that uses this looks the same without saying so. -->
<Field.Root class="field-root" {required} {disabled} {readOnly} invalid={error !== undefined}>
	<div class="field-box" data-float={filled || undefined}>
		<Field.Label><FieldText {label} {icon} {required} /></Field.Label>
		{@render children()}
	</div>

	{#if error}
		<Field.ErrorText>{error}</Field.ErrorText>
	{:else if hint}
		<Field.HelperText>{hint}</Field.HelperText>
	{/if}
</Field.Root>
