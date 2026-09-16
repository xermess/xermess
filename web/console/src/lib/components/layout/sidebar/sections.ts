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
	label: string;
	icon: ComponentType;
	/** How finished the section is: `preview` shows placeholder data, `soon`
	    is not built yet. Left out, it is the real thing. */
	status?: 'preview' | 'soon';
	/** Whether the administrator may open it. Left out, anyone may. */
	allowed?: (admin: Admin) => boolean;
};

/** A group with no label sits at the top, on its own. */
export type SidebarGroup = { label?: string; items: SidebarItem[] };

/** The dashboard's sections, in the order the sidebar lists them. The page
    checks permissions again on the server; this only decides what is shown. */
export const sections: SidebarGroup[] = [
	{
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
			},
			{
				route: '/admin/(panel)/dashboard/sso',
				label: 'SSO integrations',
				icon: RiLinksLine,
				status: 'soon'
			}
		]
	},
	{
		label: 'Authentication',
		items: [
			{
				route: '/admin/(panel)/dashboard/database',
				label: 'Database',
				icon: RiDatabase2Line,
				status: 'preview'
			},
			{
				route: '/admin/(panel)/dashboard/social',
				label: 'Social',
				icon: RiShareLine,
				status: 'preview'
			},
			{
				route: '/admin/(panel)/dashboard/flows',
				label: 'Login flows',
				icon: RiGitBranchLine,
				status: 'preview'
			}
		]
	},
	{
		label: 'User management',
		items: [
			{
				route: '/admin/(panel)/dashboard/users',
				label: 'Users',
				icon: RiGroupLine,
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
		label: 'Settings',
		items: [
			{
				route: '/admin/(panel)/dashboard/organization',
				label: 'Organization',
				icon: RiBuildingLine,
				allowed: (admin) => can(admin, 'organization.read')
			},
			{
				route: '/admin/(panel)/dashboard/languages',
				label: 'Languages',
				icon: RiTranslate2,
				status: 'preview'
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
