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
 * Rows of every page loaded so far, deduplicated by id (offset pages can repeat a row after an
 * insert or delete).
 */
export function uniqueById<Row extends { id: string }>(rows: Row[]): Row[] {
	const seen = new Set<string>();

	return rows.filter((row) => !seen.has(row.id) && (seen.add(row.id), true));
}

/**
 * How many rows a fully-loaded list shows at first. The filter arguments are unused but read,
 * so a `$derived` of this resets to one page when the filter changes.
 */
// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function firstPage(...filter: unknown[]): number {
	return LIST_PAGE_SIZE;
}
