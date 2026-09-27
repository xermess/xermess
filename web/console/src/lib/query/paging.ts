/** How many rows a list shows before "Show more", and asks the API for at a
    time. The same as the API's own default. */
export const LIST_PAGE_SIZE = 50;

/** A page of a list the API pages by offset, and counts. */
type OffsetPage = { total: number; offset: number };

/**
 * Where the page after this one starts, or undefined on the last. `count` is
 * how many rows this page held; the next begins right after them.
 */
export function nextOffset(page: OffsetPage, count: number): number | undefined {
	const next = page.offset + count;

	return count > 0 && next < page.total ? next : undefined;
}

/**
 * The rows of every page read so far, each once.
 *
 * Pages by offset slide when a row is added or removed between two of them,
 * so the same row can arrive twice; a table keyed by id cannot show it twice.
 */
export function uniqueById<Row extends { id: string }>(rows: Row[]): Row[] {
	const seen = new Set<string>();

	return rows.filter((row) => !seen.has(row.id) && (seen.add(row.id), true));
}

/**
 * How many rows a list that holds every row at once starts by showing.
 *
 * The arguments are what the list is filtered by. They are not used: reading
 * them is what makes a `$derived` of this go back to one page whenever the
 * filter changes, while "Show more" can still write to it in between.
 */
// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function firstPage(...filter: unknown[]): number {
	return LIST_PAGE_SIZE;
}
