import { queryOptions } from '@tanstack/svelte-query';

import { languagesApi, type LanguageList, type LocaleApp } from '$lib/api';
import { keys } from './keys';

/** The languages and what each covers, seeded with what the server rendered. */
export function languagesOptions(initial: LanguageList) {
	return queryOptions({
		queryKey: keys.languages.list,
		queryFn: () => languagesApi.list(),
		initialData: initial
	});
}

/** One language's text for one app. It is only asked for when its tab is
    opened, and never goes stale on its own: nobody else is typing into it. */
export function translationOptions(code: string, app: LocaleApp) {
	return queryOptions({
		queryKey: keys.languages.translation(code, app),
		queryFn: () => languagesApi.translation(code, app),
		staleTime: Infinity
	});
}
