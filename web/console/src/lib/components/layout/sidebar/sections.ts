import type { RouteId } from '$app/types';
import type { ComponentType } from 'svelte';
import {
	RiAdminLine,
	RiAppsLine,
	RiBuildingLine,
	RiCodeBoxLine,
	RiDatabase2Line,
	RiFileList3Line,
	RiGitBranchLine,
	RiGroupLine,
	RiLinksLine,
	RiPulseLine,
	RiShareLine,
	RiShieldKeyholeLine,
	RiShieldUserLine,
	RiTranslate2
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

export type SidebarItem = {
	route: Section;
	/** The message key its name is looked up by, not the name: the sidebar
	    has the translator, and this list is read on the server too. */
	key: string;
	icon: ComponentType;
	/** How finished the section is: `preview` shows placeholder data, `soon`
	    is not built yet. Left out, it is the real thing. */
	status?: 'preview' | 'soon';
	/** Whether the administrator may open it. Left out, anyone may. */
	allowed?: (admin: Admin) => boolean;
};

/** A group with no heading sits at the top, on its own. */
export type SidebarGroup = { key?: string; items: SidebarItem[] };

/** The dashboard's sections, in the order the sidebar lists them. The page
    checks permissions again on the server; this only decides what is shown. */
export const sections: SidebarGroup[] = [
	{
		items: [
			{
				route: '/admin/(panel)/dashboard',
				key: 'nav.activity',
				icon: RiPulseLine,
				allowed: (admin) => can(admin, 'activity.read')
			},
			{
				route: '/admin/(panel)/dashboard/logs',
				key: 'nav.logs',
				icon: RiFileList3Line,
				allowed: (admin) => can(admin, 'activity.read')
			}
		]
	},
	{
		key: 'nav.applications',
		items: [
			{
				route: '/admin/(panel)/dashboard/applications',
				key: 'nav.applications',
				icon: RiAppsLine,
				allowed: (admin) => canAnywhere(admin, 'applications.read')
			},
			{
				route: '/admin/(panel)/dashboard/apis',
				key: 'nav.apis',
				icon: RiCodeBoxLine,
				allowed: (admin) => can(admin, 'apis.read')
			},
			{
				route: '/admin/(panel)/dashboard/sso',
				key: 'nav.sso',
				icon: RiLinksLine,
				status: 'soon'
			}
		]
	},
	{
		key: 'nav.authentication',
		items: [
			{
				route: '/admin/(panel)/dashboard/database',
				key: 'nav.database',
				icon: RiDatabase2Line,
				allowed: (admin) => can(admin, 'database.read')
			},
			{
				route: '/admin/(panel)/dashboard/social',
				key: 'nav.social',
				icon: RiShareLine,
				allowed: (admin) => can(admin, 'social.read')
			},
			{
				route: '/admin/(panel)/dashboard/flows',
				key: 'nav.flows',
				icon: RiGitBranchLine,
				allowed: (admin) => can(admin, 'login_flows.read')
			}
		]
	},
	{
		key: 'nav.user_management',
		items: [
			{
				route: '/admin/(panel)/dashboard/users',
				key: 'nav.users',
				icon: RiGroupLine,
				allowed: (admin) => can(admin, 'users.read')
			},
			{
				route: '/admin/(panel)/dashboard/roles',
				key: 'nav.roles',
				icon: RiShieldUserLine,
				allowed: (admin) =>
					canAnywhere(admin, 'users.read') || canAnywhere(admin, 'applications.read')
			}
		]
	},
	{
		key: 'nav.administration',
		items: [
			{
				route: '/admin/(panel)/dashboard/admins',
				key: 'nav.admins',
				icon: RiAdminLine,
				allowed: (admin) => admin.is_super_admin
			},
			{
				route: '/admin/(panel)/dashboard/admin-roles',
				key: 'nav.admin_roles',
				icon: RiShieldKeyholeLine,
				allowed: (admin) => admin.is_super_admin
			}
		]
	},
	{
		key: 'nav.settings',
		items: [
			{
				route: '/admin/(panel)/dashboard/organization',
				key: 'nav.organization',
				icon: RiBuildingLine,
				allowed: (admin) => can(admin, 'organization.read')
			},
			{
				route: '/admin/(panel)/dashboard/languages',
				key: 'nav.languages',
				icon: RiTranslate2,
				allowed: (admin) => can(admin, 'languages.read')
			}
		]
	}
];

/** The groups an administrator sees: what their roles do not allow is left
    out, and a group left empty goes with it. */
export function visibleSections(admin: Admin | undefined): SidebarGroup[] {
	return sections
		.map((group) => ({
			...group,
			items: group.items.filter((item) => !item.allowed || (admin && item.allowed(admin)))
		}))
		.filter((group) => group.items.length > 0);
}
