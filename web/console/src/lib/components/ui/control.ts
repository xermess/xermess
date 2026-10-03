import type { ComponentType } from 'svelte';

/**
 * Props every control shares, mapped one-to-one onto data attributes in styles/controls.css. A
 * new size or palette is CSS there plus a value here.
 */

/** How tall a control is. `md` is the default, and what a toolbar is built
 *  from; `sm` is for controls inside a row or a bar. */
export type Size = 'sm' | 'md' | 'lg';

/** How much of the palette a control uses: `solid` is filled and loud,
 *  `subtle` is filled and quiet, `outline` is a border, `ghost` is nothing
 *  until it is pointed at, `plain` is a link that happens to be a button. */
export type Variant = 'solid' | 'subtle' | 'outline' | 'ghost' | 'plain';

/** Which colours those variants use. See styles/palettes.css. */
export type ColorPalette = 'neutral' | 'danger' | 'success' | 'info' | 'warning';

export type ControlProps = {
	size?: Size;
	variant?: Variant;
	colorPalette?: ColorPalette;
	/** Waiting on something: the control shows a spinner and refuses clicks. */
	loading?: boolean;
	disabled?: boolean;
	/** Drawn before the label. */
	icon?: ComponentType;
};
