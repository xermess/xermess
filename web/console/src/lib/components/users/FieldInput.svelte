<script lang="ts">
	import type { UserField } from '$lib/api';
	import { DatePicker, Input, Switch } from '$lib/components/ui';
	import { fieldIcons } from './fieldIcons';

	type Props = {
		field: UserField;
		/**
		 * Booleans stay booleans, everything else is input text; the form converts. It may be
		 * undefined before the form fills in, so there is deliberately no default (Svelte errors
		 * binding undefined to a prop with one).
		 */
		value: string | boolean | undefined;
	};

	let { field, value = $bindable() }: Props = $props();

	/** The input type that suits the field, so numbers get the browser's
	    own stepper. Dates have a picker of their own, below. */
	const inputType = $derived(field.type === 'number' ? 'number' : 'text');
</script>

{#if field.type === 'bool'}
	<Switch label={field.label} checked={value === true} onChange={(on) => (value = on)} />
{:else if field.type === 'date'}
	<DatePicker
		label={field.label}
		icon={fieldIcons.date}
		value={String(value ?? '')}
		onChange={(day) => (value = day)}
		required={field.is_required}
		clearable={!field.is_required}
	/>
{:else}
	<Input
		label={field.label}
		icon={fieldIcons[field.type]}
		value={String(value ?? '')}
		oninput={(event) => (value = event.currentTarget.value)}
		type={inputType}
		required={field.is_required}
	/>
{/if}
