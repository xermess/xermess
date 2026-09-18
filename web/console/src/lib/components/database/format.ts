import type { DatabaseColumn } from '$lib/api';

/** How a value is drawn: which of the three it is decides the colour, so a
    null, an empty string and the text "null" cannot be mistaken for one
    another. */
export type Cell =
	| { kind: 'value'; text: string }
	| { kind: 'empty'; text: string }
	| { kind: 'hidden'; text: string };

/** The longest a value is shown before it is cut. A cell is one line of a
    grid, and a token or a JSON blob would otherwise be the whole row. */
const maxLength = 120;

/** One value of one row, as the grid shows it.

    The server has already turned times into strings and bytes into how many
    there are, so what arrives here is what JSON can carry: a string, a
    number, a boolean, null, or an object for a column holding JSON. */
export function cell(value: unknown, column: DatabaseColumn): Cell {
	if (column.hidden) {
		return { kind: 'hidden', text: 'hidden' };
	}

	if (value === null || value === undefined) {
		return { kind: 'empty', text: 'null' };
	}

	if (typeof value === 'string') {
		if (value === '') return { kind: 'empty', text: 'empty' };

		return { kind: 'value', text: cut(value) };
	}

	if (typeof value === 'boolean' || typeof value === 'number') {
		return { kind: 'value', text: String(value) };
	}

	// A column holding JSON: the list of scopes, a role's permissions, a
	// user's own fields. An empty one is drawn as empty rather than as "[]",
	// which reads as a value.
	if (Array.isArray(value) && value.length === 0) {
		return { kind: 'empty', text: 'empty' };
	}

	return { kind: 'value', text: cut(JSON.stringify(value)) };
}

function cut(text: string): string {
	return text.length > maxLength ? `${text.slice(0, maxLength)}…` : text;
}

/** What one table holds, for the few whose name does not say it. Tables the
    panel has a page of its own for are named after it, so somebody looking at
    the list knows where to go to change what they are reading. */
export const describes: Record<string, string> = {
	admin_role_assignments: 'Which admin roles each administrator holds',
	admin_security: 'Whether administrators must sign in with a second factor',
	admin_user_sessions: "Administrators' sessions in this panel",
	admin_users: 'The administrators — see Administrators',
	api_scopes: 'The scopes each API defines',
	apis: 'The APIs tokens are issued for — see APIs',
	applications: 'The apps signing users in — see Applications',
	audit_logs: 'The activity log — see Logs',
	authorization_codes: 'Codes issued at the authorize endpoint, waiting to be exchanged',
	authorization_requests: 'Sign-ins under way, by the handle the sign-in pages carry',
	goose_db_version: 'Which migrations have run',
	login_flows: 'The login flows — see Login flows',
	mfas: "Administrators' authenticators",
	organizations: 'The organisation this installation belongs to — see Organization',
	password_resets: 'Reset links that have been sent',
	refresh_tokens: 'Refresh tokens that have been issued',
	roles: 'The admin roles — see Admin roles',
	signing_keys: 'The keys tokens are signed with',
	social_logins: 'Sign-ins started at a provider, waiting to come back',
	social_providers: 'The providers users may sign in with — see Social',
	user_fields: 'What a user record is made of — see Users',
	user_identities: 'The accounts elsewhere that sign users in',
	user_roles: 'The roles users hold — see Roles',
	user_sessions: "Users' sessions",
	users: 'The accounts themselves — see Users'
};
