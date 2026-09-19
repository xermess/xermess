<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import {
		RiDownload2Line,
		RiErrorWarningLine,
		RiSearchLine,
		RiUpload2Line
	} from 'svelte-remixicon';
	import type { LocaleApp } from '$lib/api';
	import { Alert, Button, Icon, Switch } from '$lib/components/ui';
	import { useTranslator } from '$lib/i18n';
	import { nest } from '$lib/i18n/flatten';
	import { translationOptions } from '$lib/query';
	import { download, missingParameters, readTranslationFile } from './translations';

	type Props = {
		code: string;
		/** The language's names, which an exported file carries so whoever
		    translates it knows what it is. */
		name: string;
		native: string;
		app: LocaleApp;
		readOnly?: boolean;
		/** The text as it is being edited, by key: null until the saved text
		    has arrived. The drawer saves it. */
		draft: Record<string, string> | null;
		/** Told whether the draft differs from what is saved, each time that
		    changes, so the drawer knows what it has to save. */
		onDirty: (dirty: boolean) => void;
	};

	let {
		code,
		name,
		native,
		app,
		readOnly = false,
		draft = $bindable(null),
		onDirty
	}: Props = $props();

	const t = useTranslator();

	const saved = createQuery(() => translationOptions(code, app));

	// The draft starts as what is saved, once that has arrived, and is the
	// editor's from then on: a refetch behind it must not wipe what is typed.
	$effect(() => {
		if (draft === null && saved.data) draft = { ...saved.data.messages };
	});

	$effect(() => {
		const original = saved.data?.messages ?? {};
		const current = draft ?? original;

		const keys = new Set([...Object.keys(original), ...Object.keys(current)]);
		onDirty([...keys].some((key) => (original[key] ?? '').trim() !== (current[key] ?? '').trim()));
	});

	let search = $state('');
	let onlyMissing = $state(false);

	const translated = (key: string) => (draft?.[key] ?? '').trim() !== '';

	const total = $derived(saved.data?.keys.length ?? 0);
	const done = $derived(saved.data?.keys.filter(translated).length ?? 0);

	/** The keys left once the search and the filter have had their say. The
	    search reads the key, the English and the translation, so somebody can
	    find a sentence by any of the three. */
	const visible = $derived.by(() => {
		if (!saved.data) return [];

		const term = search.trim().toLowerCase();
		const base = saved.data.base;

		return saved.data.keys.filter((key) => {
			if (onlyMissing && translated(key)) return false;
			if (term === '') return true;

			return [key, base[key] ?? '', draft?.[key] ?? ''].some((value) =>
				value.toLowerCase().includes(term)
			);
		});
	});

	function edit(key: string, value: string) {
		draft = { ...draft, [key]: value };
	}

	let notice = $state<{ tone: 'success' | 'danger'; text: string } | null>(null);
	let fileInput = $state<HTMLInputElement>();

	/** A translation file, merged over what is there: every key it has text
	    for replaces this language's, and the rest are left alone — so a file
	    covering half the app finishes half of it rather than emptying the
	    other half. */
	async function importFile(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		input.value = '';

		if (!file || !saved.data) return;

		const result = await readTranslationFile(file, saved.data.keys);
		if (!result.ok) {
			notice = { tone: 'danger', text: t('languages.import_invalid', { file: file.name }) };
			return;
		}

		draft = { ...draft, ...result.messages };

		const imported = t('languages.imported', {
			count: Object.keys(result.messages).length,
			file: file.name
		});
		notice = {
			tone: 'success',
			text:
				result.skipped > 0
					? `${imported} ${t('languages.imported_skipped', { count: result.skipped })}`
					: imported
		};
	}

	/** Every key, translated or not, so the file is also the template for
	    whoever translates the rest: an empty value is one still to do. It is
	    nested by screen, the shape of the files under locales/, so it can be
	    dropped in there as it is. */
	function exportFile() {
		if (!saved.data) return;

		const body: Record<string, string> = { $name: name, $native: native };
		for (const key of saved.data.keys) body[key] = draft?.[key] ?? '';

		download(`${code}.${app}.json`, JSON.stringify(nest(body), null, 2) + '\n');
	}
</script>

<div class="editor">
	{#if saved.isPending}
		<p class="state">{t('languages.loading')}</p>
	{:else if saved.isError || !saved.data}
		<Alert>{t('languages.load_failed')}</Alert>
	{:else}
		<div class="toolbar">
			<label class="search">
				<Icon icon={RiSearchLine} />
				<input
					type="search"
					placeholder={t('languages.search_text')}
					bind:value={search}
					aria-label={t('action.search')}
				/>
			</label>

			<Switch label={t('languages.only_missing')} bind:checked={onlyMissing} />
		</div>

		<div class="summary">
			<span class="progress">
				<span class="track">
					<span class="fill" class:whole={done === total} style="--filled: {(done * 100) / total}%"
					></span>
				</span>
				{t('languages.progress', { done, total })}
			</span>

			<span class="files">
				{#if !readOnly}
					<input
						bind:this={fileInput}
						class="file"
						type="file"
						accept=".json,application/json"
						onchange={importFile}
						tabindex="-1"
						aria-hidden="true"
					/>
					<Button size="sm" variant="subtle" onclick={() => fileInput?.click()}>
						<Icon icon={RiUpload2Line} />
						{t('languages.import')}
					</Button>
				{/if}
				<Button size="sm" variant="subtle" onclick={exportFile}>
					<Icon icon={RiDownload2Line} />
					{t('languages.export')}
				</Button>
			</span>
		</div>

		{#if notice}
			<div class="notice"><Alert tone={notice.tone}>{notice.text}</Alert></div>
		{/if}

		{#if visible.length === 0}
			<p class="state">{t('languages.no_match')}</p>
		{:else}
			<ol class="keys">
				{#each visible as key (key)}
					{@const english = saved.data.base[key] ?? ''}
					{@const value = draft?.[key] ?? ''}
					{@const lost = value.trim() ? missingParameters(english, value) : []}
					<li class:missing={!value.trim()}>
						<div class="source">
							<code>{key}</code>
							<p lang="en">{english}</p>
						</div>

						<div class="target">
							<textarea
								{value}
								lang={code}
								rows="1"
								placeholder={english}
								readonly={readOnly}
								aria-label={key}
								oninput={(event) => edit(key, event.currentTarget.value)}></textarea>

							{#if lost.length > 0}
								<small class="warning">
									<Icon icon={RiErrorWarningLine} />
									{t('languages.placeholder_missing', { names: lost.join(', ') })}
								</small>
							{/if}
						</div>
					</li>
				{/each}
			</ol>
		{/if}
	{/if}
</div>

<style>
	.editor {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}

	.state {
		margin: var(--space-4) 0;
		color: var(--color-text-hint);
		text-align: center;
	}

	.toolbar {
		display: flex;
		align-items: center;
		gap: var(--space-3);
	}

	.search {
		flex: 1;
		display: flex;
		align-items: center;
		gap: var(--space-2);
		height: var(--control-height);
		padding: 0 13px;
		border-radius: var(--radius-sm);
		background: var(--color-input);
		color: var(--color-text-hint);
		transition: background-color var(--speed-fast);
	}

	.search:focus-within {
		background: var(--color-input-focus);
	}

	.search input {
		flex: 1;
		min-width: 0;
		border: none;
		background: transparent;
		color: var(--color-text);
		font: inherit;
		font-size: var(--text-base);
	}

	.search input:focus {
		outline: none;
	}

	.summary {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: var(--space-2);
	}

	.progress {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		font-variant-numeric: tabular-nums;
	}

	.track {
		display: block;
		width: 6rem;
		height: 6px;
		border-radius: var(--radius-pill);
		background: var(--color-secondary-alt);
		overflow: hidden;
	}

	.fill {
		display: block;
		width: var(--filled);
		height: 100%;
		border-radius: inherit;
		background: var(--color-text-hint);
		transition: width var(--speed-fast);
	}

	.fill.whole {
		background: var(--color-success);
	}

	.files {
		display: inline-flex;
		gap: var(--space-2);
	}

	.file {
		display: none;
	}

	.keys {
		display: flex;
		flex-direction: column;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	/* The English on the left, the translation beside it: the way a translator
	   reads, one sentence against the other. */
	li {
		display: grid;
		grid-template-columns: minmax(0, 1fr) minmax(0, 1.2fr);
		align-items: start;
		gap: var(--space-3);
		padding: var(--space-3) 0;
		border-top: 1px solid var(--color-border);
	}

	.source {
		display: flex;
		flex-direction: column;
		gap: 2px;
		min-width: 0;
	}

	code {
		color: var(--color-text-hint);
		font-family: var(--font-mono);
		font-size: var(--text-xs);
		overflow-wrap: anywhere;
	}

	.source p {
		margin: 0;
		line-height: 1.45;
		overflow-wrap: anywhere;
	}

	.target {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
	}

	textarea {
		width: 100%;
		min-height: var(--control-height);
		padding: 10px 13px;
		border: 1px solid transparent;
		border-radius: var(--radius-sm);
		background: var(--color-input);
		color: var(--color-text);
		font: inherit;
		font-size: var(--text-base);
		line-height: 1.45;
		resize: vertical;
		/* Grows with what is typed, where the browser can; elsewhere the
		   handle in the corner does. */
		field-sizing: content;
		transition: background-color var(--speed-fast);
	}

	textarea:focus {
		background: var(--color-input-focus);
		outline: none;
	}

	textarea::placeholder {
		color: var(--color-text-hint);
		opacity: 0.7;
	}

	/* A key with nothing written is the one somebody is looking for. */
	li.missing textarea {
		border-color: var(--color-border);
		border-style: dashed;
		background: transparent;
	}

	li.missing textarea:focus {
		background: var(--color-input-focus);
	}

	.warning {
		display: inline-flex;
		align-items: center;
		gap: var(--space-1);
		color: var(--color-warning);
		font-size: var(--text-sm);
	}

	@media (max-width: 40rem) {
		li {
			grid-template-columns: 1fr;
			gap: var(--space-2);
		}
	}
</style>
