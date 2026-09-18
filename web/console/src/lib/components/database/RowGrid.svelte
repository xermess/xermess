<script lang="ts">
	import { RiEyeOffLine, RiKey2Line } from 'svelte-remixicon';
	import type { DatabaseColumn } from '$lib/api';
	import { Icon, Tooltip } from '$lib/components/ui';
	import { cell } from './format';

	type Props = {
		columns: DatabaseColumn[];
		rows: Record<string, unknown>[];
		/** Shown in place of the rows when there are none. */
		empty: string;
	};

	let { columns, rows, empty }: Props = $props();
</script>

<!-- The table's own columns, in the order the database holds them. It is not
     the DataTable: that one is built for records with an id and a row snippet
     the page writes, and a table here may be a join table with neither. What
     it shares is the look — the same heights, the same rules, running to the
     edges of the page. -->
<div class="scroller">
	<table>
		<thead>
			<tr>
				{#each columns as column (column.name)}
					<th>
						<span class="label">
							{#if column.primary_key}
								<Tooltip label="The column rows are identified by">
									{#snippet children(trigger)}
										<span class="mark" {...trigger()}>
											<Icon icon={RiKey2Line} size="0.8rem" />
										</span>
									{/snippet}
								</Tooltip>
							{:else if column.hidden}
								<Tooltip label="A password, a key or a token: never read">
									{#snippet children(trigger)}
										<span class="mark" {...trigger()}>
											<Icon icon={RiEyeOffLine} size="0.8rem" />
										</span>
									{/snippet}
								</Tooltip>
							{/if}
							{column.name}
							<span class="type">{column.type}</span>
						</span>
					</th>
				{/each}
			</tr>
		</thead>

		<tbody>
			{#each rows as row, index (index)}
				<tr>
					{#each columns as column (column.name)}
						{@const value = cell(row[column.name], column)}
						<td class={value.kind}>{value.text}</td>
					{/each}
				</tr>
			{:else}
				<tr class="none">
					<td colspan={Math.max(columns.length, 1)}>{empty}</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>

<style>
	.scroller {
		overflow-x: auto;
		border-top: 1px solid var(--color-border);
	}

	table {
		width: 100%;
		border-collapse: separate;
		border-spacing: 0;
	}

	th,
	td {
		height: 45px;
		padding: 0 var(--space-3);
		border-bottom: 1px solid var(--color-border);
		font-size: var(--text-base);
		font-weight: 400;
		text-align: left;
		white-space: nowrap;
	}

	th {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		font-weight: 700;
	}

	th:first-child,
	td:first-child {
		padding-left: var(--page-gutter);
	}

	th:last-child,
	td:last-child {
		padding-right: var(--page-gutter);
	}

	.label {
		display: flex;
		align-items: center;
		gap: var(--space-2);
	}

	/* The column's type beside its name, the way PocketBase marks a field. */
	.type {
		color: var(--color-text-disabled);
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		font-weight: 400;
	}

	.mark {
		display: inline-flex;
		color: var(--color-text-hint);
	}

	/* Values are read against each other down a column, so they are set in
	   the monospace face whatever they hold. */
	td {
		font-family: var(--font-mono);
		font-size: var(--text-sm);
	}

	/* A null, an empty string and a column that is never read are each drawn
	   as what they are rather than as text that could be a value. */
	td.empty,
	td.hidden {
		color: var(--color-text-disabled);
		font-style: italic;
	}

	tr:hover td {
		background: var(--row-hover);
	}

	tr.none td {
		color: var(--color-text-hint);
		font-family: var(--font-sans);
		font-size: var(--text-base);
		font-style: normal;
		text-align: center;
	}

	tr.none:hover td {
		background: transparent;
	}
</style>
