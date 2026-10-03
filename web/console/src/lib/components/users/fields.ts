import type { UserField, UserRecord } from '$lib/api';

/**
 * Reads and writes a field's value: built-in fields are properties of the user, additional ones
 * live in `data`.
 */

/** What this record holds for that field. */
export function valueOf(user: UserRecord, field: UserField): unknown {
	if (field.is_builtin) {
		return (user as unknown as Record<string, unknown>)[field.name];
	}

	return user.data?.[field.name];
}

/** The built-in fields, in the order the API gave them. */
export function builtins(fields: UserField[]): UserField[] {
	return fields.filter((field) => field.is_builtin);
}

/** The fields this organisation added. */
export function additional(fields: UserField[]): UserField[] {
	return fields.filter((field) => !field.is_builtin);
}
