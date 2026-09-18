<script lang="ts">
	import { untrack } from 'svelte';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { RiArrowGoBackLine } from 'svelte-remixicon';
	import { ApiError, languagesApi, type Language, type ShippedLanguage } from '$lib/api';
	import {
		Alert,
		Button,
		Drawer,
		FormSection,
		Icon,
		Input,
		Select,
		SwitchField
	} from '$lib/components/ui';
	import { useTranslator } from '$lib/i18n';
	import { keys } from '$lib/query';
	import { languageCode, namesOf } from './translations';

	type Props = {
		open: boolean;
		/** The languages there are, to copy one and to refuse a code twice. */
		languages: Language[];
		/** The shipped languages this installation does not have. */
		shipped: ShippedLanguage[];
		/** Called with the language once it exists, so the page can open it. */
		onCreated: (language: Language) => void;
	};

	let { open = $bindable(false), languages, shipped, onCreated }: Props = $props();

	const queryClient = useQueryClient();
	const t = useTranslator();

	let code = $state('');
	let name = $state('');
	let native = $state('');
	/** What the new language's text starts as: a language's code, or
	    NOTHING. Not an empty string, which a Select reads as unchosen. */
	const NOTHING = '-';

	let copyFrom = $state(NOTHING);
	let enabled = $state(false);

	/** Whether the names were typed rather than filled in from the code: a
	    name somebody wrote is never replaced behind their back. */
	let namesTyped = $state(false);

	let error = $state('');
	let saving = $state(false);

	$effect(() => {
		if (!open) return;

		code = '';
		name = '';
		native = '';
		copyFrom = NOTHING;
		enabled = false;
		namesTyped = false;
		error = '';
	});

	const trimmed = $derived(code.trim());
	const shippedMatch = $derived(shipped.find((one) => one.code === trimmed));

	const codeError = $derived.by(() => {
		if (trimmed === '') return undefined;
		if (!languageCode.test(trimmed)) return t('languages.code_invalid');
		if (languages.some((one) => one.code.toLowerCase() === trimmed.toLowerCase())) {
			return t('languages.code_taken');
		}
		return undefined;
	});

	const ready = $derived(
		trimmed !== '' && codeError === undefined && name.trim() !== '' && native.trim() !== ''
	);

	// Typing a code fills in its names, the way the browser spells them.
	$effect(() => {
		const current = trimmed;
		untrack(() => fillNames(current));
	});

	function fillNames(current: string) {
		if (namesTyped) return;

		const match = shipped.find((one) => one.code === current);
		const names = languageCode.test(current) ? namesOf(current) : null;
		name = match?.name ?? names?.name ?? '';
		native = match?.native ?? names?.native ?? '';

		// A code the server ships a translation of starts from it, unless
		// somebody has already said otherwise.
		if (match && copyFrom === NOTHING) copyFrom = match.code;
		if (!match && shipped.some((one) => one.code === copyFrom)) copyFrom = NOTHING;
	}

	function restore(language: ShippedLanguage) {
		code = language.code;
		name = language.name;
		native = language.native;
		copyFrom = language.code;
		namesTyped = false;
	}

	const sources = $derived([
		{ value: NOTHING, label: t('languages.start_empty') },
		...(shippedMatch
			? [
					{
						value: shippedMatch.code,
						label: t('languages.start_shipped'),
						description: shippedMatch.native
					}
				]
			: []),
		...languages.map((language) => ({
			value: language.code,
			label: t('languages.start_copy', { language: language.native }),
			description: language.name
		}))
	]);

	const create = createMutation(() => ({
		mutationFn: () =>
			languagesApi.create({
				code: trimmed,
				name: name.trim(),
				native: native.trim(),
				enabled,
				copy_from: copyFrom === NOTHING ? undefined : copyFrom
			}),
		onSuccess: async ({ language }) => {
			open = false;
			await queryClient.invalidateQueries({ queryKey: keys.languages.list });
			onCreated(language);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : t('languages.create_failed');
		},
		onSettled: () => {
			saving = false;
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (!ready || saving) return;

		error = '';
		saving = true;
		create.mutate();
	}
</script>

<Drawer
	bind:open
	title={t('languages.new_title')}
	description={t('languages.new_description')}
	width="36rem"
	onsubmit={submit}
>
	{#if error}
		<div class="message"><Alert>{error}</Alert></div>
	{/if}

	{#if shipped.length > 0}
		<FormSection title={t('languages.restore')} description={t('languages.restore_hint')}>
			<div class="restore">
				{#each shipped as language (language.code)}
					<Button size="sm" variant="outline" onclick={() => restore(language)}>
						<Icon icon={RiArrowGoBackLine} />
						<span lang={language.code}>{language.native}</span>
					</Button>
				{/each}
			</div>
		</FormSection>
	{/if}

	<FormSection title={t('languages.names')}>
		<Input
			label={t('languages.code')}
			bind:value={code}
			hint={t('languages.code_hint')}
			error={codeError}
			placeholder="uz"
			autocomplete="off"
			spellcheck={false}
			required
		/>

		<div class="pair">
			<Input
				label={t('languages.name')}
				bind:value={name}
				oninput={() => (namesTyped = true)}
				hint={t('languages.name_hint')}
				required
			/>
			<Input
				label={t('languages.native')}
				bind:value={native}
				oninput={() => (namesTyped = true)}
				hint={t('languages.native_hint')}
				lang={trimmed || undefined}
				required
			/>
		</div>
	</FormSection>

	<FormSection title={t('languages.start_from')}>
		<Select label={t('languages.start_from')} bind:value={copyFrom} options={sources} />

		<SwitchField
			label={t('languages.offer_now')}
			description={t('languages.offer_now_hint')}
			bind:checked={enabled}
		/>
	</FormSection>

	{#snippet footer()}
		<div class="actions">
			<Button variant="subtle" onclick={() => (open = false)}>{t('action.cancel')}</Button>
			<Button type="submit" loading={saving} disabled={!ready || saving}>
				{t('languages.create')}
			</Button>
		</div>
	{/snippet}
</Drawer>

<style>
	.message {
		margin-bottom: var(--space-4);
	}

	.restore {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-2);
	}

	.pair {
		display: grid;
		grid-template-columns: 1fr 1fr;
		align-items: start;
		gap: var(--space-3);
	}

	.actions {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		margin-left: auto;
	}

	@media (max-width: 36rem) {
		.pair {
			grid-template-columns: 1fr;
		}
	}
</style>
