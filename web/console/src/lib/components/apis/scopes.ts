import type { API, APIInput, SigningAlgorithm } from '$lib/api';
import type { SelectOption } from '$lib/components/ui';

/** One scope being edited, with a key of its own so a row keeps its identity
    while its name is typed. `id` is set for a scope already stored, which
    keeps every application's allowance and role's grant of it. */
export type ScopeRow = {
	key: number;
	id?: string;
	name: string;
	description: string;
	default: boolean;
};

let nextKey = 0;

export function scopeRow(from?: Partial<ScopeRow>): ScopeRow {
	return { name: '', description: '', default: false, ...from, key: nextKey++ };
}

/** A scope name as the server will store it. */
const tidyScope = (value: string) => value.trim().toLowerCase();

const pattern = /^[a-z][a-z0-9]*([:._-][a-z0-9]+)*$/;

/** The OpenID Connect scopes, which an API scope may not be called. */
const openid = ['openid', 'profile', 'email', 'offline_access', 'roles'];

/** What is wrong with a row's name, said beside it before saving. */
export function scopeProblem(rows: ScopeRow[], row: ScopeRow): string | undefined {
	const value = tidyScope(row.name);

	if (value === '') return undefined;
	if (!pattern.test(value)) return 'Use lower case words joined by : . - or _, such as orders:read';
	if (openid.includes(value)) return 'That is an OpenID Connect scope';
	if (rows.filter((it) => tidyScope(it.name) === value).length > 1) return 'Listed twice';

	return undefined;
}

/** The rows as the API wants them: tidied, blanks left out. */
export function scopeInput(rows: ScopeRow[]): APIInput['scopes'] {
	return rows
		.filter((row) => tidyScope(row.name) !== '')
		.map((row) => ({
			id: row.id,
			name: tidyScope(row.name),
			description: row.description.trim(),
			default: row.default
		}));
}

export const algorithms: SelectOption<SigningAlgorithm>[] = [
	{
		value: 'RS256',
		label: 'RS256',
		description: 'RSA with SHA-256. Recommended: every JWT library supports it.'
	},
	{ value: 'PS256', label: 'PS256', description: 'RSA-PSS with SHA-256.' },
	{ value: 'ES256', label: 'ES256', description: 'ECDSA P-256 with SHA-256. Smaller tokens.' }
];

/** Everything an update sends, from the API as stored and what changed. */
export function apiInput(api: API, changes: Partial<APIInput>): APIInput {
	return {
		name: api.name,
		identifier: api.identifier,
		description: api.description,
		enforce_roles: api.enforce_roles,
		signing_algorithm: api.signing_algorithm,
		token_lifetime: api.token_lifetime,
		allow_offline_access: api.allow_offline_access,
		scopes: api.scopes.map((scope) => ({ ...scope })),
		...changes
	};
}

/** How the role-based access switch is explained, wherever it appears. */
export const roleBasedAccess = {
	label: 'Role-based access',
	description:
		"When enabled, users can receive only the scopes granted by their assigned roles. When disabled, access is determined by the application's allowed scopes."
};

/** A token lifetime in seconds, as a reader would say it. */
export function lifetimeLabel(seconds: number): string {
	if (seconds <= 0) return "Application's";
	if (seconds % 3600 === 0) return `${seconds / 3600} h`;
	if (seconds % 60 === 0) return `${seconds / 60} min`;

	return `${seconds} s`;
}
