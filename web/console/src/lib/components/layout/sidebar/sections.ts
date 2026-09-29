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
	RiShareLine,
	RiShieldKeyholeLine,
	RiShieldUserLine
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

/** A named group of pages. Its heading folds it away, and which are folded
    is remembered; an administrator who cannot see any of its pages never
    sees the group either. */
export type SidebarGroup = {
	/** A name the cookie can hold, so which groups are folded survives a
	    reload. It is not the label: renaming a group should not unfold it. */
	id: string;
	label: string;
	items: SidebarItem[];
};

/** The sidebar in order: what is happening first, then each subject the
    panel manages, then the installation's own settings. Every group is the
    same kind of thing — a heading and its pages — so the column reads as one
    list, top to bottom.

    A group is a subject, not a bucket — which is why One-time codes sits
    under Authentication, where it is read, rather than under Settings, where
    it merely lives. */
const groups: SidebarGroup[] = [
	{
		id: 'overview',
		label: 'Overview',
		items: [
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
		]
	},
	{
		id: 'applications',
		label: 'Applications',
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
		items: [
			{
				route: '/admin/(panel)/dashboard/flows',
				label: 'Login flows',
				icon: RiGitBranchLine,
				allowed: (admin) => can(admin, 'login_flows.read')
			},
			{
				route: '/admin/(panel)/dashboard/social',
				label: 'Social sign-in',
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
		label: 'User management',
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

/** What an administrator sees: the pages their roles allow, and a group only
    where it still has one. The page checks the permission again on the
    server; this only decides what is shown. */
export function visibleGroups(admin: Admin | undefined): SidebarGroup[] {
	return groups
		.map((group) => ({
			...group,
			items: group.items.filter((item) => !item.allowed || (admin && item.allowed(admin)))
		}))
		.filter((group) => group.items.length > 0);
}
