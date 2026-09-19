import type { RoleListParams } from './roles';

/**
 * The names the cache knows things by.
 *
 * Keys are built here rather than written out at each call, so invalidating
 * "the users" after a write cannot miss a spelling. A key is a list read
 * left to right: everything under `users.all` goes when a user changes.
 */
export const keys = {
	users: {
		all: ['users'] as const,
		/** One page of the list, as the search box and filter describe it. */
		list: (params: { search: string; verified: string; role: string }) =>
			['users', 'list', params.search, params.verified, params.role] as const,
		/** The fields a user record is made of. */
		fields: ['users', 'fields'] as const,
		/** Every role one user holds. */
		mappings: (id: string) => ['users', 'mappings', id] as const
	},

	apis: {
		all: ['apis'] as const,
		list: (search: string) => ['apis', 'list', search] as const,
		one: (id: string) => ['apis', 'one', id] as const,
		/** The applications that may, or may not, use one API. */
		applications: (id: string) => ['apis', 'applications', id] as const,
		logs: (id: string) => ['apis', 'logs', id] as const
	},

	applications: {
		/** What one application may do with each API. */
		access: (id: string) => ['applications', 'access', id] as const,
		all: ['applications'] as const,
		/** One page of the list, as the search box and filters describe it. */
		list: (params: { search: string; type: string }) =>
			['applications', 'list', params.search, params.type] as const,
		/** Every application the administrator can see, for pickers. */
		choices: ['applications', 'choices'] as const
	},

	roles: {
		all: ['roles'] as const,
		/** One tab of roles, as its filters describe it. */
		list: (params: RoleListParams) =>
			['roles', 'list', params.scope, params.application, params.search, params.isDefault] as const,
		/** Every role, for the pickers that offer them as choices. */
		choices: ['roles', 'choices'] as const
	},

	admins: {
		all: ['admins'] as const,
		/** One page of the list, as the search box and filters describe it. */
		list: (params: { search: string; status: string; role: string }) =>
			['admins', 'list', params.search, params.status, params.role] as const,
		/** Every admin role. */
		roles: (search: string) => ['admins', 'roles', search] as const,
		/** The admin permission catalog, which never changes while running. */
		permissions: ['admins', 'permissions'] as const
	},

	organization: {
		/** There is one organisation, so this key is the whole of it. */
		settings: ['organization', 'settings'] as const
	},

	social: {
		all: ['social'] as const,
		/** Every configured provider, and the kinds one may be. */
		providers: ['social', 'providers'] as const
	},

	flows: {
		all: ['flows'] as const,
		/** Every login flow, and the steps one can be made of. */
		list: ['flows', 'list'] as const
	},

	sso: {
		all: ['sso'] as const,
		/** Every connection. */
		list: ['sso', 'list'] as const
	},

	languages: {
		all: ['languages'] as const,
		/** Every language, with how much of each app it covers. */
		list: ['languages', 'list'] as const,
		/** Every language's text, and one language's text for one app. */
		translations: ['languages', 'translation'] as const,
		translation: (code: string, app: string) => ['languages', 'translation', code, app] as const
	},

	sessions: {
		all: ['sessions'] as const,
		/** The list, as the search box and the user filter describe it. */
		list: (params: { search: string; user: string }) =>
			['sessions', 'list', params.search, params.user] as const
	},

	admin: {
		/** The dashboard's counts and recent activity, which a change to
		    anything it counts invalidates. */
		overview: ['admin', 'overview'] as const
	}
} as const;
