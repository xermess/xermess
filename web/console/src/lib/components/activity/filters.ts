import type { LogFilter } from '$lib/api';
import { actionsIn, categoryLabels, refusedActions, type Category } from './actions';

/** What the logs page can be narrowed to by kind: a category of the action
    catalog, or the sign-ins that were turned away. */
export type Kind = Category | 'refused';

export const kindLabels: Record<Kind, string> = {
	...categoryLabels,
	refused: 'Refused sign-ins'
};

/** The logs page's filters as its address holds them, and the filter the
    server is asked with. */
export type LogView = {
	q: string;
	kind: Kind | '';
	actor: string;
	from: string;
	to: string;
	filter: LogFilter;
};

const isKind = (value: string): value is Kind => value in kindLabels;

/** Reads the address. Anything it does not recognise is left out rather than
    refused: an old link still opens the log. */
export function filterFrom(params: URLSearchParams): LogView {
	const q = params.get('q')?.trim() ?? '';
	const raw = params.get('kind') ?? '';
	const kind = isKind(raw) ? raw : '';
	const actor = params.get('actor')?.trim() ?? '';
	const day = (value: string | null) => (value && /^\d{4}-\d{2}-\d{2}$/.test(value) ? value : '');
	const from = day(params.get('from'));
	const to = day(params.get('to'));

	const actions = kind === 'refused' ? refusedActions : kind ? actionsIn(kind) : [];

	return {
		q,
		kind,
		actor,
		from,
		to,
		filter: {
			q: q || undefined,
			actions: actions.length > 0 ? actions : undefined,
			actor: actor || undefined,
			from: from || undefined,
			to: to || undefined
		}
	};
}

/** The address of the logs page narrowed to these filters, for a link from
    elsewhere — the dashboard's chart, its sign-ins, its busiest people. */
export function logsHref(base: string, view: Partial<Omit<LogView, 'filter'>>): string {
	const params = new URLSearchParams();
	for (const [key, value] of Object.entries(view)) {
		if (value) params.set(key, value);
	}

	const query = params.toString();
	return query ? `${base}?${query}` : base;
}
