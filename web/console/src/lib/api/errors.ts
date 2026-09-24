import { ApiError } from './client';

/**
 * What to tell an administrator about an error.
 *
 * Every problem the API answers with carries a sentence as well as a code —
 * `{"error": "The mail server did not take the message: …", "code":
 * "mail_test_failed"}` — with its parameters already filled in. The panel is
 * written in English, so that sentence is the one to show: there is no
 * catalog here to look the code up in, and nothing would be gained by having
 * one that said the same thing twice.
 *
 * Anything that is not an answer at all — a dropped connection, a bug — has
 * no sentence, so the caller's `fallback` says what it was trying to do.
 */
export function messageOf(err: unknown, fallback = 'Something went wrong. Try again.'): string {
	if (!(err instanceof ApiError)) return fallback;

	return err.message || fallback;
}
