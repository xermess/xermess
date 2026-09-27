import type { RouteId } from '$app/types';
import type { ComponentType } from 'svelte';
import {
	RiAdminLine,
	RiAppsLine,
	RiBuildingLine,
	RiCodeBoxLine,
	RiComputerLine,
	RiDatabase2Line,
	RiFileList3Line,
	RiGitBranchLine,
	RiGroupLine,
	RiKeyLine,
	RiLinksLine,
	RiMailLine,
	RiPulseLine,
	RiSettings3Line,
	RiShareLine,
	RiShieldKeyholeLine,
	RiShieldUserLine,
	RiTeamLine
} from 'svelte-remixicon';
import type { Admin } from '$lib/api';
import { can, canAnywhere } from '$lib/permissions';

/** A page the sidebar can lead to: one of the dashboard's, with no parameters
    to fill in. It is narrowed to the dashboard on purpose — every other route
    in the app would otherwise join the union, and past 25 members TypeScript
    can no longer match it against resolve()'s. */
export type Section = Extract<
	Exclude<RouteId, `${string}[${string}`>,
	`/admin/(panel)/dashboard${string}`
>;

/** A row that leads somewhere. */
export type SidebarItem = {
	route: Section;
	/** What the sidebar calls it. */
	label: string;
	icon: ComponentType;
	/** Whether the administrator may open it. Left out, anyone may. */
	allowed?: (admin: Admin) => boolean;
};

/** A row that opens to reveal its pages. It leads nowhere itself: an
    administrator who cannot see any of its children never sees it either. */
export type SidebarBranch = {
	/** A name the cookie can hold, so which branches are open survives a
	    reload. It is not the label: renaming a section should not fold it. */
	id: string;
	label: string;
	icon: ComponentType;
	items: SidebarItem[];
};

/** The sidebar in order: the pages that answer "what is happening" on their
    own in an Overview group at the top, and everything else under the
    subject it belongs to.
    A branch is a subject, not a bucket — which is why One-time codes sits
    under Authentication, where it is read, rather than under Settings, where
    it merely lives. */
export const overview: SidebarItem[] = [
	{
		route: '/admin/(panel)/dashboard',
		label: 'Activity',
		icon: RiPulseLine,
		allowed: (admin) => can(admin, 'activity.read')
	},
	{
		route: '/admin/(panel)/dashboard/logs',
		label: 'Logs',
		icon: RiFileList3Line,
		allowed: (admin) => can(admin, 'activity.read')
	}
];

export const branches: SidebarBranch[] = [
	{
		id: 'applications',
		label: 'Applications',
		icon: RiAppsLine,
		items: [
			{
				route: '/admin/(panel)/dashboard/applications',
				label: 'Applications',
				icon: RiAppsLine,
				allowed: (admin) => canAnywhere(admin, 'applications.read')
			},
			{
				route: '/admin/(panel)/dashboard/apis',
				label: 'APIs',
				icon: RiCodeBoxLine,
				allowed: (admin) => can(admin, 'apis.read')
			}
		]
	},
	{
		id: 'authentication',
		label: 'Authentication',
		icon: RiShieldKeyholeLine,
		items: [
			{
				route: '/admin/(panel)/dashboard/flows',
				label: 'Login flows',
				icon: RiGitBranchLine,
				allowed: (admin) => can(admin, 'login_flows.read')
			},
			{
				route: '/admin/(panel)/dashboard/social',
				label: 'Social',
				icon: RiShareLine,
				allowed: (admin) => can(admin, 'social.read')
			},
			{
				route: '/admin/(panel)/dashboard/sso',
				label: 'SSO integrations',
				icon: RiLinksLine,
				allowed: (admin) => can(admin, 'sso.read')
			},
			{
				route: '/admin/(panel)/dashboard/otp',
				label: 'One-time codes',
				icon: RiKeyLine,
				allowed: (admin) => admin.is_super_admin
			}
		]
	},
	{
		id: 'users',
		label: 'Users',
		icon: RiTeamLine,
		items: [
			{
				route: '/admin/(panel)/dashboard/users',
				label: 'Users',
				icon: RiGroupLine,
				allowed: (admin) => can(admin, 'users.read')
			},
			{
				route: '/admin/(panel)/dashboard/sessions',
				label: 'Sessions',
				icon: RiComputerLine,
				allowed: (admin) => can(admin, 'users.read')
			},
			{
				route: '/admin/(panel)/dashboard/roles',
				label: 'Roles',
				icon: RiShieldUserLine,
				allowed: (admin) =>
					canAnywhere(admin, 'users.read') || canAnywhere(admin, 'applications.read')
			}
		]
	},
	{
		id: 'administration',
		label: 'Administration',
		icon: RiAdminLine,
		items: [
			{
				route: '/admin/(panel)/dashboard/admins',
				label: 'Administrators',
				icon: RiAdminLine,
				allowed: (admin) => admin.is_super_admin
			},
			{
				route: '/admin/(panel)/dashboard/admin-roles',
				label: 'Admin roles',
				icon: RiShieldKeyholeLine,
				allowed: (admin) => admin.is_super_admin
			}
		]
	},
	{
		id: 'settings',
		label: 'Settings',
		icon: RiSettings3Line,
		items: [
			{
				route: '/admin/(panel)/dashboard/settings',
				label: 'General',
				icon: RiBuildingLine,
				allowed: (admin) => can(admin, 'organization.read') || can(admin, 'languages.read')
			},
			{
				route: '/admin/(panel)/dashboard/mail',
				label: 'Mail',
				icon: RiMailLine,
				allowed: (admin) => admin.is_super_admin
			},
			{
				route: '/admin/(panel)/dashboard/cache',
				label: 'Cache',
				icon: RiDatabase2Line,
				allowed: (admin) => admin.is_super_admin
			}
		]
	}
];

/** Whether a page is the one being read. `id` is the route id of the page being
    read — `/admin/(panel)/dashboard/users` — and the route is one of the same
    shape, so the two are compared as they are rather than through `resolve`.
    That answers a path, and SvelteKit has answered those relatively since 2.0,
    so a resolved path never equals the pathname and no row is ever marked.
    Activity is the dashboard's own page, so it only matches exactly; the
    others also match anything below them. The sidebar and the header ask the
    same question the same way. */
export function isCurrentSection(route: Section, id: string | null | undefined): boolean {
	return route === '/admin/(panel)/dashboard'
		? id === route
		: id === route || Boolean(id?.startsWith(`${route}/`));
}

/** What an administrator sees: the pages their roles allow, and a branch only
    where it still has one. The page checks the permission again on the
    server; this only decides what is shown. */
export function visibleOverview(admin: Admin | undefined): SidebarItem[] {
	return overview.filter((item) => !item.allowed || (admin && item.allowed(admin)));
}

export function visibleBranches(admin: Admin | undefined): SidebarBranch[] {
	return branches
		.map((branch) => ({
			...branch,
			items: branch.items.filter((item) => !item.allowed || (admin && item.allowed(admin)))
		}))
		.filter((branch) => branch.items.length > 0);
}

/** Every page the sidebar leads to, in the order it lists them — for the
    command palette, which is one flat list however the column is grouped. */
export function allSections(admin: Admin | undefined): { group: string; items: SidebarItem[] }[] {
	const overviewItems = visibleOverview(admin);

	return [
		...(overviewItems.length > 0 ? [{ group: 'Dashboard', items: overviewItems }] : []),
		...visibleBranches(admin).map((branch) => ({ group: branch.label, items: branch.items }))
	];
}
