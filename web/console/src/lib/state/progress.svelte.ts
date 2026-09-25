/**
 * The navigation bar's timing, apart from its drawing: when it appears, how
 * it creeps while a page loads, and how it finishes. NavigationProgress.svelte
 * draws `value` and `visible` and calls `start` and `finish`; nothing here
 * touches the page, so the timing can be tested with a fake clock.
 */

/** A navigation that finishes this fast shows nothing: a bar that flashes
    for a frame on every quick page reads as flicker, not as progress. */
export const DELAY = 80;

/** How often the bar creeps forward while the page is still loading, and
    how far it may creep: it never reaches the end on its own, because only
    the page arriving says the wait is over. */
export const TRICKLE = 250;
export const CEILING = 0.92;

/** How long the full bar stays before it fades. */
export const HOLD = 200;

export class NavigationProgress {
	/** How far along, from 0 to 1. */
	value = $state(0);
	visible = $state(false);
	/** Set while the bar jumps back to the start, so it does not visibly
	    slide backwards from the end. */
	instant = $state(false);

	#delay: ReturnType<typeof setTimeout> | undefined;
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

		// A navigation that starts while the bar is still out — a redirect,
		// or a second click — carries on with the same bar.
		if (this.visible || this.#delay) return;

		this.#delay = setTimeout(async () => {
			this.#delay = undefined;
			this.instant = true;
			this.value = 0;
			this.visible = true;

			await this.#settle();

			// The page may have arrived in the meantime, and finished the bar.
			if (this.value !== 0) return;

			this.instant = false;
			this.value = 0.1;
			this.#creep();
		}, DELAY);
	}

	finish() {
		clearTimeout(this.#delay);
		clearTimeout(this.#trickle);
		this.#delay = undefined;

		if (!this.visible) return;

		this.value = 1;
		this.#hide = setTimeout(() => (this.visible = false), HOLD);
	}

	dispose() {
		clearTimeout(this.#delay);
		clearTimeout(this.#trickle);
		clearTimeout(this.#hide);
	}

	/** Each step covers a share of what is left, so the bar slows as it goes
	    and never quite arrives. A little randomness keeps it from looking
	    like a metronome. */
	#creep() {
		this.#trickle = setTimeout(() => {
			this.value += (CEILING - this.value) * (0.08 + Math.random() * 0.08);
			this.#creep();
		}, TRICKLE);
	}
}
