import type { LoginFlow, LoginStepSpec } from '$lib/api';
import { apiGet, requirePermission } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * The login flows this installation signs people in with, and the steps one
 * can be made of.
 *
 * There are as many flows as an administrator has written — a handful — so
 * the endpoint answers with all of them and the page narrows the list itself,
 * as the Social page does. The search box lives in the URL, so the back
 * button walks through it and a filtered list can be linked to.
 */
export const load: PageServerLoad = async ({ fetch, parent, url }) => {
	requirePermission((await parent()).admin, 'login_flows.read');

	const { flows, step_kinds } = await apiGet<{
		flows: LoginFlow[];
		step_kinds: LoginStepSpec[];
	}>('/admin/login-flows', fetch);

	return {
		flows,
		stepKinds: step_kinds,
		search: url.searchParams.get('search')?.trim() ?? ''
	};
};
