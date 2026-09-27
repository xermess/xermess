import { describe, expect, it } from 'vitest';

import { lifetimeLabel, scopeInput, scopeProblem, scopeRow } from './scopes';

describe('scopeProblem', () => {
	it('accepts the names the server accepts', () => {
		const rows = [scopeRow({ name: 'orders:read' }), scopeRow({ name: 'billing.invoices.export' })];

		expect(rows.map((row) => scopeProblem(rows, row))).toEqual([undefined, undefined]);
	});

	it('refuses a badly formed name, an OpenID scope and a repeat', () => {
		const bad = scopeRow({ name: 'Orders Read' });
		const openid = scopeRow({ name: 'email' });
		const first = scopeRow({ name: 'orders:read' });
		const repeat = scopeRow({ name: ' ORDERS:READ ' });
		const rows = [bad, openid, first, repeat];

		expect(scopeProblem(rows, bad)).toMatch(/lower case/);
		expect(scopeProblem(rows, openid)).toMatch(/OpenID/);
		expect(scopeProblem(rows, repeat)).toBe('Listed twice');
	});

	it('says nothing about a row not filled in yet', () => {
		const empty = scopeRow();

		expect(scopeProblem([empty], empty)).toBeUndefined();
	});
});

describe('scopeInput', () => {
	it('tidies the rows and leaves the empty ones out', () => {
		const rows = [
			scopeRow({ id: 's1', name: ' Orders:Read ', description: ' Read orders ', is_default: true }),
			scopeRow()
		];

		expect(scopeInput(rows)).toEqual([
			{ id: 's1', name: 'orders:read', description: 'Read orders', is_default: true }
		]);
	});
});

describe('lifetimeLabel', () => {
	it('says a lifetime the way a reader would', () => {
		expect(lifetimeLabel(0)).toBe("Application's");
		expect(lifetimeLabel(900)).toBe('15 min');
		expect(lifetimeLabel(7200)).toBe('2 h');
		expect(lifetimeLabel(90)).toBe('90 s');
	});
});
