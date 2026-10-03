/**
 * Toasts for actions on the account pages, drawn by <Toaster /> in the root layout. Callers
 * pass already translated text.
 *
 *     notify.success(t('profile.saved'));
 *     notify.error(messageOf(err, t));
 *
 * Errors about the one form on a sign-in page stay beside that form. Written in-house to keep
 * the app dependent on Svelte alone.
 */

export type ToastTone = 'success' | 'danger';

export type Toast = {
	id: number;
	tone: ToastTone;
	title: string;
	description?: string;
};

/** How long a toast stays, in milliseconds. An error stays longer: it may
    need reading twice, and acting on. */
const DURATION: Record<ToastTone, number> = { success: 5000, danger: 8000 };

/** No more than this many at once; the oldest goes first. */
const MAX = 3;

let next = 0;
/** Each toast's timer. Bookkeeping, never rendered, so not reactive. */
type Timer = { left: number; started: number; handle?: ReturnType<typeof setTimeout> };
const timers: Record<number, Timer> = {};

export const toasts = $state<Toast[]>([]);

function add(tone: ToastTone, title: string, description?: string) {
	// Only a browser shows a toast: the server renders none, and a list kept
	// in a module would be shared by every request it renders.
	if (typeof window === 'undefined') return;

	const toast: Toast = { id: ++next, tone, title, description };
	toasts.push(toast);
	while (toasts.length > MAX) dismiss(toasts[0].id);

	timers[toast.id] = { left: DURATION[tone], started: 0 };
	resume(toast.id);
}

export function dismiss(id: number) {
	const timer = timers[id];
	if (timer?.handle) clearTimeout(timer.handle);
	delete timers[id];

	const index = toasts.findIndex((toast) => toast.id === id);
	if (index >= 0) toasts.splice(index, 1);
}

/** Holds a toast while it is pointed at or focused: nobody loses a message
    while reading it. */
export function pause(id: number) {
	const timer = timers[id];
	if (!timer?.handle) return;

	clearTimeout(timer.handle);
	timer.handle = undefined;
	timer.left -= Date.now() - timer.started;
}

export function resume(id: number) {
	const timer = timers[id];
	if (!timer || timer.handle) return;

	timer.started = Date.now();
	timer.handle = setTimeout(() => dismiss(id), Math.max(timer.left, 1500));
}

export const notify = {
	success: (title: string, description?: string) => add('success', title, description),
	error: (title: string, description?: string) => add('danger', title, description)
};
