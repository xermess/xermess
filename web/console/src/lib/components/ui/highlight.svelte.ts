/**
 * One shared Shiki highlighter, loaded lazily so pages without code never download it. It uses
 * Shiki core with JSON only and the JavaScript regex engine (no WebAssembly); colours are
 * `--syntax-*` tokens from lib/theme/theme.ts.
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
 * The highlighted markup of `source()`, updated as it changes; call during component init.
 * Updates are coalesced to one per frame and stale results dropped. `html` is '' until the
 * first result and on the server.
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
