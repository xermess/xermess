/**
 * Code, highlighted: one Shiki highlighter for the whole panel, built the
 * first time a block of code is drawn and shared after that.
 *
 * Shiki is imported only then, so it is a chunk of its own that a page
 * without code never downloads. It is Shiki's core with the JSON grammar and
 * the JavaScript regex engine — no WebAssembly, and no grammar the panel does
 * not show. The theme is Shiki's CSS-variables one: every token is coloured
 * with a `--syntax-*` variable, and those are in lib/theme/theme.ts with the
 * rest of the palette, light and dark.
 */
import type { HighlighterCore } from 'shiki/core';

let loading: Promise<HighlighterCore> | undefined;

function load(): Promise<HighlighterCore> {
	loading ??= (async () => {
		const [{ createCssVariablesTheme, createHighlighterCore }, { createJavaScriptRegexEngine }] =
			await Promise.all([import('shiki/core'), import('shiki/engine/javascript')]);

		return createHighlighterCore({
			themes: [createCssVariablesTheme({ name: 'panel', variablePrefix: '--syntax-' })],
			langs: [import('shiki/langs/json.mjs')],
			engine: createJavaScriptRegexEngine()
		});
	})();
	return loading;
}

/** JSON as a `<pre class="shiki">` of coloured spans, or '' if it cannot be
    highlighted — a block left uncoloured is still the code. */
async function highlight(code: string): Promise<string> {
	try {
		return (await load()).codeToHtml(code, { lang: 'json', theme: 'panel' });
	} catch {
		return '';
	}
}

/**
 * The highlighted markup of `source()`, kept up to date as it changes. Call
 * it while a component initialises.
 *
 * Changes are coalesced to one highlight per frame, so a burst of them — a
 * field being typed in — costs one pass rather than one per change, and a
 * result that arrives after a newer change is dropped. Until the first
 * result, and on the server, `html` is '': the caller shows the code plain.
 */
export function highlighted(source: () => string): { readonly html: string } {
	let html = $state('');

	$effect(() => {
		const code = source();
		let current = true;

		const frame = requestAnimationFrame(() => {
			highlight(code).then((result) => {
				if (current) html = result;
			});
		});

		return () => {
			current = false;
			cancelAnimationFrame(frame);
		};
	});

	return {
		get html() {
			return html;
		}
	};
}
