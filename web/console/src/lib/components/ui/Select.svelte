<script lang="ts" generics="Value extends string">
	import type { ComponentType } from 'svelte';
	import { createListCollection, Select as ArkSelect } from '@ark-ui/svelte/select';
	import { Portal } from '@ark-ui/svelte/portal';
	import { RiArrowDownSLine, RiCheckLine, RiCloseLine } from 'svelte-remixicon';
	import FieldText from './FieldText.svelte';
	import Icon from './Icon.svelte';
	import type { SelectOption } from './select';

	type Props = {
		label: string;
		/** The chosen value. An empty string means nothing is chosen yet, so a
		    select that may be left empty is typed `'' | …` by its caller. */
		value: Value;
		/** What can be chosen. A plain string is its own label. */
		options: readonly (Value | SelectOption<Value>)[];
		/** Shown in place of a value while nothing is chosen. */
		placeholder?: string;
		/** Drawn before the label, to say what kind of value this is. */
		icon?: ComponentType;
		/** A line under the control, for what someone needs to know. */
		hint?: string;
		/** A line under the control, for what went wrong. It replaces the hint
		    and marks the field as invalid. */
		error?: string;
		required?: boolean;
		disabled?: boolean;
		/** Shown, and read out, but not open to being changed. */
		readOnly?: boolean;
		/** Offers a button that puts the field back to nothing chosen. */
		clearable?: boolean;
		/** Submitted with a surrounding form, through a hidden native select. */
		name?: string;
		onChange?: (value: Value) => void;
	};

	let {
		label,
		value = $bindable(),
		options,
		placeholder = 'Select…',
		icon,
		hint,
		error,
		required,
		disabled,
		readOnly,
		clearable = false,
		name,
		onChange
	}: Props = $props();

	const items = $derived(
		options.map((option) =>
			typeof option === 'string' ? ({ value: option } as SelectOption<Value>) : option
		)
	);

	/** What Ark walks with the keyboard and matches against typing. Its
	    itemToString is what typeahead searches, so it has to be the text that
	    is actually on screen. */
	const collection = $derived(
		createListCollection({
			items,
			itemToValue: (item) => item.value,
			itemToString: (item) => item.label ?? item.value,
			isItemDisabled: (item) => item.disabled === true
		})
	);

	/** The chosen option, for the icon in the closed control: Ark's ValueText
	    renders the label and nothing else. */
	const chosen = $derived(items.find((item) => item.value === value));

	/** Once one option carries an icon, every row keeps the space for one,
	    so the labels in the panel line up rather than stepping in and out. */
	const gutter = $derived(items.some((item) => item.icon));

	const canClear = $derived(clearable && value !== '' && !disabled && !readOnly);

	/** Whether the list is open: the box shows focus and the label rises
	    while it is, the same as a field being typed in. */
	let open = $state(false);
</script>

<!-- A select is the same outlined box and floating label as an Input, so a
     form reads as one column of fields whatever kind of value each one holds.
     The difference is only what happens on a click: a panel of options
     rather than a caret.

     Ark gives this the parts, the keyboard (arrows, home/end, and typing a
     few letters to jump), and a hidden native select for form submission.
     The closed field is styled in styles/fields.css, the open panel in
     styles/ark.css. -->
<ArkSelect.Root
	class="field-root"
	data-disabled={disabled || undefined}
	{collection}
	{name}
	{required}
	{disabled}
	{readOnly}
	invalid={error !== undefined}
	value={value === '' ? [] : [value]}
	positioning={{ sameWidth: true, gutter: 4 }}
	onOpenChange={(details) => (open = details.open)}
	onValueChange={(details) => {
		value = (details.value[0] ?? '') as Value;
		onChange?.(value);
	}}
>
	<div class="field-box" data-float={value !== '' || undefined} data-open={open || undefined}>
		<ArkSelect.Label><FieldText {label} {icon} {required} /></ArkSelect.Label>

		<ArkSelect.Control>
			<ArkSelect.Trigger>
				{#if chosen?.icon}
					<Icon icon={chosen.icon} />
				{/if}
				<ArkSelect.ValueText {placeholder} />
				<ArkSelect.Indicator>
					<Icon icon={RiArrowDownSLine} />
				</ArkSelect.Indicator>
			</ArkSelect.Trigger>

			{#if canClear}
				<ArkSelect.ClearTrigger aria-label="Clear {label}">
					<Icon icon={RiCloseLine} />
				</ArkSelect.ClearTrigger>
			{/if}
		</ArkSelect.Control>
	</div>

	<!-- The panel is portalled so it is never clipped by a drawer that
	     scrolls or by a table cell that hides its overflow; it is placed
	     against the control by Ark. -->
	<Portal>
		<ArkSelect.Positioner>
			<ArkSelect.Content>
				<ArkSelect.List>
					{#each items as item (item.value)}
						<ArkSelect.Item {item}>
							{#if item.icon}
								<Icon icon={item.icon} />
							{:else if gutter}
								<span data-part="item-gutter"></span>
							{/if}

							<span data-part="item-body">
								<ArkSelect.ItemText>{item.label ?? item.value}</ArkSelect.ItemText>
								{#if item.description}
									<span data-part="item-description">{item.description}</span>
								{/if}
							</span>

							<ArkSelect.ItemIndicator>
								<Icon icon={RiCheckLine} />
							</ArkSelect.ItemIndicator>
						</ArkSelect.Item>
					{:else}
						<p data-part="empty">Nothing to choose from</p>
					{/each}
				</ArkSelect.List>
			</ArkSelect.Content>
		</ArkSelect.Positioner>
	</Portal>

	<ArkSelect.HiddenSelect />

	{#if error}
		<span data-part="error-text">{error}</span>
	{:else if hint}
		<span data-part="helper-text">{hint}</span>
	{/if}
</ArkSelect.Root>
