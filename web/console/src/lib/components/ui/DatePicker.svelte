<script lang="ts">
	import type { ComponentType } from 'svelte';
	import { DatePicker as ArkDatePicker } from '@ark-ui/svelte/date-picker';
	import { Portal } from '@ark-ui/svelte/portal';
	import {
		RiArrowLeftSLine,
		RiArrowRightSLine,
		RiCalendarLine,
		RiCloseLine
	} from 'svelte-remixicon';
	import { formatDay, parseDay, timeZone, toDate, toIso } from './date';
	import FieldText from './FieldText.svelte';
	import Icon from './Icon.svelte';

	type Props = {
		label: string;
		/** The day, as an ISO date — "2026-03-12" — or an empty string for none:
		    the same value a native date input has, so a form that had one takes
		    this without changing what it sends. */
		value: string;
		/** The first and last days that can be chosen, as ISO dates. */
		min?: string;
		max?: string;
		icon?: ComponentType;
		hint?: string;
		error?: string;
		required?: boolean;
		disabled?: boolean;
		readOnly?: boolean;
		/** Offers a button that empties the field. */
		clearable?: boolean;
		/** A toolbar's filter: a control's height, the label beside the
		    value. See Field. */
		compact?: boolean;
		name?: string;
		id?: string;
		onChange?: (value: string) => void;
	};

	let {
		label,
		value = $bindable(''),
		min,
		max,
		icon,
		hint,
		error,
		required,
		disabled,
		readOnly,
		clearable = false,
		compact = false,
		name,
		id,
		onChange
	}: Props = $props();

	const selected = $derived(toDate(value));
	const canClear = $derived(clearable && value !== '' && !disabled && !readOnly);

	/** Whether the calendar is open: the box shows focus and the label rises
	    while it is, as for a select. */
	let open = $state(false);
</script>

<!--
	A date field: typed or picked from a calendar that opens on the chosen month. Ark provides
	keyboard handling and parsing; styles are in styles/fields.css and styles/ark.css.
-->
<ArkDatePicker.Root
	class="field-root"
	data-invalid={error !== undefined || undefined}
	{id}
	{name}
	{required}
	{disabled}
	{readOnly}
	invalid={error !== undefined}
	locale="en-GB"
	{timeZone}
	startOfWeek={1}
	fixedWeeks
	value={selected ? [selected] : []}
	min={toDate(min)}
	max={toDate(max)}
	format={formatDay}
	parse={parseDay}
	placeholder="dd/mm/yyyy"
	positioning={{ placement: 'bottom-start', gutter: 4 }}
	onOpenChange={(details) => (open = details.open)}
	onValueChange={(details) => {
		const [day] = details.value;
		const next = day ? toIso(day) : '';
		if (next === value) return;

		value = next;
		onChange?.(next);
	}}
>
	<div
		class="field-box"
		data-float={value !== '' || undefined}
		data-open={open || undefined}
		data-compact={compact || undefined}
	>
		<ArkDatePicker.Label><FieldText {label} {icon} {required} /></ArkDatePicker.Label>

		<ArkDatePicker.Control>
			<ArkDatePicker.Input />

			{#if canClear}
				<ArkDatePicker.ClearTrigger aria-label="Clear {label}">
					<Icon icon={RiCloseLine} />
				</ArkDatePicker.ClearTrigger>
			{/if}

			<ArkDatePicker.Trigger aria-label="Choose {label}">
				<Icon icon={RiCalendarLine} />
			</ArkDatePicker.Trigger>
		</ArkDatePicker.Control>
	</div>

	<!-- Portalled, as a select's list is, so a dialog that scrolls or a toolbar
	     that hides its overflow never cuts the calendar off. -->
	<Portal>
		<ArkDatePicker.Positioner>
			<ArkDatePicker.Content>
				<ArkDatePicker.View view="day">
					<ArkDatePicker.Context>
						{#snippet render(context)}
							{@const api = context()}
							{@render header()}
							<ArkDatePicker.Table>
								<ArkDatePicker.TableHead>
									<ArkDatePicker.TableRow>
										{#each api.weekDays as day (day.long)}
											<ArkDatePicker.TableHeader abbr={day.long}>
												{day.narrow}
											</ArkDatePicker.TableHeader>
										{/each}
									</ArkDatePicker.TableRow>
								</ArkDatePicker.TableHead>
								<ArkDatePicker.TableBody>
									{#each api.weeks as week, index (index)}
										<ArkDatePicker.TableRow>
											{#each week as day (day.toString())}
												<ArkDatePicker.TableCell value={day}>
													<ArkDatePicker.TableCellTrigger>{day.day}</ArkDatePicker.TableCellTrigger>
												</ArkDatePicker.TableCell>
											{/each}
										</ArkDatePicker.TableRow>
									{/each}
								</ArkDatePicker.TableBody>
							</ArkDatePicker.Table>
						{/snippet}
					</ArkDatePicker.Context>
				</ArkDatePicker.View>

				<ArkDatePicker.View view="month">
					<ArkDatePicker.Context>
						{#snippet render(context)}
							{@const api = context()}
							{@render header()}
							<ArkDatePicker.Table>
								<ArkDatePicker.TableBody>
									{#each api.getMonthsGrid( { columns: 3, format: 'short' } ) as months, index (index)}
										<ArkDatePicker.TableRow>
											{#each months as month (month.value)}
												<ArkDatePicker.TableCell value={month.value}>
													<ArkDatePicker.TableCellTrigger
														>{month.label}</ArkDatePicker.TableCellTrigger
													>
												</ArkDatePicker.TableCell>
											{/each}
										</ArkDatePicker.TableRow>
									{/each}
								</ArkDatePicker.TableBody>
							</ArkDatePicker.Table>
						{/snippet}
					</ArkDatePicker.Context>
				</ArkDatePicker.View>

				<ArkDatePicker.View view="year">
					<ArkDatePicker.Context>
						{#snippet render(context)}
							{@const api = context()}
							{@render header()}
							<ArkDatePicker.Table>
								<ArkDatePicker.TableBody>
									{#each api.getYearsGrid({ columns: 3 }) as years, index (index)}
										<ArkDatePicker.TableRow>
											{#each years as year (year.value)}
												<ArkDatePicker.TableCell value={year.value}>
													<ArkDatePicker.TableCellTrigger
														>{year.label}</ArkDatePicker.TableCellTrigger
													>
												</ArkDatePicker.TableCell>
											{/each}
										</ArkDatePicker.TableRow>
									{/each}
								</ArkDatePicker.TableBody>
							</ArkDatePicker.Table>
						{/snippet}
					</ArkDatePicker.Context>
				</ArkDatePicker.View>
			</ArkDatePicker.Content>
		</ArkDatePicker.Positioner>
	</Portal>

	{#if error}
		<span data-part="error-text">{error}</span>
	{:else if hint}
		<span data-part="helper-text">{hint}</span>
	{/if}
</ArkDatePicker.Root>

<!-- The row over every view: back, what is on show — which steps out to the
     next view up when pressed — and forward. Each view renders it inside its
     own View, so the buttons move by that view's step: a month, a year, a
     decade. -->
{#snippet header()}
	<ArkDatePicker.ViewControl>
		<ArkDatePicker.PrevTrigger aria-label="Previous">
			<Icon icon={RiArrowLeftSLine} />
		</ArkDatePicker.PrevTrigger>
		<ArkDatePicker.ViewTrigger>
			<ArkDatePicker.RangeText />
		</ArkDatePicker.ViewTrigger>
		<ArkDatePicker.NextTrigger aria-label="Next">
			<Icon icon={RiArrowRightSLine} />
		</ArkDatePicker.NextTrigger>
	</ArkDatePicker.ViewControl>
{/snippet}
