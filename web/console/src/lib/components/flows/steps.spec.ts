import { describe, expect, it } from 'vitest';
import type { LoginStepSpec } from '$lib/api';
import {
	TEMPLATES,
	draftOf,
	freeSlug,
	fromFile,
	problemsOf,
	slugFrom,
	toFile,
	type FlowDraft
} from './steps';

const kinds = (['identifier', 'password', 'social', 'totp'] as const).map(
	(step): LoginStepSpec => ({
		step,
		label: step,
		description: '',
		fixed: step === 'identifier',
		implemented: step !== 'totp'
	})
);

const flow = (changes: Partial<FlowDraft> = {}): FlowDraft => ({
	...TEMPLATES[0].draft,
	name: 'Staff',
	slug: 'staff',
	...changes
});

describe('problemsOf', () => {
	const cases: [string, Partial<FlowDraft>, string[]][] = [
		['a flow that can be saved', {}, []],
		['no name', { name: ' ' }, ['flows.problem_name']],
		['an identifier with capitals', { slug: 'Staff' }, ['flows.problem_slug']],
		[
			'nothing that lets anybody in today',
			{ steps: ['identifier', 'totp'] },
			['flows.problem_proof']
		],
		['the default turned off', { is_default: true, enabled: false }, ['flows.problem_default_off']],
		['a session past ninety days', { session_lifetime_hours: 24 * 91 }, ['flows.problem_lifetime']]
	];

	it.each(cases)('%s', (_, changes, want) => {
		expect(problemsOf(flow(changes)).map((problem) => problem.key)).toEqual(want);
	});

	it('every template can be saved once it is named', () => {
		for (const template of TEMPLATES) {
			expect(problemsOf({ ...template.draft, name: 'x', slug: 'x' })).toEqual([]);
		}
	});
});

describe('identifiers', () => {
	it('are made from a name', () => {
		expect(slugFrom('Staff sign-in (EU)')).toBe('staff-sign-in-eu');
	});

	it('are made free when taken', () => {
		expect(freeSlug('staff', ['staff', 'staff-2'])).toBe('staff-3');
		expect(freeSlug('staff', [])).toBe('staff');
	});
});

describe('files', () => {
	it('come back as the flow that went out, minus being the default', () => {
		const draft = flow({ is_default: true, require_verified_email: true });
		const back = fromFile(JSON.stringify(toFile(draft)), kinds);

		expect(back).toEqual({ ...draftOf(draft), is_default: false });
	});

	it('say what is wrong with one that is not a flow', () => {
		expect(() => fromFile('not json', kinds)).toThrow('flows.import_not_json');
		expect(() => fromFile('{"name":"x","steps":["dance"]}', kinds)).toThrow(
			'flows.import_not_flow'
		);
	});
});
