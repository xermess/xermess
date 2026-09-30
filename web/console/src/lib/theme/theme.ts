/**
 * The console's theme: every colour, size, typeface and speed the panel is
 * drawn with, in one place.
 *
 * This file is the source. `bun run theme` renders it into
 * lib/styles/tokens.css and lib/styles/fonts.css, which the app loads, and
 * `bun run check` fails while those two are behind it — so a value is changed
 * here, never in the generated CSS. Components go on naming tokens the way
 * they always have (`var(--color-brand)`, `var(--space-3)`); nothing reads
 * this file at run time.
 *
 * The layout and spacing come from the PocketBase admin UI; the colour is
 * Telegram's: a white or a black page, cool greys a step apart, and one brand
 * blue, #0088cc, for everything chosen, pressed or current.
 */
import type { Theme } from './types';

/* ---- Type ------------------------------------------------------------ */

/** The system's faces: the fallback while Chirp downloads, and for scripts
    it does not cover. The emoji faces keep an emoji in a name in colour
    rather than as a box. */
const systemSans = [
	'system-ui',
	'-apple-system',
	'BlinkMacSystemFont',
	"'Segoe UI'",
	"'Helvetica Neue'",
	"'Noto Sans'",
	'Ubuntu',
	'Cantarell',
	'Arial',
	'sans-serif',
	"'Apple Color Emoji'",
	"'Segoe UI Emoji'",
	"'Noto Color Emoji'"
].join(', ');

const systemMono = [
	'ui-monospace',
	'SFMono-Regular',
	"'SF Mono'",
	'Menlo',
	'Consolas',
	"'Liberation Mono'",
	"'DejaVu Sans Mono'",
	'monospace'
].join(', ');

/* ---- Colour ---------------------------------------------------------- */

/**
 * The light palette, and the name of every colour token there is. Every value
 * is solid — a colour mixed with transparency is a different grey over every
 * surface it lands on.
 */
const light = {
	// The brand: whatever is chosen, pressed or current.
	'color-brand': '#0088cc',
	'color-brand-hover': '#0077b3',
	'color-brand-text': '#fff',
	// The brand as a quiet fill: a selected row, a chosen chip.
	'color-brand-subtle': '#e5f3fa',

	'color-surface': '#fff',
	'color-body': '#fff',
	// The faintest tint: the header strip of a panel, and the wash under a
	// hovered row.
	'color-surface-alt': '#f7f7f8',
	'color-primary': 'var(--color-brand)',
	'color-primary-hover': 'var(--color-brand-hover)',
	'color-primary-text': 'var(--color-brand-text)',
	'color-secondary': '#f1f2f4',
	'color-secondary-alt': '#e6e8eb',
	'color-accent': 'var(--color-brand)',
	'color-accent-text': 'var(--color-brand-text)',

	// Text is near-black and the hint a dark grey: a light grey on white
	// reads as disabled rather than as secondary.
	'color-text': '#16191d',
	'color-text-hint': '#5b6168',
	'color-text-disabled': '#969ba1',

	// A field has no fill of its own and a hairline border, Telegram's: it
	// takes the colour of whatever it sits on — the page, a panel, a dialog
	// — so it never reads as a block of a different grey.
	'color-input': 'transparent',
	'color-input-border': '#d3d7dc',
	'color-input-border-hover': '#aab0b7',
	'color-border': '#e4e6e9',

	'color-success': '#31a24c',
	'color-danger': '#e53935',
	'color-warning': '#e8871e',
	'color-info': 'var(--color-brand)',

	'surface-success': '#e8f5eb',
	'surface-danger': '#fce9e9',
	'surface-info': 'var(--color-brand-subtle)',
	'surface-warning': '#fcf1e3',

	'shadow-sm': '0 1px 2px 0 rgb(0 0 0 / 8%)',
	// The drop under an open panel.
	'shadow-panel': '0 6px 16px -6px rgb(0 0 0 / 14%)',
	'shadow-md': '0 4px 20px 0 rgb(0 0 0 / 10%)',
	selection: '#b3dcf0',
	// What the page behind an open dialog is dimmed with.
	'color-overlay': 'rgb(0 0 0 / 30%)',

	'scrollbar-thumb': '#d4d6da',
	'scrollbar-thumb-hover': '#b6b9be',

	'tooltip-surface': '#2b2b2e',
	'tooltip-text': '#fff',

	// A toast is the theme turned over — black on the light page, white on
	// the dark one — so it is the one thing on screen that stands out without
	// a colour of its own. Its icon is the brand, whatever the toast says —
	// the shape tells a success from a failure — and it is the brand as the
	// other theme writes it, since that is what reads on black.
	'toast-surface': '#16191d',
	'toast-text': '#fff',
	'toast-hint': '#aab0b7',
	'toast-icon': '#3ea6e6',

	'row-hover': '#f7f7f8',

	// The sidebar's column, header block included: a step off the page, so
	// the column reads as a panel of its own, and a hover a step darker than
	// the usual so it still shows on that tint. A page's name is in the full
	// text colour — on white a grey name reads as disabled — and its icon a
	// step quieter. The page being read is the one row filled solid in the
	// brand: nothing else in the column is blue, so it is found at a glance.
	'nav-surface': 'var(--color-surface-alt)',
	'nav-hover': 'var(--color-secondary-alt)',
	'nav-text': 'var(--color-text)',
	'nav-icon': 'var(--color-text-hint)',
	'nav-current': 'var(--color-brand)',
	'nav-current-text': 'var(--color-brand-text)',

	// What floats over the page comes in two kinds, each with its own set: a
	// popover — a menu, a select's list — and a dialog — a full-window one, a modal,
	// the command palette. In the light theme both are white like the page,
	// set apart by a hairline, a shadow and, for a dialog, the dimmed page.
	'popover-surface': 'var(--color-surface)',
	'popover-surface-alt': 'var(--color-surface-alt)',
	'popover-secondary': 'var(--color-secondary)',
	'popover-secondary-alt': 'var(--color-secondary-alt)',
	'popover-border': 'var(--color-border)',
	'popover-input-border': 'var(--color-input-border)',
	'popover-input-border-hover': 'var(--color-input-border-hover)',
	'popover-row-hover': 'var(--row-hover)',

	'dialog-surface': 'var(--popover-surface)',
	'dialog-surface-alt': 'var(--popover-surface-alt)',
	'dialog-secondary': 'var(--popover-secondary)',
	'dialog-secondary-alt': 'var(--popover-secondary-alt)',
	'dialog-border': 'var(--popover-border)',
	'dialog-input-border': 'var(--popover-input-border)',
	'dialog-input-border-hover': 'var(--popover-input-border-hover)',
	'dialog-row-hover': 'var(--popover-row-hover)',

	// Code, highlighted (ui/CodeBlock): Shiki colours each token with one of
	// these, so a block of JSON follows the theme like everything else. In
	// JSON a key is a keyword, a string value a string, and a number, true,
	// false or null a constant.
	'syntax-foreground': 'var(--color-text)',
	'syntax-background': 'transparent',
	'syntax-token-keyword': '#0550ae',
	'syntax-token-string': '#0a7a3b',
	'syntax-token-string-expression': '#0a7a3b',
	'syntax-token-constant': '#b54708',
	'syntax-token-function': '#8250df',
	'syntax-token-parameter': '#953800',
	'syntax-token-comment': 'var(--color-text-hint)',
	'syntax-token-punctuation': 'var(--color-text-hint)',
	'syntax-token-link': 'var(--color-info)',

	// Read by `color-scheme` in base.css, so native controls — a date
	// picker, a scrollbar — follow the theme.
	'color-scheme': 'light'
} as const;

/** Every colour token, by name. */
export type ColorName = keyof typeof light;

/** The black the dark theme is built on: the page, and the step every
    surface above it is measured from. */
const black = '#000';

/**
 * The dark palette: only what differs from the light one.
 *
 * The page is true black, and its greys are cool and faintly blue (#16181c,
 * #202327, #2f3336). Elevation in a black theme is lightness, not shadow — a
 * shadow has nothing to fall on — so what floats is a small step lighter
 * than what it floats over, and every step comes from that same ramp rather
 * than a neutral grey beside it: a dialog and a menu read as the page's own
 * material, raised. The page is #000; a dialog sits at #0a0b0d, its panel
 * heads at #101215; a popover one step above at #0f1114, so a select opened
 * in a dialog still stands off it.
 */
const dark: Partial<Record<ColorName, string>> = {
	'color-brand-hover': '#1a9ad9',
	'color-brand-subtle': '#06202e',

	'color-surface': black,
	'color-body': black,
	'color-surface-alt': '#0a0b0c',
	'color-secondary': '#16181c',
	'color-secondary-alt': '#202327',

	'color-text': '#e7e9ea',
	'color-text-hint': '#8b9096',
	'color-text-disabled': '#50555a',

	'color-input-border': '#2f3336',
	'color-input-border-hover': '#4c5257',
	'color-border': '#202327',

	'color-success': '#4cbb5e',
	'color-danger': '#ef5350',
	'color-warning': '#f0a040',
	// Text-level brand, lifted so it reads on black; fills keep #0088cc.
	'color-info': '#3ea6e6',

	'surface-success': '#081a0f',
	'surface-danger': '#220b0b',
	'surface-warning': '#221808',

	// Almost no shadows: surfaces are set apart by their colour and a
	// hairline. The exception is what floats — a menu, a modal — whose
	// shadow falls on nothing over the black page but gives it an edge over
	// a dialog, which is lighter.
	'shadow-sm': 'none',
	'shadow-panel': 'none',
	'shadow-md': '0 12px 32px -8px rgb(0 0 0 / 70%)',
	selection: '#0a4d73',
	// The page behind a dialog is dimmed with black, the theme's own colour:
	// the page is black already, but its words, panels and borders are not,
	// and dimming them is what brings the sheet forward.
	'color-overlay': 'rgb(0 0 0 / 72%)',

	'scrollbar-thumb': '#2f3336',
	'scrollbar-thumb-hover': '#4c5257',

	'tooltip-surface': '#202327',

	'toast-surface': '#fff',
	'toast-text': '#16191d',
	'toast-hint': '#5b6168',
	'toast-icon': '#0088cc',

	'row-hover': '#0e1012',
	'nav-hover': 'var(--color-secondary)',

	'popover-surface': '#0f1114',
	'popover-surface-alt': '#15171a',
	'popover-secondary': '#1b1e22',
	'popover-secondary-alt': '#23262b',
	'popover-border': '#25282c',
	'popover-input-border': '#2f3336',
	'popover-input-border-hover': '#4c5257',
	'popover-row-hover': '#181b1f',

	'dialog-surface': '#0a0b0d',
	'dialog-surface-alt': '#101215',
	'dialog-secondary': '#16181c',
	'dialog-secondary-alt': '#1d2024',
	'dialog-border': '#1d2024',
	'dialog-input-border': '#2a2d31',
	'dialog-input-border-hover': '#43484d',
	'dialog-row-hover': '#111316',

	'syntax-token-keyword': '#79c0ff',
	'syntax-token-string': '#7ee787',
	'syntax-token-string-expression': '#7ee787',
	'syntax-token-constant': '#ffa657',
	'syntax-token-function': '#d2a8ff',
	'syntax-token-parameter': '#ffa657',

	'color-scheme': 'dark'
};

/* ---- The theme ------------------------------------------------------- */

export const theme: Theme<ColorName> = {
	fonts: {
		// Chirp is the console's typeface, bundled from assets/fonts so the
		// panel has the same face on every machine. It is X's own typeface
		// rather than an open family; see assets/fonts/README.md before
		// shipping it publicly.
		//
		// Chirp is static and covers Latin and Latin Extended. The panel's
		// 400, 500 and 700 weights are bundled, with 600 drawn from 700.
		// unicode-range keeps an English page to the smaller Latin file;
		// other scripts use the system fallback.
		baseUrl: '../../assets/fonts/',
		subsets: {
			'latin-ext':
				'U+0100-02BA, U+02BD-02C5, U+02C7-02CC, U+02CE-02D7, U+02DD-02FF, U+0304, U+0308, U+0329, U+1D00-1DBF, U+1E00-1E9F, U+1EF2-1EFF, U+2020, U+20A0-20AB, U+20AD-20C0, U+2113, U+2C60-2C7F, U+A720-A7FF',
			latin:
				'U+0000-00FF, U+0131, U+0152-0153, U+02BB-02BC, U+02C6, U+02DA, U+02DC, U+0304, U+0308, U+0329, U+2000-206F, U+20AC, U+2122, U+2191, U+2193, U+2212, U+2215, U+FEFF, U+FFFD'
		},
		faces: [
			{
				family: 'Chirp',
				style: 'normal',
				weights: [400, 500, 700],
				subsets: ['latin-ext', 'latin'],
				file: (subset, weight) => `chirp/chirp-${subset}-${weight}-normal.woff2`
			},
			// Google Sans Code is the face for code: ids, keys, JSON. It is an
			// open font (SIL OFL 1.1, the licence beside the files), taken from
			// the @fontsource/google-sans-code release at the same weights and
			// alphabets as Chirp.
			{
				family: 'Google Sans Code',
				style: 'normal',
				weights: [400, 500, 700],
				subsets: ['latin-ext', 'latin'],
				file: (subset, weight) =>
					`google-sans-code/google-sans-code-${subset}-${weight}-normal.woff2`
			},
			// Poppins is the display face: the header's wordmark, and nothing
			// that runs to a sentence. One weight in one alphabet is all a word
			// in English needs. Open (SIL OFL 1.1), from @fontsource/poppins.
			{
				family: 'Poppins',
				style: 'normal',
				weights: [600],
				subsets: ['latin'],
				file: (subset, weight) => `poppins/poppins-${subset}-${weight}-normal.woff2`
			}
		]
	},

	scales: [
		{
			prefix: 'font',
			tokens: {
				'system-sans': systemSans,
				'system-mono': systemMono,
				sans: "'Chirp', var(--font-system-sans)",
				mono: "'Google Sans Code', var(--font-system-mono)",
				display: "'Poppins', var(--font-system-sans)"
			}
		},
		{
			prefix: 'text',
			tokens: { xs: '12px', sm: '13px', base: '14px', lg: '15px', xl: '19px', '2xl': '24px' }
		},
		{ prefix: '', tokens: { 'line-height': '1.5' } },
		{
			prefix: 'radius',
			tokens: {
				sm: '6px',
				md: '10px',
				lg: '14px',
				pill: '999px',
				// The panel's three kinds of corner, which every field, control
				// and surface names. A field is the one thing drawn with a tight
				// corner, so a box to type in never reads as a button; a control
				// — a button, a chip, a filter — is a pill; a surface — a card, a
				// table, a dialog — is softly rounded. The sizes above are for
				// the small parts inside those: a menu's row, a checkbox, a tag
				// of code.
				field: 'var(--radius-sm)',
				control: 'var(--radius-pill)',
				surface: 'var(--radius-lg)'
			}
		},
		// On a 5px step, like PocketBase.
		{
			prefix: 'space',
			tokens: { 1: '5px', 2: '10px', 3: '15px', 4: '20px', 5: '30px', 6: '40px' }
		},
		{
			prefix: '',
			tokens: {
				// Every control is one of these heights: md is the default, and
				// what a toolbar row is built from — a button, an input, a
				// filter — so they line up without anyone measuring. sm is for
				// controls inside a row or a bar.
				'control-height': '45px',
				'control-height-sm': '35px',
				'control-height-lg': '52px',
				// A row in a list of places to go, rather than a control to
				// press: a column of 45px rows would be a column of buttons.
				'nav-item-height': '34px',
				'header-height': '55px',
				'sidebar-width': '240px',
				// Every dashboard page sits in one frame, centred and no wider
				// than this, so a wide monitor gets margins rather than a search
				// box two thousand pixels long.
				'content-max-width': '1400px',
				// How wide a form's column may grow inside that frame.
				'form-max-width': '760px',
				// The frame's inset from the window. It widens with the window:
				// see `responsive` below.
				'page-gutter': '16px',
				// How wide the column of a full-window dialog is: a form beside
				// the titles of its sections, not a line across the window.
				'dialog-column-width': '880px',
				// The thumb is drawn inside a transparent border, so the bar reads
				// thinner than the space it reserves (styles/scrollbars.css).
				'scrollbar-size': '10px'
			}
		},
		// Motion. A dialog arrives and leaves, and the sidebar slides, over
		// PocketBase's --modalAnimationSpeed.
		{ prefix: '', tokens: { 'speed-fast': '70ms', speed: '140ms', 'speed-slow': '200ms' } }
	],

	breakpoints: {
		phone: '34rem',
		small: '40rem',
		tablet: '48rem',
		// Below this the sidebar is a panel the header opens.
		sidebar: '55rem',
		laptop: '64rem',
		desktop: '80rem',
		wide: '90rem'
	},

	// The gutter widens with the window: tight on a phone, roomy on a monitor.
	responsive: [
		{ minWidth: '48rem', tokens: { 'page-gutter': '24px' } },
		{ minWidth: '90rem', tokens: { 'page-gutter': '32px' } }
	],

	colors: { light, dark },

	// Inside anything that floats, the page's surface tokens are swapped for
	// its set, so every component in it — panels, rows, fields, dividers — is
	// drawn at that level without knowing where it is.
	layers: {
		swapped: {
			'color-surface': 'surface',
			'color-surface-alt': 'surface-alt',
			'color-secondary': 'secondary',
			'color-secondary-alt': 'secondary-alt',
			'color-border': 'border',
			'color-input-border': 'input-border',
			'color-input-border-hover': 'input-border-hover',
			'row-hover': 'row-hover'
		},
		sets: [
			{
				prefix: 'popover',
				selectors: [
					"[data-scope='menu'][data-part='content']",
					"[data-scope='select'][data-part='content']",
					"[data-scope='popover'][data-part='content']"
				]
			},
			{ prefix: 'dialog', selectors: ["[data-scope='dialog'][data-part='content']"] }
		]
	}
};

export const breakpoints = theme.breakpoints;
