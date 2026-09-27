import type { UserField, UserRecord } from '$lib/api';

/**
 * Reading and writing a field's value, whichever kind it is.
 *
 * A built-in field is a column of the record and a property of the object; an
 * additional one lives under its name in `data`. Everything that draws a
 * table cell or an input goes through here, so the difference is in one file
 * rather than in every component that shows a field.
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
