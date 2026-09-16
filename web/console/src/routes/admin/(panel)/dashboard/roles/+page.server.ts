import { error } from '@sveltejs/kit';
import type { API, Application, ApplicationPage, RolePage } from '$lib/api';
import { can, canAnywhere } from '$lib/permissions';
import { APPLICATION_CHOICES_LIMIT, ROLE_CHOICES_LIMIT } from '$lib/query';
import { apiGet } from '$lib/server/api';
import type { PageServerLoad } from './$types';

/**
 * Roles, on two tabs: the global roles, and every application's roles. The
 * tab, the search and the filter live in the URL, so a link opens the same
 * list; an application's panel links here with `application` set, which
 * narrows the second tab to that application's roles.
 *
 * Every role the administrator can see is loaded as well as the page: the tab
 * counts come from it, and a role can include roles of other scopes.
 */
export const load: PageServerLoad = async ({ fetch, parent, url }) => {
	const { admin } = await parent();

	if (!canAnywhere(admin, 'users.read') && !canAnywhere(admin, 'applications.read')) {
		error(403, 'Your roles do not allow you to see this page.');
	}

	const seesApplications = canAnywhere(admin, 'applications.read');

	const tab =
		url.searchParams.get('tab') === 'application' && seesApplications ? 'application' : 'global';
	const search = url.searchParams.get('search')?.trim() ?? '';
	const isDefault = url.searchParams.get('default') ?? '';

	// The API scopes a role can grant, for an administrator who may see APIs;
	// null otherwise, so the role panel leaves grants alone.
	const apis = can(admin, 'apis.read')
		? (await apiGet<{ apis: API[] }>('/admin/apis', fetch)).apis
		: null;

	const [applications, choices] = await Promise.all([
		seesApplications
			? apiGet<ApplicationPage>(
					`/admin/applications?limit=${APPLICATION_CHOICES_LIMIT}`,
					fetch
				).then((it) => it.applications)
			: Promise.resolve<Application[]>([]),
		apiGet<RolePage>(`/admin/user-roles?limit=${ROLE_CHOICES_LIMIT}`, fetch)
	]);

	const named = url.searchParams.get('application') ?? '';
	const application =
		tab === 'application' && applications.some((app) => app.id === named) ? named : '';

	const query = new URLSearchParams({ scope: tab });
	if (application) query.set('application', application);
	if (search) query.set('search', search);
	if (isDefault === 'true' || isDefault === 'false') query.set('default', isDefault);

	const page = await apiGet<RolePage>(`/admin/user-roles?${query}`, fetch);

	return {
		page,
		applications,
		apis,
		choices: choices.roles,
		tab: tab as 'global' | 'application',
		application,
		search,
		isDefault
	};
};
