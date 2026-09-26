<script lang="ts" generics="Row extends { id: string }">
	import type { Snippet } from 'svelte';
	import { RiArrowRightLine, RiInbox2Line } from 'svelte-remixicon';
	import Checkbox from './Checkbox.svelte';
	import Icon from './Icon.svelte';
	import IconButton from './IconButton.svelte';
	import type { Column } from './table';

	type Props = {
		columns: Column[];
		rows: Row[];
		/** Shown in place of the rows when there are none. */
		empty: string;
		/** One row's cells, in column order, as `<td>` elements. */
		row: Snippet<[Row]>;
		/** Given when a row leads somewhere: rows become clickable and a
		    column of arrows is pinned against the right edge. */
		onOpen?: (row: Row) => void;
		/** What that row's arrow is called, for anyone not looking at it. */
		label?: (row: Row) => string;
		/** The ids of the ticked rows. Passing it, together with onSelect,
		    is what puts a column of checkboxes at the front of the table.
		
		    The caller owns the list: the table says what was ticked and shows
		    what it is given, so there is one copy of the truth rather than two
		    that can drift apart. */
		selected?: string[];
		onSelect?: (ids: string[]) => void;
	};

	let { columns, rows, empty, row, onOpen, label, selected, onSelect }: Props = $props();

	const selectable = $derived(selected !== undefined && onSelect !== undefined);
	const ticked = $derived(new Set(selected ?? []));
	const allTicked = $derived(rows.length > 0 && rows.every((item) => ticked.has(item.id)));

	/** Ticked, part-ticked, or not, for the box in the header. */
	const headerState = $derived<boolean | 'indeterminate'>(
		allTicked ? true : rows.some((item) => ticked.has(item.id)) ? 'indeterminate' : false
	);

	function tickAll(checked: boolean) {
		onSelect?.(checked ? rows.map((item) => item.id) : []);
	}

	function tick(id: string, checked: boolean) {
		const ids = selected ?? [];

		onSelect?.(checked ? [...ids, id] : ids.filter((it) => it !== id));
	}

	/** No labels, no header: the activity and session tables are lists of
	    self-evident values and read better without one. */
	const heading = $derived(columns.some((column) => column.label));

	/** True once the table has been scrolled sideways, which is when the
	    pinned column needs an edge to show what is passing under it. */
	let scrolled = $state(false);

	function track(event: Event) {
		scrolled = (event.currentTarget as HTMLElement).scrollLeft > 0;
	}
</script>

<div class="table-card">
	<div class="scroller" class:scrolled onscroll={track}>
		<table>
			{#if heading}
				<thead>
					<tr>
						{#if selectable}
							<th class="tick">
								<Checkbox checked={headerState} onChange={tickAll} title="Select every row" />
							</th>
						{/if}

						{#each columns as column (column.key)}
							<th style:min-width={column.min} class:end={column.align === 'end'}>
								<span class="label">
									{#if column.icon}
										<Icon icon={column.icon} size="0.9375rem" />
									{/if}
									{column.label}
								</span>
							</th>
						{/each}
						{#if onOpen}
							<th class="pin"></th>
						{/if}
					</tr>
				</thead>
			{/if}

			<tbody>
				{#each rows as item (item.id)}
					<tr class:clickable={onOpen !== undefined} onclick={() => onOpen?.(item)}>
						{#if selectable}
							<!-- Ticking a row is not opening it, so the click stops here. -->
							<td class="tick" onclick={(event) => event.stopPropagation()}>
								<Checkbox
									checked={ticked.has(item.id)}
									onChange={(checked) => tick(item.id, checked)}
									title="Select this row"
								/>
							</td>
						{/if}

						{@render row(item)}

						{#if onOpen}
							<td class="pin">
								<!-- The row itself is clickable; this is the same action as
							     something to tab to, and the click it fires on its way up
							     is the one the row handles. -->
								<IconButton
									icon={RiArrowRightLine}
									label={label?.(item) ?? 'Open'}
									size="sm"
									placement="left"
								/>
							</td>
						{/if}
					</tr>
				{:else}
					<tr>
						<td class="empty" colspan={columns.length + (onOpen ? 1 : 0) + (selectable ? 1 : 0)}>
							<span class="empty-state">
								<Icon icon={RiInbox2Line} size="1.5rem" />
								{empty}
							</span>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
</div>

<style>
	/* The table is a card: framed and rounded, so it reads as one object on
	   the page at any width, and scrolls sideways inside its frame when its
	   columns need more room than the page has. */
	.table-card {
		overflow: hidden;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		background: var(--color-surface);
	}

	.scroller {
		overflow-x: auto;
	}

	table {
		width: 100%;
		border-collapse: separate;
		border-spacing: 0;
	}

	/* The cells of a row come from the caller's snippet, so they carry that
	   component's scope rather than this one's: every rule that has to reach
	   a `td` says so with :global, anchored to the table itself. */
	th,
	table :global(td) {
		height: 52px;
		transition: background-color var(--speed-fast);
		padding: 0 var(--space-3);
		border-bottom: 1px solid var(--color-border);
		font-size: var(--text-base);
		font-weight: 400;
		text-align: left;
		vertical-align: middle;
		white-space: nowrap;
	}

	table :global(tbody tr:last-child td) {
		border-bottom: none;
	}

	/* The header is a tinted band of small labels: it names the columns
	   without competing with what is in them. */
	th {
		height: 40px;
		background: var(--color-surface-alt);
		color: var(--color-text-hint);
		font-size: var(--text-xs);
		font-weight: 600;
		letter-spacing: 0.02em;
	}

	.label {
		display: flex;
		align-items: center;
		gap: 6px;
	}

	.label :global(svg) {
		opacity: 0.8;
	}

	th.end .label {
		justify-content: flex-end;
	}

	th.end,
	table :global(td.end) {
		text-align: right;
	}

	/* The first and last cells are inset a little further than the gaps
	   between columns, so the text does not crowd the frame. */
	th:first-child,
	table :global(td:first-child) {
		padding-left: var(--space-4);
	}

	th:last-child,
	table :global(td:last-child) {
		padding-right: var(--space-4);
	}

	tr.clickable {
		cursor: pointer;
	}

	/* A whole row changing colour is a lot of colour, so the wash is faint:
	   enough to follow the pointer across a wide table, not enough to shout.
	   The pinned cells take it too, or they would stay white over it. */
	tr.clickable:hover :global(td) {
		background: var(--row-hover);
	}

	/* The checkboxes hold the left edge the way the arrows hold the right,
	   so a wide table can still be ticked while it is scrolled. */
	.tick {
		position: sticky;
		left: 0;
		z-index: 1;
		width: 1%;
		padding-right: 0;
		/* The checkbox is the cell's only content; without a line box around
		   it, it sits on the row's centre rather than on a text baseline. */
		line-height: 0;
		background: var(--color-surface);
		transition: background-color var(--speed-fast);
	}

	th.tick {
		background: var(--color-surface-alt);
	}

	.tick::after {
		position: absolute;
		top: 0;
		right: 0;
		bottom: 0;
		width: 1px;
		background: var(--color-border);
		opacity: 0;
		transition: opacity var(--speed);
		content: '';
	}

	.scrolled .tick::after {
		opacity: 1;
	}

	.scrolled .tick {
		box-shadow: 8px 0 8px -8px rgb(0 0 0 / 25%);
	}

	/* The arrow stays against the right edge while the rest of the row
	   scrolls under it, so the way into a record is always in view. */
	.pin {
		position: sticky;
		right: 0;
		z-index: 1;
		width: 1%;
		padding-left: var(--space-2);
		background: var(--color-surface);
		transition: background-color var(--speed-fast);
	}

	th.pin {
		background: var(--color-surface-alt);
	}

	/* An edge on the pinned column, drawn only once something is passing
	   under it. The line reads in both themes where a shadow alone would
	   disappear against the dark one. */
	.pin::before {
		position: absolute;
		top: 0;
		bottom: 0;
		left: 0;
		width: 1px;
		background: var(--color-border);
		opacity: 0;
		transition: opacity var(--speed);
		content: '';
	}

	.scrolled .pin {
		box-shadow: -8px 0 8px -8px rgb(0 0 0 / 25%);
	}

	.scrolled .pin::before {
		opacity: 1;
	}

	/* The arrow is quiet until the row is under the pointer. */
	.pin :global(.control) {
		margin-left: auto;
		color: var(--color-text-disabled);
	}

	tr.clickable:hover .pin :global(.control) {
		color: var(--color-text);
	}

	.empty {
		height: auto;
		padding: var(--space-6) var(--space-4);
		white-space: normal;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-2);
		color: var(--color-text-hint);
		text-align: center;
	}

	.empty-state :global(svg) {
		color: var(--color-text-disabled);
	}
</style>
