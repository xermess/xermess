<script lang="ts">
	import type { ComponentType } from 'svelte';
	import type { HTMLInputAttributes } from 'svelte/elements';
	import { Field as ArkField } from '@ark-ui/svelte/field';
	import CopyButton from './CopyButton.svelte';
	import Field from './Field.svelte';

	type Props = Omit<HTMLInputAttributes, 'value' | 'size'> & {
		label: string;
		value: string;
		icon?: ComponentType;
		hint?: string;
		error?: string;
		required?: boolean;
		disabled?: boolean;
		readOnly?: boolean;
		/** Puts a copy button inside the field, for values that are copied
		    more than typed: ids, client ids, URLs. */
		copyable?: boolean;
		/** A short unit after the value, such as "minutes". */
		suffix?: string;
	};

	let {
		label,
		value = $bindable(''),
		icon,
		hint,
		error,
		required,
		disabled,
		readOnly,
		copyable = false,
		suffix,
		...input
	}: Props = $props();
</script>

<Field
	{label}
	{icon}
	{hint}
	{error}
	{required}
	{disabled}
	{readOnly}
	filled={value !== '' && value !== null && value !== undefined}
>
	{#if copyable || suffix}
		<div class="adorned" class:copyable class:suffixed={suffix !== undefined}>
			<ArkField.Input bind:value readonly={readOnly} {...input} />
			{#if suffix}
				<span class="suffix">{suffix}</span>
			{/if}
			{#if copyable}
				<span class="action"><CopyButton {value} {label} /></span>
			{/if}
		</div>
	{:else}
		<ArkField.Input bind:value readonly={readOnly} {...input} />
	{/if}
</Field>

<style>
	/* The value keeps the whole width; what sits after it — a unit, a copy
	   button — is laid over its right end, level with the value. */
	.adorned {
		position: relative;
	}

	.copyable :global([data-part='input']) {
		padding-right: 46px;
	}

	.suffixed :global([data-part='input']) {
		padding-right: 5.5rem;
	}

	/* Both sit at the right end, above the input they are laid over: the
	   copy button centred on the box, the unit on the value's own line. */
	.action {
		position: absolute;
		top: 50%;
		right: 6px;
		z-index: 2;
		transform: translateY(-50%);
	}

	.suffix {
		position: absolute;
		top: var(--field-value-top);
		right: 14px;
		z-index: 2;
		line-height: var(--field-line);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		pointer-events: none;
	}

	.copyable.suffixed .suffix {
		right: 46px;
	}
</style>
