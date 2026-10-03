import type { ComponentType } from 'svelte';

/**
 * One option of a Select. Only `value` is required, and bare strings are accepted; icon and
 * description are for richer lists.
 */
export type SelectOption<Value extends string = string> = {
	value: Value;
	/** The text on screen, and what typing a few letters matches against.
	 *  Defaults to the value. */
	label?: string;
	/** A quieter second line under the label. */
	description?: string;
	icon?: ComponentType;
	/** Shown, but not open to being chosen. */
	disabled?: boolean;
};
