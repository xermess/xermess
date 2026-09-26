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
import { BRAND } from '$lib/brand';

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

/** What a step is called: the server's catalog (model.LoginStepSpecs), and
    its own name for a step this panel is older than. */
export function labelFor(step: LoginStep, kinds: LoginStepSpec[]): string {
	return kinds.find((kind) => kind.step === step)?.label ?? step;
}

/** What a step does, from the same catalog. */
export function describe(step: LoginStep, kinds: LoginStepSpec[]): string {
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
	allow_sign_in: boolean;
	allow_registration: boolean;
	allow_password_reset: boolean;
	allow_remember_me: boolean;
	verify_email_on_register: boolean;
	require_verified_email: boolean;
	allow_email_change: boolean;
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
		allow_sign_in: flow.allow_sign_in,
		allow_registration: flow.allow_registration,
		allow_password_reset: flow.allow_password_reset,
		allow_remember_me: flow.allow_remember_me,
		verify_email_on_register: flow.verify_email_on_register,
		require_verified_email: flow.require_verified_email,
		allow_email_change: flow.allow_email_change,
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

/** What would stop a draft saving — the same rules the server holds, said as
    the flow is drawn. */
export function problemsOf(draft: FlowDraft): string[] {
	const problems: string[] = [];

	if (draft.name.trim() === '') problems.push('Name the flow.');
	if (!SLUG.test(draft.slug))
		problems.push('The identifier has to be lower case letters, numbers and dashes.');
	if (!draft.steps.some((step) => PROOF_STEPS.includes(step)))
		problems.push('Add Password or Other accounts: without one, nobody can sign in.');
	if (draft.is_default && !draft.enabled) problems.push('The default flow cannot be turned off.');
	if (
		!Number.isInteger(draft.session_lifetime_hours) ||
		draft.session_lifetime_hours < 1 ||
		draft.session_lifetime_hours > MAX_SESSION_HOURS
	)
		problems.push('A session has to last between one hour and ninety days.');

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
	/** What the template is called on the "new flow" page, and what it is
	    for — which becomes the new flow's own description. */
	name: string;
	hint: string;
	icon: ComponentType;
	draft: FlowDraft;
};

const base: Omit<FlowDraft, 'name' | 'slug' | 'steps'> = {
	description: '',
	is_default: false,
	enabled: true,
	allow_sign_in: true,
	allow_registration: true,
	allow_password_reset: true,
	allow_remember_me: true,
	verify_email_on_register: false,
	require_verified_email: false,
	allow_email_change: false,
	session_lifetime_hours: 24 * 14
};

/** Where a new flow starts: the shapes most installations want, written out
    so an administrator changes one rather than building from nothing. The
    names and descriptions are message keys, said in the panel's language. */
export const TEMPLATES: FlowTemplate[] = [
	{
		id: 'password',
		name: 'Password',
		hint: 'An email and a password, with the accounts from the Social page beside them.',
		icon: RiKey2Line,
		draft: { ...base, name: '', slug: '', steps: ['identifier', 'password', 'social'] }
	},
	{
		id: 'passwordless',
		name: 'Other accounts only',
		hint: 'No passwords here: people sign in with Google, GitHub and the other providers.',
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
		name: 'Staff',
		hint: 'A password, a verified address, no sign-ups, and eight-hour sessions.',
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
		name: 'Blank',
		hint: 'An email and a password, and nothing else to start from.',
		icon: RiFileAddLine,
		draft: { ...base, name: '', slug: '', steps: ['identifier', 'password'] }
	}
];

// ---- JSON -------------------------------------------------------------------

/** What an exported flow says it is, written once so the type and the file
    agree. The project's name is part of it, which is why it comes from the
    brand rather than being spelled here. */
const FLOW_SCHEMA = `${BRAND.slug}.login-flow/1`;

/** What Export writes and Import reads: a flow without anything that belongs
    to one installation — its id, its applications, whether it is the
    default here. */
export type FlowFile = Omit<FlowDraft, 'is_default'> & { $schema: typeof FLOW_SCHEMA };

export function toFile(draft: FlowDraft): FlowFile {
	const copy = draftOf(draft);
	return {
		$schema: FLOW_SCHEMA,
		name: copy.name,
		slug: copy.slug,
		description: copy.description,
		enabled: copy.enabled,
		steps: copy.steps,
		allow_sign_in: copy.allow_sign_in,
		allow_registration: copy.allow_registration,
		allow_password_reset: copy.allow_password_reset,
		allow_remember_me: copy.allow_remember_me,
		verify_email_on_register: copy.verify_email_on_register,
		require_verified_email: copy.require_verified_email,
		allow_email_change: copy.allow_email_change,
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
		throw new Error('That file is not JSON.');
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
		throw new Error('That file is not a login flow exported from here.');
	}

	return {
		...base,
		name: file.name,
		slug: typeof file.slug === 'string' ? file.slug : slugFrom(file.name),
		description: typeof file.description === 'string' ? file.description : '',
		enabled: file.enabled !== false,
		steps: file.steps as LoginStep[],
		// A switch the file does not mention keeps the default: the ones that
		// are on unless said otherwise read `!== false`, and the ones that are
		// off unless said otherwise read `=== true`. That way a flow exported
		// before a switch existed imports as a flow made today would.
		allow_sign_in: file.allow_sign_in !== false,
		allow_registration: file.allow_registration !== false,
		allow_password_reset: file.allow_password_reset !== false,
		allow_remember_me: file.allow_remember_me !== false,
		verify_email_on_register: file.verify_email_on_register === true,
		require_verified_email: file.require_verified_email === true,
		allow_email_change: file.allow_email_change === true,
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
