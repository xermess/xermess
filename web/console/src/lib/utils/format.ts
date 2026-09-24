/**
 * How values are shown to a reader.
 *
 * Dates arrive from the API as RFC 3339 strings and are shown in the reader's
 * own locale and time zone. Anything that is not a date is handed back as it
 * came: "Invalid Date" in a table tells nobody anything, while the value
 * itself at least says what was stored.
 */
export function formatDateTime(value: string): string {
	const date = new Date(value);

	return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

/** The day alone, for a value whose time of day means nothing. */
export function formatDate(value: string): string {
	const date = new Date(value);

	return Number.isNaN(date.getTime()) ? value : date.toLocaleDateString();
}

const MINUTE = 60_000;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

/** How long ago something happened, the way a feed says it: "just now",
    "5 min ago", "3 h ago", "yesterday", "4 days ago", then the date. */
export function formatRelative(value: string, now: Date = new Date()): string {
	const date = new Date(value);
	if (Number.isNaN(date.getTime())) return value;

	const elapsed = now.getTime() - date.getTime();

	if (elapsed < MINUTE) return 'just now';
	if (elapsed < HOUR) return `${Math.floor(elapsed / MINUTE)} min ago`;
	if (elapsed < DAY && sameDay(date, now)) return `${Math.floor(elapsed / HOUR)} h ago`;

	const days = calendarDaysBetween(date, now);
	if (days === 1) return 'yesterday';
	if (days < 7) return `${days} days ago`;

	return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

/** The heading a day's entries sit under: "Today", "Yesterday", or the
    weekday and date. */
export function formatDayHeading(value: string, now: Date = new Date()): string {
	const date = new Date(value);
	if (Number.isNaN(date.getTime())) return value;

	const days = calendarDaysBetween(date, now);
	if (days === 0) return 'Today';
	if (days === 1) return 'Yesterday';

	return date.toLocaleDateString(undefined, { weekday: 'long', month: 'short', day: 'numeric' });
}

/** The time of day alone, for an entry already under its day's heading. */
export function formatTime(value: string): string {
	const date = new Date(value);

	return Number.isNaN(date.getTime())
		? value
		: date.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
}

/** A short date for a chart axis, from a YYYY-MM-DD day. */
export function formatShortDay(day: string): string {
	const [year, month, date] = day.split('-').map(Number);
	const parsed = new Date(year, month - 1, date);

	return Number.isNaN(parsed.getTime())
		? day
		: parsed.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
}

/** A browser and system out of a user agent — "Chrome on macOS" — for a
    list of sessions nobody wants to read raw user agents in. */
export function describeUserAgent(agent: string): string {
	if (!agent) return 'Unknown device';

	const browser =
		[
			['Edge', /Edg\//],
			['Opera', /OPR\//],
			['Firefox', /Firefox\//],
			['Chrome', /Chrome\//],
			['Safari', /Version\/.*Safari\//],
			['curl', /^curl\//]
		].find(([, pattern]) => (pattern as RegExp).test(agent))?.[0] ?? 'Browser';

	const system =
		[
			['iOS', /iPhone|iPad/],
			['Android', /Android/],
			['macOS', /Mac OS X|Macintosh/],
			['Windows', /Windows/],
			['Linux', /Linux/]
		].find(([, pattern]) => (pattern as RegExp).test(agent))?.[0] ?? '';

	return system ? `${browser} on ${system}` : String(browser);
}

function sameDay(a: Date, b: Date): boolean {
	return calendarDaysBetween(a, b) === 0;
}

/** Whole calendar days from `from` to `to`, in the reader's time zone. */
function calendarDaysBetween(from: Date, to: Date): number {
	const start = new Date(from.getFullYear(), from.getMonth(), from.getDate());
	const end = new Date(to.getFullYear(), to.getMonth(), to.getDate());

	return Math.round((end.getTime() - start.getTime()) / DAY);
}

/** Two letters standing in for a picture: the first letters of a name in
    several words, or the start of a name in one. */
export function initials(name: string): string {
	const words = name.split(/\s+/).filter(Boolean);
	const letters = words.length > 1 ? words[0][0] + words[1][0] : (words[0] ?? '').slice(0, 2);

	return letters.toUpperCase();
}
