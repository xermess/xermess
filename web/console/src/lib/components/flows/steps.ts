import {
	RiAtLine,
	RiCheckboxCircleLine,
	RiFileTextLine,
	RiKey2Line,
	RiMailLine,
	RiShareLine,
	RiTimerFlashLine
} from 'svelte-remixicon';
import type { ComponentType } from 'svelte';
import type { LoginStep, LoginStepSpec } from '$lib/api';

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

/** What a step is called, falling back to its own name for one this panel is
    older than. */
export function labelFor(step: LoginStep, kinds: LoginStepSpec[]): string {
	return kinds.find((kind) => kind.step === step)?.label ?? step;
}

export function specFor(step: LoginStep, kinds: LoginStepSpec[]): LoginStepSpec | undefined {
	return kinds.find((kind) => kind.step === step);
}

/** How long a session lasts, in the units people say it in. */
export function lifetime(hours: number): string {
	if (hours % 24 === 0 && hours >= 24) {
		const days = hours / 24;
		return `${days} ${days === 1 ? 'day' : 'days'}`;
	}

	return `${hours} ${hours === 1 ? 'hour' : 'hours'}`;
}
