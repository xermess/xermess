import { ApiError } from './client';

/**
 * The message to show an administrator for an error. The API's English sentence comes with its
 * parameters filled in, and the panel is English, so it is shown as is; anything without one (a
 * dropped connection, a bug) shows `fallback`.
 */
export function messageOf(err: unknown, fallback = 'Something went wrong. Try again.'): string {
	if (!(err instanceof ApiError)) return fallback;

	return err.message || fallback;
}
