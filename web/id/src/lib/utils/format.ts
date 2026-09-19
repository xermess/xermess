import type { Translate } from '$lib/i18n';

// Dates and times are said in the page's language by the browser's own Intl,
// which knows every language's words and grammar for them — the language is
// the translator's, so the server's first render and the browser agree.

/** A date as a person reads it: "Sep 15, 2026", "15 сент. 2026 г.". */
export function formatDate(value: string | null | undefined, t: Translate): string {
	if (!value) return '—';
	return new Date(value).toLocaleDateString(t.language, {
		day: 'numeric',
		month: 'short',
		year: 'numeric'
	});
}

const units: [number, Intl.RelativeTimeFormatUnit][] = [
	[60 * 60 * 24 * 365, 'year'],
	[60 * 60 * 24 * 30, 'month'],
	[60 * 60 * 24, 'day'],
	[60 * 60, 'hour'],
	[60, 'minute']
];

/** How long ago something was: "now", "5 minutes ago", "3 дня назад". */
export function timeAgo(value: string, t: Translate, now = Date.now()): string {
	const seconds = Math.max(0, Math.round((now - new Date(value).getTime()) / 1000));
	const format = new Intl.RelativeTimeFormat(t.language, { numeric: 'auto' });

	for (const [size, unit] of units) {
		const count = Math.floor(seconds / size);
		if (count >= 1) return format.format(-count, unit);
	}

	return format.format(0, 'second');
}

/** A user agent as a person would name the device: "Chrome on macOS". The
    browsers and systems are names, and read the same in every language. */
export function describeDevice(userAgent: string, t: Translate): string {
	const ua = userAgent || '';

	const browser =
		[
			['Edge', /Edg\//],
			['Opera', /OPR\//],
			['Firefox', /Firefox\//],
			['Chrome', /Chrome\//],
			['Safari', /Safari\//]
		].find(([, pattern]) => (pattern as RegExp).test(ua))?.[0] ?? t('device.unknown_browser');

	const system =
		[
			['iPhone', /iPhone/],
			['iPad', /iPad/],
			['Android', /Android/],
			['Windows', /Windows/],
			['macOS', /Mac OS X|Macintosh/],
			['Linux', /Linux/]
		].find(([, pattern]) => (pattern as RegExp).test(ua))?.[0] ?? t('device.unknown_system');

	return t('device.on', { browser: String(browser), system: String(system) });
}

/** The initials of a user for an avatar: "AL", or the email's first letter. */
export function initials(user: { first_name: string; last_name: string; email: string }): string {
	const letters = `${user.first_name.charAt(0)}${user.last_name.charAt(0)}`.trim();
	return (letters || user.email.charAt(0)).toUpperCase();
}
