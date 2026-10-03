/**
 * The shape of a theme, kept apart from its values (theme.ts) and renderer (render.ts) so the
 * renderer can be tested with its own theme.
 */

/** A token's name as the stylesheet knows it, without the leading dashes:
    "color-brand", "space-3". */
export type TokenName = string;

/** A flat set of tokens: name to CSS value. */
export type Tokens<Name extends string = TokenName> = Record<Name, string>;

/** One typeface the app bundles, in every weight and subset it ships. */
export type FontFace = {
	family: string;
	style: 'normal' | 'italic';
	weights: number[];
	/** Written in this order for each weight. A browser picks by
	    unicode-range, so the order only matters to someone reading. */
	subsets: string[];
	/** Where one weight of one subset lives, from `fonts.baseUrl`. */
	file: (subset: string, weight: number) => string;
};

export type Theme<ColorName extends string = string> = {
	fonts: {
		/** Relative to the generated fonts.css, which is in lib/styles. */
		baseUrl: string;
		/** The characters each subset covers, as unicode-range writes them. */
		subsets: Record<string, string>;
		faces: FontFace[];
	};

	/** Tokens with the same value in both themes, grouped by what they are.
	    Each group's names become `--<prefix>-<name>`; a group without a
	    prefix writes its names as they are. */
	scales: { prefix: string; tokens: Tokens }[];

	/** The window widths the layout changes at. CSS cannot read a variable
	    in a media query, so these are for TypeScript (a MediaQuery) and for
	    reading; a stylesheet writes the number and names the breakpoint in a
	    comment. */
	breakpoints: Record<string, string>;

	/** Tokens that change as the window widens, applied from `minWidth` up. */
	responsive: { minWidth: string; tokens: Partial<Tokens> }[];

	/** The light theme defines every colour; the dark one changes what it
	    must and inherits the rest. */
	colors: {
		light: Tokens<ColorName>;
		dark: Partial<Tokens<ColorName>>;
	};

	/** What floats over the page — a menu, a dialog — draws the page's own
	    tokens from a set of its own. Each layer names the parts it is drawn
	    on and the prefix of its set; `swapped` says which page token each of
	    the set's names stands in for. */
	layers: {
		swapped: Record<ColorName, string> | Partial<Record<ColorName, string>>;
		sets: { prefix: string; selectors: string[] }[];
	};
};

/** A token as a CSS value — `cssVar('color-brand')` is
    `var(--color-brand)` — for the rare style set from TypeScript, such as a
    chart's colours. */
export function cssVar(name: TokenName): string {
	return `var(--${name})`;
}
