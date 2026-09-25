import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { CEILING, DELAY, HOLD, NavigationProgress, TRICKLE } from './progress.svelte';

describe('the navigation progress bar', () => {
	beforeEach(() => {
		vi.useFakeTimers();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it('shows nothing for a navigation that finishes before the delay', async () => {
		const bar = new NavigationProgress();

		bar.start();
		await vi.advanceTimersByTimeAsync(DELAY - 10);
		bar.finish();
		await vi.advanceTimersByTimeAsync(1000);

		expect(bar.visible).toBe(false);
		expect(bar.value).toBe(0);
	});

	it('appears after the delay and creeps without reaching the end', async () => {
		const bar = new NavigationProgress();

		bar.start();
		await vi.advanceTimersByTimeAsync(DELAY);

		expect(bar.visible).toBe(true);
		expect(bar.instant).toBe(false);
		expect(bar.value).toBe(0.1);

		let last = bar.value;
		for (let step = 0; step < 40; step++) {
			await vi.advanceTimersByTimeAsync(TRICKLE);
			expect(bar.value).toBeGreaterThan(last);
			last = bar.value;
		}

		expect(bar.value).toBeLessThan(CEILING);
		bar.dispose();
	});

	it('fills when the page arrives, then hides', async () => {
		const bar = new NavigationProgress();

		bar.start();
		await vi.advanceTimersByTimeAsync(DELAY + TRICKLE * 3);
		bar.finish();

		expect(bar.value).toBe(1);
		expect(bar.visible).toBe(true);

		await vi.advanceTimersByTimeAsync(HOLD);
		expect(bar.visible).toBe(false);

		// Nothing is left running once it has gone.
		const settled = bar.value;
		await vi.advanceTimersByTimeAsync(TRICKLE * 4);
		expect(bar.value).toBe(settled);
	});

	it('carries on with the same bar when a navigation starts while it is out', async () => {
		const bar = new NavigationProgress();

		bar.start();
		await vi.advanceTimersByTimeAsync(DELAY + TRICKLE * 4);
		const along = bar.value;

		bar.start();
		await vi.advanceTimersByTimeAsync(0);

		expect(bar.visible).toBe(true);
		expect(bar.value).toBeGreaterThanOrEqual(along);
		bar.dispose();
	});

	it('starts again from nothing after the last one has finished', async () => {
		const bar = new NavigationProgress();

		bar.start();
		await vi.advanceTimersByTimeAsync(DELAY + TRICKLE);
		bar.finish();
		await vi.advanceTimersByTimeAsync(HOLD);

		bar.start();
		await vi.advanceTimersByTimeAsync(DELAY);

		expect(bar.visible).toBe(true);
		expect(bar.value).toBe(0.1);
		bar.dispose();
	});

	it('does not start creeping if the page arrives while the bar is being put in place', async () => {
		let release: () => void = () => {};
		const bar = new NavigationProgress(() => new Promise((resolve) => (release = resolve)));

		bar.start();
		await vi.advanceTimersByTimeAsync(DELAY);
		bar.finish();
		release();
		await vi.advanceTimersByTimeAsync(TRICKLE * 4);

		expect(bar.value).toBe(1);
	});
});
