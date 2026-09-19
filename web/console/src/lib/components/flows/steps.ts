import {
	RiAtLine,
	RiCheckboxCircleLine,
	RiFileAddLine,
	RiFileTextLine,
	RiKey2Line,
	RiMailLine,
	RiShareLine,
	RiShieldCheckLine,
	RiTimerFlashLine
} from 'svelte-remixicon';
import type { ComponentType } from 'svelte';
import type { LoginStep, LoginStepSpec } from '$lib/api';
import type { Translate } from '$lib/i18n';

/** The mark beside each step, wherever one is listed. The step catalog itself
    is the server's — what a step is called and whether it is run yet comes
    from there — and this is the one thing about a step the panel decides. */
const marks: Record<LoginStep, ComponentType> = {
	identifier: RiAtLine,
	password: RiKey2Line,
	social: RiShareLine,
	email_code: RiMailLine,
	totp: RiTimerFlashLine,
	terms: RiFileTextLine,
	consent: RiCheckboxCircleLine
};

export function markFor(step: LoginStep): ComponentType {
	return marks[step] ?? RiCheckboxCircleLine;
}

/** What a step is called in the panel's language: its `flows.step_<step>`
    when the panel has one, else the server's English, else its own name for
    a step this panel is older than. */
export function labelFor(step: LoginStep, kinds: LoginStepSpec[], t?: Translate): string {
	const key = `flows.step_${step}`;
	if (t?.has(key)) return t(key);
	return kinds.find((kind) => kind.step === step)?.label ?? step;
}

/** What a step does, in the panel's language as labelFor is. */
export function describe(step: LoginStep, kinds: LoginStepSpec[], t?: Translate): string {
	const key = `flows.step_${step}_hint`;
	if (t?.has(key)) return t(key);
	return kinds.find((kind) => kind.step === step)?.description ?? '';
}

export function specFor(step: LoginStep, kinds: LoginStepSpec[]): LoginStepSpec | undefined {
	return kinds.find((kind) => kind.step === step);
}

// ---- A flow being edited ----------------------------------------------------

/** What the editor changes: a flow without what the server keeps about it. */
export type FlowDraft = {
	name: string;
	slug: string;
	description: string;
	is_default: boolean;
	enabled: boolean;
	steps: LoginStep[];
	allow_registration: boolean;
	allow_password_reset: boolean;
	require_verified_email: boolean;
	session_lifetime_hours: number;
};

export function draftOf(flow: FlowDraft): FlowDraft {
	return {
		name: flow.name,
		slug: flow.slug,
		description: flow.description,
		is_default: flow.is_default,
		enabled: flow.enabled,
		steps: [...flow.steps],
		allow_registration: flow.allow_registration,
		allow_password_reset: flow.allow_password_reset,
		require_verified_email: flow.require_verified_email,
		session_lifetime_hours: flow.session_lifetime_hours
	};
}

/** The steps that prove who somebody is and that the sign-in pages run today.
    The server refuses a flow with neither (model.proofSteps), and the editor
    says so before anybody presses Save. */
export const PROOF_STEPS: LoginStep[] = ['password', 'social'];

/** The longest a session may last: ninety days, as the server holds it. */
export const MAX_SESSION_HOURS = 24 * 90;

const SLUG = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

/** What would stop a draft saving, as message keys and their parameters —
    the same rules the server holds, said as the flow is drawn. */
export function problemsOf(draft: FlowDraft): { key: string; params?: Record<string, string> }[] {
	const problems: { key: string; params?: Record<string, string> }[] = [];

	if (draft.name.trim() === '') problems.push({ key: 'flows.problem_name' });
	if (!SLUG.test(draft.slug)) problems.push({ key: 'flows.problem_slug' });
	if (!draft.steps.some((step) => PROOF_STEPS.includes(step)))
		problems.push({ key: 'flows.problem_proof' });
	if (draft.is_default && !draft.enabled) problems.push({ key: 'flows.problem_default_off' });
	if (
		!Number.isInteger(draft.session_lifetime_hours) ||
		draft.session_lifetime_hours < 1 ||
		draft.session_lifetime_hours > MAX_SESSION_HOURS
	)
		problems.push({ key: 'flows.problem_lifetime' });

	return problems;
}

/** A name turned into an identifier: "Staff sign-in" is staff-sign-in. */
export function slugFrom(name: string): string {
	return name
		.toLowerCase()
		.normalize('NFKD')
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-+|-+$/g, '')
		.slice(0, 64);
}

/** A slug nobody has yet: the one asked for, or it with -2, -3… after it. */
export function freeSlug(wanted: string, taken: string[]): string {
	const base = wanted || 'flow';
	if (!taken.includes(base)) return base;

	for (let n = 2; ; n++) {
		const candidate = `${base.slice(0, 60)}-${n}`;
		if (!taken.includes(candidate)) return candidate;
	}
}

// ---- Templates ----------------------------------------------------------------

export type FlowTemplate = {
	id: string;
	icon: ComponentType;
	draft: FlowDraft;
};

const base: Omit<FlowDraft, 'name' | 'slug' | 'steps'> = {
	description: '',
	is_default: false,
	enabled: true,
	allow_registration: true,
	allow_password_reset: true,
	require_verified_email: false,
	session_lifetime_hours: 24 * 14
};

/** Where a new flow starts: the shapes most installations want, written out
    so an administrator changes one rather than building from nothing. The
    names and descriptions are message keys, said in the panel's language. */
export const TEMPLATES: FlowTemplate[] = [
	{
		id: 'password',
		icon: RiKey2Line,
		draft: { ...base, name: '', slug: '', steps: ['identifier', 'password', 'social'] }
	},
	{
		id: 'passwordless',
		icon: RiShareLine,
		draft: {
			...base,
			name: '',
			slug: '',
			steps: ['identifier', 'social'],
			allow_password_reset: false
		}
	},
	{
		id: 'staff',
		icon: RiShieldCheckLine,
		draft: {
			...base,
			name: '',
			slug: '',
			steps: ['identifier', 'password'],
			allow_registration: false,
			require_verified_email: true,
			session_lifetime_hours: 8
		}
	},
	{
		id: 'blank',
		icon: RiFileAddLine,
		draft: { ...base, name: '', slug: '', steps: ['identifier', 'password'] }
	}
];

// ---- JSON -------------------------------------------------------------------

/** What Export writes and Import reads: a flow without anything that belongs
    to one installation — its id, its applications, whether it is the
    default here. */
export type FlowFile = Omit<FlowDraft, 'is_default'> & { $schema: 'xermess.login-flow/1' };

export function toFile(draft: FlowDraft): FlowFile {
	const copy = draftOf(draft);
	return {
		$schema: 'xermess.login-flow/1',
		name: copy.name,
		slug: copy.slug,
		description: copy.description,
		enabled: copy.enabled,
		steps: copy.steps,
		allow_registration: copy.allow_registration,
		allow_password_reset: copy.allow_password_reset,
		require_verified_email: copy.require_verified_email,
		session_lifetime_hours: copy.session_lifetime_hours
	};
}

/** Reads an exported flow, or says why it is not one. Only the shape is
    checked here: the server holds it to the rules when it is saved. */
export function fromFile(text: string, kinds: LoginStepSpec[]): FlowDraft {
	let raw: unknown;
	try {
		raw = JSON.parse(text);
	} catch {
		throw new Error('flows.import_not_json');
	}

	const file = raw as Partial<FlowFile>;
	const known = new Set(kinds.map((kind) => kind.step));
	if (
		typeof file !== 'object' ||
		file === null ||
		typeof file.name !== 'string' ||
		!Array.isArray(file.steps) ||
		!file.steps.every((step) => known.has(step as LoginStep))
	) {
		throw new Error('flows.import_not_flow');
	}

	return {
		...base,
		name: file.name,
		slug: typeof file.slug === 'string' ? file.slug : slugFrom(file.name),
		description: typeof file.description === 'string' ? file.description : '',
		enabled: file.enabled !== false,
		steps: file.steps as LoginStep[],
		allow_registration: file.allow_registration !== false,
		allow_password_reset: file.allow_password_reset !== false,
		require_verified_email: file.require_verified_email === true,
		session_lifetime_hours:
			typeof file.session_lifetime_hours === 'number' ? file.session_lifetime_hours : 24 * 14
	};
}

/** Hands the browser a file to save. */
export function download(name: string, content: string) {
	const url = URL.createObjectURL(new Blob([content], { type: 'application/json' }));
	const link = Object.assign(document.createElement('a'), { href: url, download: name });
	link.click();
	URL.revokeObjectURL(url);
}
