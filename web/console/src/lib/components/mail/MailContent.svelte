<script lang="ts">
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { RiFileTextLine, RiRestartLine, RiTranslate2 } from 'svelte-remixicon';
	import {
		ApiError,
		mailApi,
		type MailContent,
		type MailLanguageContent,
		type MailMessageSpec
	} from '$lib/api';
	import { Alert, Button, Input, Panel, Select, Tag, Textarea } from '$lib/components/ui';
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

	let error = $state('');
	let saved = $state(false);

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
			saved = true;
			error = '';
			// The words are the sign-in pages' text, so the languages go with
			// the mail content: both are reading the same rows.
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.mail.content }),
				queryClient.invalidateQueries({ queryKey: keys.languages.all }),
				queryClient.invalidateQueries({ queryKey: keys.admin.overview })
			]);
		},
		onError: (err: unknown) => {
			error = err instanceof ApiError ? err.message : 'Could not save these words';
		}
	}));

	function submit(event: SubmitEvent) {
		event.preventDefault();

		if (save.isPending || !dirty) return;

		error = '';
		saved = false;
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
	<p class="lead">
		What each email says. An email goes out in the reader's language, so these are that language's
		words — the same rows the Languages page holds, shown here as the messages they make. A field
		left empty falls back to the English this release ships, which is what the placeholder shows.
	</p>

	{#if error}
		<Alert>{error}</Alert>
	{:else if saved && !dirty}
		<Alert tone="success">Saved. The next message in this language uses these words.</Alert>
	{/if}

	<Panel title="Language" icon={RiTranslate2}>
		{#snippet meta()}
			{#if language && !language.offered}
				<Tag tone="warning" small>Not offered</Tag>
			{/if}
		{/snippet}

		<Select
			label="Writing in"
			bind:value={code}
			options={choices}
			hint="Add and remove languages on the Languages page."
		/>
	</Panel>

	{#each content.messages as spec (spec.kind)}
		<Panel title={spec.label} icon={RiFileTextLine}>
			{#snippet meta()}
				{#if written(spec)}
					<Tag tone="info" small>Written</Tag>
				{:else}
					<Tag small>Shipped text</Tag>
				{/if}
			{/snippet}

			<p class="note">{spec.description}</p>

			<div class="fields">
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
			</div>

			<div class="foot">
				<p class="params">
					{#each spec.params as param (param.name)}
						<code title={param.description}>{'{' + param.name + '}'}</code>
					{/each}
					<span>are filled in when the message is sent.</span>
				</p>

				{#if written(spec)}
					<Button size="sm" variant="subtle" onclick={() => revert(spec)}>
						<RiRestartLine size="15" />
						Use the shipped text
					</Button>
				{/if}
			</div>
		</Panel>
	{/each}

	{#if dirty}
		<div class="actions">
			<span class="pending">Unsaved changes</span>

			<Button
				variant="subtle"
				onclick={() => (drafts = wordsOf(language))}
				disabled={save.isPending}
			>
				Discard
			</Button>
			<Button type="submit" loading={save.isPending} disabled={save.isPending}>
				{save.isPending ? 'Saving…' : 'Save changes'}
			</Button>
		</div>
	{/if}
</form>

<style>
	form {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
	}

	.lead,
	.note {
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		line-height: 1.5;
	}

	.note {
		margin-bottom: var(--space-3);
	}

	.fields {
		display: flex;
		flex-direction: column;
		gap: var(--space-3);
	}

	.foot {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-2);
		margin-top: var(--space-3);
	}

	/* The placeholders a message may use, each readable on its own: a reader
	   should be able to copy one out of the line. */
	.params {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: var(--space-2);
		margin: 0;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.params code {
		padding: 0.1rem 0.35rem;
		border-radius: var(--radius-sm);
		background: var(--color-surface-alt);
		font-size: var(--text-xs);
	}

	.actions {
		position: sticky;
		bottom: 0;
		z-index: 1;
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: var(--space-2);
		margin-inline: calc(var(--space-4) * -1);
		padding: var(--space-3) var(--space-4);
		border-top: 1px solid var(--color-secondary-alt);
		background: var(--color-surface);
		box-shadow: var(--shadow-panel);
	}

	.pending {
		margin-right: auto;
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}
</style>
