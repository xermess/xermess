<script lang="ts">
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { RiFileTextLine, RiRestartLine, RiTranslate2 } from 'svelte-remixicon';
	import {
		mailApi,
		type MailContent,
		type MailLanguageContent,
		type MailMessageSpec
	} from '$lib/api';
	import {
		Button,
		Code,
		FormSection,
		Input,
		SaveBar,
		Select,
		Tag,
		Textarea,
		notify
	} from '$lib/components/ui';
	import { keys } from '$lib/query';

	type Props = {
		content: MailContent;
	};

	let { content }: Props = $props();

	const queryClient = useQueryClient();

	/** Which language is being written. English first, since it is what every
	    other language falls back to. */
	// svelte-ignore state_referenced_locally
	let code = $state(content.languages[0]?.code ?? 'en');

	const language = $derived(
		content.languages.find((one) => one.code === code) ?? content.languages[0]
	);

	const choices = $derived(
		content.languages.map((one) => ({
			value: one.code,
			label: `${one.name} (${one.code})`,
			description: one.offered ? undefined : 'Not offered on the sign-in pages'
		}))
	);

	/** One language's words as the fields hold them: what it says, and the
	    empty string where it says nothing — which is how a field is sent to
	    clear an override, so what is on screen and what the server stores are
	    the same thing. */
	function wordsOf(one: MailLanguageContent | undefined): Record<string, string> {
		const words: Record<string, string> = {};

		for (const spec of content.messages) {
			words[spec.subject_key] = one?.messages[spec.subject_key] ?? '';
			words[spec.body_key] = one?.messages[spec.body_key] ?? '';
		}

		return words;
	}

	/** What is being typed, by key. Filled in before the first render rather
	    than in an effect: a field cannot bind to a key that is not there yet. */
	// svelte-ignore state_referenced_locally
	let drafts = $state(wordsOf(language));

	/** The language the fields were last filled from. The picker changing
	    fills them again — switching away and back shows what is stored, not
	    what was half-typed for somebody else's language — and a refetch in the
	    background leaves what is being typed alone. */
	// svelte-ignore state_referenced_locally
	let filledFrom = $state(code);

	$effect(() => {
		if (code === filledFrom) return;

		filledFrom = code;
		drafts = wordsOf(content.languages.find((one) => one.code === code));
	});

	const dirty = $derived(
		content.messages.some(
			(spec) =>
				drafts[spec.subject_key] !== (language?.messages[spec.subject_key] ?? '') ||
				drafts[spec.body_key] !== (language?.messages[spec.body_key] ?? '')
		)
	);

	const save = createMutation(() => ({
		mutationFn: () => mailApi.saveContent(code, drafts),
		onSuccess: async () => {
			notify.success('Saved', 'The next message in this language uses these words.');
			// The words are the sign-in pages' text, so the languages go with
			// the mail content: both are reading the same rows.
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.mail.content }),
				queryClient.invalidateQueries({ queryKey: keys.languages.all }),
				queryClient.invalidateQueries({ queryKey: keys.admin.overview })
			]);
		},
		onError: (err: unknown) => {
			notify.error(err, 'Could not save these words');
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (save.isPending || !dirty) return;

		save.mutate();
	}

	/** Puts one message back to the text this release shipped, by clearing
	    the language's own — the server stores nothing for an empty value, so
	    the shipped words come back. */
	function revert(spec: MailMessageSpec) {
		drafts = { ...drafts, [spec.subject_key]: '', [spec.body_key]: '' };
	}

	/** What a field shows when nothing has been written for it: the English
	    that will actually be sent. */
	function shipped(key: string): string {
		return language?.base[key] ?? '';
	}

	function written(spec: MailMessageSpec): boolean {
		return drafts[spec.subject_key] !== '' || drafts[spec.body_key] !== '';
	}
</script>

<form onsubmit={submit}>
	<FormSection
		title="Language"
		description="An email goes out in the reader's language. A field left empty falls back to the English this release ships, which is what its placeholder shows."
		icon={RiTranslate2}
	>
		{#snippet meta()}
			{#if language && !language.offered}
				<Tag tone="warning" small>Not offered</Tag>
			{/if}
		{/snippet}

		<Select
			label="Writing in"
			bind:value={code}
			options={choices}
			hint="The same text the Languages page holds. Add and remove languages there."
		/>
	</FormSection>

	{#each content.messages as spec (spec.kind)}
		<FormSection title={spec.label} description={spec.description} icon={RiFileTextLine}>
			{#snippet meta()}
				{#if written(spec)}
					<Tag tone="info" small>Written</Tag>
				{:else}
					<Tag small>Shipped text</Tag>
				{/if}
			{/snippet}

			{#snippet action()}
				{#if written(spec)}
					<Button size="sm" variant="subtle" onclick={() => revert(spec)}>
						<RiRestartLine size="15" />
						Use the shipped text
					</Button>
				{/if}
			{/snippet}

			<Input
				label="Subject"
				bind:value={drafts[spec.subject_key]}
				maxlength={200}
				placeholder={shipped(spec.subject_key)}
			/>

			<Textarea
				label="Body"
				bind:value={drafts[spec.body_key]}
				rows={8}
				placeholder={shipped(spec.body_key)}
			/>

			<p class="params">
				{#each spec.params as param (param.name)}
					<Code tone="quiet" title={param.description}>{'{' + param.name + '}'}</Code>
				{/each}
				<span>are filled in when the message is sent.</span>
			</p>
		</FormSection>
	{/each}

	{#if dirty}
		<SaveBar saving={save.isPending} ondiscard={() => (drafts = wordsOf(language))} />
	{/if}
</form>

<style>
	/* FormSection spaces the sections itself; the bar needs a gap above it. */
	form > :global(.save-bar) {
		margin-top: var(--space-5);
	}

	/* The placeholders a message may use, each readable on its own: a reader
	   should be able to copy one out of the line. */
	.params {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-1) var(--space-2);
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}
</style>
