/**
 * Toasts: what an action came to — saved, sent, could not delete — said in a
 * corner of the screen, rather than in a message at the top of a form that
 * may be scrolled out of sight.
 *
 * One toaster serves the whole panel; <Toaster /> in the root layout draws
 * it. A component reports the outcome of what the administrator just did:
 *
 *   notify.success('Settings saved');
 *   notify.error(err, 'Could not save the settings');
 *
 * What describes the page itself, rather than an action — "your roles let
 * you read this", "could not load the log" — stays an <Alert> in its place.
 */
import { createToaster } from '@ark-ui/svelte/toast';
import { messageOf } from '$lib/api';

export const toaster = createToaster({
	placement: 'bottom',
	gap: 10,
	max: 4,
	duration: 5000,
	// Paused while the tab is hidden, so a toast is not gone before anyone
	// has had the chance to read it.
	pauseOnPageIdle: true
});

/** An error stays longer: it may need reading twice, and acting on. */
const ERROR_DURATION = 8000;

export const notify = {
	/** Something worked. `description` says what follows from it. */
	success(title: string, description?: string) {
		toaster.success({ title, description });
	},

	/** Something failed: the server's sentence when it sent one, the
	    fallback otherwise — never a stack trace or a status code. */
	error(err: unknown, fallback: string) {
		toaster.error({ title: messageOf(err, fallback), duration: ERROR_DURATION });
	},

	/** A failure the panel found itself, before asking the server. */
	problem(message: string) {
		toaster.error({ title: message, duration: ERROR_DURATION });
	}
};
