/**
 * The navigation bar's timing, apart from its drawing: how it starts, how it
 * creeps while a page loads, and how it finishes. NavigationProgress.svelte
 * calls `start` as each navigation begins and `finish` once it has, and draws
 * `value` and `visible`; nothing here touches the page, so the timing can be
 * tested with a fake clock.
 *
 * It starts at once, the way NProgress and nextjs-toploader do, and every
 * navigation shows it: a quick one as a short sweep, a slow one as a bar
 * that creeps until the page arrives.
 */

/** How often the bar creeps forward while the page is still loading, and
    how far it may creep: it never reaches the end on its own, because only
    the page arriving says the wait is over. */
export const TRICKLE = 200;
export const CEILING = 0.94;

/** Where the bar jumps to as soon as it appears, so even a short wait shows
    a clear line rather than a sliver. */
export const START = 0.3;

/** How long the full bar stays before it fades. */
export const HOLD = 80;

export class NavigationProgress {
	/** How far along, from 0 to 1. */
	value = $state(0);
	visible = $state(false);
	/** Set while the bar jumps back to the start, so it does not visibly
	    slide backwards from the end. */
	instant = $state(false);

	/** A navigation is under way. */
	#running = false;
	/** Distinguishes overlapping navigations while a previous one settles. */
	#navigation = 0;
	/** Settles once the bar has been put back at the start on the page. */
	#ready: Promise<void> = Promise.resolve();

	#trickle: ReturnType<typeof setTimeout> | undefined;
	#hide: ReturnType<typeof setTimeout> | undefined;

	/** Puts the jump back to the start on the page before the bar moves
	    again. Given by the component, which owns the element. */
	readonly #settle: () => Promise<void>;

	constructor(settle: () => Promise<void> = async () => {}) {
		this.#settle = settle;
	}

	start() {
		clearTimeout(this.#hide);

		// A navigation that starts while another is still under way — a
		// redirect, or a second click — carries on with the same bar.
		if (this.#running) return;

		this.#running = true;
		const navigation = ++this.#navigation;
		clearTimeout(this.#trickle);

		this.instant = true;
		this.value = 0;
		this.visible = true;

		let ready: Promise<void>;
		try {
			ready = this.#settle();
		} catch {
			ready = Promise.reject(new Error('navigation progress could not settle'));
		}
		this.#ready = ready.then(
			() => {
				// A newer navigation may have taken over while this one settled.
				if (!this.#running || navigation !== this.#navigation) return;

				this.instant = false;
				this.value = START;
				this.#creep();
			},
			() => {
				// A failed settle means there is no bar to move. This is also
				// a teardown path, so do not leave a rejected promise behind.
				if (navigation !== this.#navigation) return;

				this.#navigation += 1;
				this.instant = false;
				this.#running = false;
				this.value = 0;
				this.visible = false;
			}
		);
	}

	/** Fills the bar and lets it go. A page that arrived before the bar was
	    even in place still gets it drawn from the start and filled, so a
	    quick navigation reads as a sweep rather than as nothing. */
	async finish() {
		if (!this.#running) return;

		const navigation = this.#navigation;
		this.#running = false;
		clearTimeout(this.#trickle);

		await this.#ready;

		// Another navigation began while this one was being put in place.
		if (this.#running || navigation !== this.#navigation) return;

		this.value = 1;
		this.#hide = setTimeout(() => (this.visible = false), HOLD);
	}

	dispose() {
		this.#running = false;
		this.#navigation += 1;
		clearTimeout(this.#trickle);
		clearTimeout(this.#hide);
	}

	/** Each step covers a share of what is left, so the bar slows as it goes
	    and never quite arrives. A little randomness keeps it from looking
	    like a metronome. */
	#creep() {
		this.#trickle = setTimeout(() => {
			this.value += (CEILING - this.value) * (0.1 + Math.random() * 0.1);
			this.#creep();
		}, TRICKLE);
	}
}
