<script lang="ts">
	import { page } from '$app/state';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { RiDeleteBinLine, RiSettings3Line, RiTranslate2 } from 'svelte-remixicon';
	import { ApiError, languagesApi, type Language, type LocaleApp } from '$lib/api';
	import {
		Alert,
		Button,
		Drawer,
		FormSection,
		Icon,
		Input,
		SwitchField,
		Tabs
	} from '$lib/components/ui';
	import { useTranslator } from '$lib/i18n';
	import { keys } from '$lib/query';
	import { reloadText } from '$lib/state/language.svelte';
	import TranslationEditor from './TranslationEditor.svelte';

	type Props = {
		open: boolean;
		/** The language being changed. */
		language?: Language | null;
		apps: LocaleApp[];
		/** Whether the administrator may change it, or only read it. */
		canWrite: boolean;
		/** The tab it opens on: its settings, or one app's text. */
		tab?: 'settings' | LocaleApp;
	};

	let {
		open = $bindable(false),
		language = null,
		apps,
		canWrite,
		tab: startingTab = 'settings'
	}: Props = $props();

	const queryClient = useQueryClient();
	const t = useTranslator();

	let tab = $state<string>('settings');
	let name = $state('');
	let native = $state('');
	let enabled = $state(false);
	let isDefault = $state(false);
	let position = $state('');

	/** Each app's text as it is being edited, and whether it differs from
	    what is saved. An app whose tab was never opened has neither. */
	let drafts = $state<Record<string, Record<string, string> | null>>({});
	let dirty = $state<Record<string, boolean>>({});

	let error = $state('');
	let saving = $state(false);
	let confirmingDelete = $state(false);

	// The form is filled in each time the panel opens.
	$effect(() => {
		if (!open || !language) return;

		tab = startingTab;
		error = '';
		confirmingDelete = false;
		name = language.name;
		native = language.native;
		enabled = language.enabled;
		isDefault = language.is_default;
		position = String(language.position);
		drafts = Object.fromEntries(apps.map((app) => [app, null]));
		dirty = {};
	});

	/** The base language is always offered, and so is the default: one is what
	    every missing translation falls back to, the other what somebody sees
	    before choosing. */
	const locked = $derived(language?.base === true || isDefault);

	const settingsChanged = $derived(
		language !== null &&
			(name.trim() !== language.name ||
				native.trim() !== language.native ||
				enabled !== language.enabled ||
				isDefault !== language.is_default ||
				Number(position) !== language.position)
	);

	const textChanged = $derived(apps.filter((app) => dirty[app]));

	const ready = $derived(
		name.trim() !== '' && native.trim() !== '' && Number.isInteger(Number(position))
	);

	function appName(app: LocaleApp): string {
		return app === 'id' ? t('languages.sign_in_pages') : t('languages.admin_panel');
	}

	const tabs = $derived([
		{ value: 'settings', label: t('languages.tab_settings'), icon: RiSettings3Line },
		...apps.map((app) => ({
			value: app,
			label: appName(app),
			icon: RiTranslate2,
			count: language && language.missing[app] > 0 ? language.missing[app] : undefined
		}))
	]);

	/**
	 * Saves whatever changed: the settings, then each app's text. They are
	 * separate requests because they are separate things — renaming a
	 * language is not rewriting it — but they are one button, because
	 * somebody who has changed both means to keep both.
	 */
	const save = createMutation(() => ({
		mutationFn: async () => {
			const code = language!.code;

			if (settingsChanged) {
				await languagesApi.update(code, {
					name: name.trim(),
					native: native.trim(),
					enabled,
					is_default: isDefault,
					position: Number(position)
				});
			}

			for (const app of textChanged) {
				await languagesApi.saveTranslation(code, app, drafts[app] ?? {});
			}

			return code;
		},
		onSuccess: async (code) => {
			open = false;

			// The panel is drawn in one of these languages, so editing the
			// one it is drawn in — its text or its name in the picker —
			// redraws it, the way it would for anybody opening it next.
			const own = code === page.data.language;

			// The editor seeds itself from its cached text, so what was cached
			// is dropped rather than refetched behind it: the next one opened
			// starts from what is saved now. English is every editor's left
			// column, which is why it is all of them and not only these.
			queryClient.removeQueries({ queryKey: keys.languages.translations });

			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.languages.list }),
				own || textChanged.includes('console') ? reloadText() : undefined
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : t('languages.save_failed');
		},
		onSettled: () => {
			saving = false;
		}
	}));

	const remove = createMutation(() => ({
		mutationFn: () => languagesApi.remove(language?.code ?? ''),
		onSuccess: async () => {
			open = false;
			queryClient.removeQueries({ queryKey: keys.languages.translations });

			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.languages.list }),
				reloadText()
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : t('languages.remove_failed');
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (!ready || saving || (!settingsChanged && textChanged.length === 0)) return;

		error = '';
		saving = true;
		save.mutate();
	}

	/** Making a language the default is also offering it. */
	function defaultChanged(on: boolean) {
		if (on) enabled = true;
	}
</script>

<Drawer
	bind:open
	title={language?.native ?? ''}
	description={t('languages.drawer_description')}
	meta={language?.code}
	width="56rem"
	onsubmit={submit}
>
	{#if error}
		<div class="message"><Alert>{error}</Alert></div>
	{/if}

	{#if language}
		<Tabs {tabs} bind:value={tab} label={t('languages.title')}>
			{#snippet panel(value)}
				<div class="panel">
					{#if value === 'settings'}
						<FormSection title={t('languages.names')}>
							<div class="pair">
								<Input
									label={t('languages.name')}
									bind:value={name}
									hint={t('languages.name_hint')}
									readOnly={!canWrite}
									required
								/>
								<Input
									label={t('languages.native')}
									bind:value={native}
									hint={t('languages.native_hint')}
									readOnly={!canWrite}
									lang={language.code}
									required
								/>
							</div>
						</FormSection>

						<FormSection title={t('languages.coverage')}>
							<div class="coverage">
								{#each apps as app (app)}
									{@const percent = language.coverage[app] ?? 0}
									<div class="row">
										<span class="app">{appName(app)}</span>
										<span class="track">
											<span class="fill" class:whole={percent === 100} style="--filled: {percent}%"
											></span>
										</span>
										<span class="value">
											{percent}%
											{#if language.missing[app] > 0}
												<small>· {t('languages.missing', { count: language.missing[app] })}</small>
											{/if}
										</span>
									</div>
								{/each}
							</div>

							<p class="note">
								{#if language.base}
									{t('languages.base_note')}
								{:else if language.shipped}
									{t('languages.shipped_note')}
								{:else}
									{t('languages.custom_note')}
								{/if}
							</p>
						</FormSection>

						<FormSection title={t('languages.availability')}>
							<SwitchField
								label={t('languages.make_default')}
								description={t('languages.make_default_hint')}
								bind:checked={isDefault}
								onChange={defaultChanged}
								disabled={!canWrite || language.is_default}
							/>

							<SwitchField
								label={t('languages.offer')}
								description={locked ? t('languages.default_locked') : t('languages.offer_hint')}
								bind:checked={enabled}
								disabled={!canWrite || locked}
							/>

							<div class="position">
								<Input
									label={t('languages.position')}
									bind:value={position}
									hint={t('languages.position_hint')}
									type="number"
									min="0"
									step="1"
									readOnly={!canWrite}
								/>
							</div>
						</FormSection>
					{:else}
						{@const app = value as LocaleApp}
						<TranslationEditor
							code={language.code}
							name={name || language.name}
							native={native || language.native}
							{app}
							readOnly={!canWrite}
							bind:draft={drafts[app]}
							onDirty={(changed) => (dirty[app] = changed)}
						/>
					{/if}
				</div>
			{/snippet}
		</Tabs>
	{/if}

	{#snippet footer()}
		{#if canWrite && language && !language.base && !language.is_default}
			{#if confirmingDelete}
				<div class="confirm">
					<span>{t('languages.remove_confirm')}</span>
					<Button variant="subtle" size="sm" onclick={() => (confirmingDelete = false)}>
						{t('languages.keep')}
					</Button>
					<Button colorPalette="danger" size="sm" onclick={() => remove.mutate()}>
						{t('languages.remove')}
					</Button>
				</div>
			{:else}
				<Button
					colorPalette="danger"
					variant="subtle"
					size="sm"
					onclick={() => (confirmingDelete = true)}
				>
					<Icon icon={RiDeleteBinLine} />
					{t('languages.remove')}
				</Button>
			{/if}
		{/if}

		<div class="actions">
			{#if textChanged.length > 0}
				<span class="unsaved">
					{t('languages.unsaved', { apps: textChanged.map(appName).join(', ') })}
				</span>
			{/if}
			<Button variant="subtle" onclick={() => (open = false)}>{t('action.cancel')}</Button>
			{#if canWrite}
				<Button
					type="submit"
					loading={saving}
					disabled={!ready || saving || (!settingsChanged && textChanged.length === 0)}
				>
					{t('action.save_changes')}
				</Button>
			{/if}
		</div>
	{/snippet}
</Drawer>

<style>
	.message {
		margin-bottom: var(--space-4);
	}

	.panel {
		padding-top: var(--space-4);
	}

	.pair {
		display: grid;
		grid-template-columns: 1fr 1fr;
		align-items: start;
		gap: var(--space-3);
	}

	.coverage {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
	}

	.row {
		display: grid;
		grid-template-columns: 10rem 1fr auto;
		align-items: center;
		gap: var(--space-3);
	}

	.app {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.track {
		height: 8px;
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
	}

	.fill.whole {
		background: var(--color-success);
	}

	.value {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		font-variant-numeric: tabular-nums;
		white-space: nowrap;
	}

	.note {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		line-height: 1.5;
	}

	.position {
		max-width: 16rem;
	}

	.confirm {
		display: flex;
		align-items: center;
		flex-wrap: wrap;
		gap: var(--space-2);
		margin-right: auto;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.actions {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		margin-left: auto;
	}

	.unsaved {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	@media (max-width: 36rem) {
		.pair,
		.row {
			grid-template-columns: 1fr;
		}

		.row {
			gap: var(--space-1);
		}

		.unsaved {
			display: none;
		}
	}
</style>
