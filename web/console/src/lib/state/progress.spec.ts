import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { CEILING, HOLD, NavigationProgress, START, TRICKLE } from './progress.svelte';

describe('the navigation progress bar', () => {
	beforeEach(() => {
		vi.useFakeTimers();
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it('appears as a navigation starts and creeps without reaching the end', async () => {
		const bar = new NavigationProgress();

		bar.start();
		expect(bar.visible).toBe(true);

		await vi.advanceTimersByTimeAsync(0);
		expect(bar.instant).toBe(false);
		expect(bar.value).toBe(START);

		let last = bar.value;
		for (let step = 0; step < 40; step++) {
			await vi.advanceTimersByTimeAsync(TRICKLE);
			expect(bar.value).toBeGreaterThan(last);
			last = bar.value;
		}

		expect(bar.value).toBeLessThan(CEILING);
		bar.dispose();
	});

	it('still sweeps across for a page that arrives at once', async () => {
		const bar = new NavigationProgress();

		bar.start();
		const finishing = bar.finish();
		expect(bar.visible).toBe(true);

		await finishing;
		expect(bar.value).toBe(1);
		expect(bar.visible).toBe(true);

		await vi.advanceTimersByTimeAsync(HOLD);
		expect(bar.visible).toBe(false);
	});

	it('fills when the page arrives, then hides, and stops creeping', async () => {
		const bar = new NavigationProgress();

		bar.start();
		await vi.advanceTimersByTimeAsync(TRICKLE * 3);
		await bar.finish();

		expect(bar.value).toBe(1);
		await vi.advanceTimersByTimeAsync(HOLD);
		expect(bar.visible).toBe(false);

		await vi.advanceTimersByTimeAsync(TRICKLE * 4);
		expect(bar.value).toBe(1);
	});

	it('carries on with the same bar when a navigation starts while one is under way', async () => {
		const bar = new NavigationProgress();

		bar.start();
		await vi.advanceTimersByTimeAsync(TRICKLE * 4);
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
		await bar.finish();
		await vi.advanceTimersByTimeAsync(HOLD);

		bar.start();
		expect(bar.value).toBe(0);
		await vi.advanceTimersByTimeAsync(0);

		expect(bar.visible).toBe(true);
		expect(bar.value).toBe(START);
		bar.dispose();
	});

	it('keeps the bar when a new navigation begins while the last is being put in place', async () => {
		const pending: (() => void)[] = [];
		const bar = new NavigationProgress(() => new Promise<void>((resolve) => pending.push(resolve)));

		bar.start();
		const finishing = bar.finish();
		bar.start();
		pending.forEach((release) => release());
		await finishing;
		await vi.advanceTimersByTimeAsync(0);

		expect(bar.visible).toBe(true);
		expect(bar.value).toBeLessThan(1);
		expect(vi.getTimerCount()).toBe(1);
		bar.dispose();
	});

	it('hides without rejecting when the bar cannot settle', async () => {
		let reject!: (reason?: unknown) => void;
		const settling = new Promise<void>((_, fail) => (reject = fail));
		const bar = new NavigationProgress(() => settling);

		bar.start();
		const finishing = bar.finish();
		reject(new Error('the bar went away'));
		await expect(finishing).resolves.toBeUndefined();

		expect(bar.visible).toBe(false);
		expect(bar.value).toBe(0);
		bar.dispose();
	});
});
