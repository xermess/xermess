import { queryOptions } from '@tanstack/svelte-query';

import { flowsApi, type LoginFlow, type LoginStepSpec } from '$lib/api';
import { keys } from './keys';

/** The login flows and the steps one can be made of, seeded with what the
    server rendered. */
export function loginFlowsOptions(initial: { flows: LoginFlow[]; step_kinds: LoginStepSpec[] }) {
	return queryOptions({
		queryKey: keys.flows.list,
		queryFn: () => flowsApi.list(),
		initialData: initial
	});
}

/** The flows, for the picker on an application. It is not seeded: the
    applications page does not fetch flows to render, and an administrator who
    may not read them gets nothing rather than an error — the picker is left
    out, and the application keeps whichever flow it has. */
export function loginFlowChoicesOptions() {
	return queryOptions({
		queryKey: keys.flows.list,
		queryFn: () => flowsApi.list(),
		retry: false
	});
}
