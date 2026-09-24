<script lang="ts">
	import { beforeNavigate, goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { SvelteFlowProvider } from '@xyflow/svelte';
	import {
		RiCheckLine,
		RiDeleteBinLine,
		RiDownload2Line,
		RiFileCopyLine,
		RiArrowGoBackLine
	} from 'svelte-remixicon';
	import { useQueryClient } from '@tanstack/svelte-query';
	import {
		flowsApi,
		messageOf,
		type LoginFlow,
		type LoginStep,
		type LoginStepSpec
	} from '$lib/api';
	import { Alert, Button, Icon, IconButton, PageHeader, Tag } from '$lib/components/ui';
	import { keys } from '$lib/query';
	import { draftOf, download, freeSlug, problemsOf, toFile, type FlowDraft } from '../steps';
	import FlowCanvas from './FlowCanvas.svelte';
	import Inspector from './Inspector.svelte';
	import Palette from './Palette.svelte';

	/**
	 * A login flow's editor: the steps a sign-in goes through on a canvas, the
	 * steps that can be added beside it, and the settings of whatever is
	 * selected on the other side. Nothing is saved until Save — the draft is
	 * the page's own, checked against the server's rules as it is drawn, and
	 * leaving with it unsaved asks first.
	 */
	type Props = {
		/** The flow being edited, or null for a new one. */
		flow: LoginFlow | null;
		/** Where the draft starts: the flow, or a template or an import. */
		initial: FlowDraft;
		kinds: LoginStepSpec[];
		/** Every flow there is, for choosing an identifier nobody has. */
		taken: string[];
		editable: boolean;
	};

	let { flow, initial, kinds, taken, editable }: Props = $props();

	const queryClient = useQueryClient();

	// svelte-ignore state_referenced_locally
	let draft = $state(draftOf(initial));
	// svelte-ignore state_referenced_locally
	let saved = $state(JSON.stringify(draftOf(initial)));
	const creating = $derived(flow === null);
	const dirty = $derived(creating || JSON.stringify(draft) !== saved);
	const problems = $derived(problemsOf(draft));

	let selected = $state('start');
	/** The gap a "+" on the canvas is waiting to fill from the palette. */
	let waiting = $state<number | null>(null);

	let error = $state('');
	let busy = $state(false);
	let confirmingDelete = $state(false);
	/** Set on the way out after a save or a delete, which leave nothing to lose. */
	let leaving = false;

	beforeNavigate(({ cancel }) => {
		if (
			!leaving &&
			editable &&
			dirty &&
			!creating &&
			!confirm('This flow has changes that are not saved. Leave anyway?')
		)
			cancel();
	});

	// ---- Changing the steps ----------------------------------------------------

	function insert(index: number, step: LoginStep) {
		if (draft.steps.includes(step)) return;

		const at = Math.max(1, Math.min(index, draft.steps.length));
		draft.steps = [...draft.steps.slice(0, at), step, ...draft.steps.slice(at)];
		waiting = null;
		selected = step;
	}

	function add(step: LoginStep) {
		insert(waiting ?? draft.steps.length, step);
	}

	function remove(step: LoginStep) {
		draft.steps = draft.steps.filter((one) => one !== step);
		if (selected === step) selected = 'start';
		waiting = null;
	}

	// ---- What the canvas says ----------------------------------------------------

	function lifetimeText(hours: number): string {
		return hours % 24 === 0 && hours >= 24 ? `${hours / 24} d` : `${hours} h`;
	}

	function summaryOf(step: LoginStep): string[] {
		switch (step) {
			case 'identifier':
				return [
					draft.allow_registration ? 'sign-ups' : 'no sign-ups',
					...(draft.require_verified_email ? ['verified address'] : [])
				];
			case 'password':
				return [draft.allow_password_reset ? 'reset link' : 'no reset link'];
			default:
				return [];
		}
	}

	// ---- Saving and the rest -----------------------------------------------------

	async function run(action: () => Promise<void>) {
		error = '';
		busy = true;
		try {
			await action();
		} catch (err) {
			error = messageOf(err);
		} finally {
			busy = false;
		}
	}

	function body(of: FlowDraft) {
		// The identifier is in every address the flow is known by, so it is
		// only sent when the flow is made.
		const { slug, ...rest } = draftOf(of);
		return creating ? { ...rest, slug } : rest;
	}

	const save = () =>
		run(async () => {
			const result = creating
				? await flowsApi.create(body(draft))
				: await flowsApi.update(flow!.id, body(draft));

			await queryClient.invalidateQueries({ queryKey: keys.flows.all });
			saved = JSON.stringify(draftOf(draft));

			if (creating) {
				leaving = true;
				await goto(resolve('/admin/(panel)/dashboard/flows/[id]', { id: result.flow.id }));
			}
		});

	const duplicate = () =>
		run(async () => {
			const copy: FlowDraft = {
				...draftOf(draft),
				name: `Copy of ${draft.name}`,
				slug: freeSlug(`${draft.slug}-copy`, taken),
				is_default: false
			};
			const result = await flowsApi.create(copy);
			await queryClient.invalidateQueries({ queryKey: keys.flows.all });

			leaving = true;
			await goto(resolve('/admin/(panel)/dashboard/flows/[id]', { id: result.flow.id }));
		});

	const remove_ = () =>
		run(async () => {
			await flowsApi.remove(flow!.id);
			await queryClient.invalidateQueries({ queryKey: keys.flows.all });

			leaving = true;
			await goto(resolve('/admin/(panel)/dashboard/flows'));
		});

	function discard() {
		draft = JSON.parse(saved);
		selected = 'start';
		waiting = null;
	}
</script>

<SvelteFlowProvider>
	<div class="heading">
		<PageHeader
			crumbs={[
				'Dashboard',
				{ label: 'Login flows', href: resolve('/admin/(panel)/dashboard/flows') },
				draft.name || 'Untitled flow'
			]}
		>
			{#snippet secondary()}
				{#if flow?.is_default}<Tag tone="info" strong>default</Tag>{/if}
				{#if !draft.enabled}<Tag>off</Tag>{/if}
				{#if editable && dirty && !creating}<Tag tone="warning">unsaved</Tag>{/if}

				<IconButton
					icon={RiDownload2Line}
					label="Export as JSON"
					onclick={() =>
						download(`${draft.slug || 'flow'}.flow.json`, JSON.stringify(toFile(draft), null, 2))}
				/>
				{#if editable && !creating}
					<IconButton icon={RiFileCopyLine} label="Duplicate" onclick={duplicate} disabled={busy} />
				{/if}
			{/snippet}

			{#snippet actions()}
				{#if editable}
					{#if !creating && !flow?.is_default}
						{#if confirmingDelete}
							<span class="confirm"
								>{`Delete this flow? ${flow?.applications ?? 0} applications fall back to the default.`}</span
							>
							<Button variant="subtle" size="sm" onclick={() => (confirmingDelete = false)}>
								Keep
							</Button>
							<Button colorPalette="danger" size="sm" loading={busy} onclick={remove_}>
								Delete
							</Button>
						{:else}
							<Button
								variant="subtle"
								colorPalette="danger"
								onclick={() => (confirmingDelete = true)}
							>
								<Icon icon={RiDeleteBinLine} />
								Delete
							</Button>
						{/if}
					{/if}
					{#if dirty && !creating}
						<Button variant="subtle" onclick={discard} disabled={busy}>
							<Icon icon={RiArrowGoBackLine} />
							Discard changes
						</Button>
					{/if}
					<Button onclick={save} loading={busy} disabled={!dirty || problems.length > 0}>
						<Icon icon={RiCheckLine} />
						{creating ? 'Create flow' : 'Save changes'}
					</Button>
				{/if}
			{/snippet}
		</PageHeader>
	</div>

	{#if error || problems.length > 0 || !editable}
		<div class="messages">
			{#if error}<Alert>{error}</Alert>{/if}
			{#if problems.length > 0}
				<Alert tone="warning">
					<span class="problems">
						{#each problems as problem (problem)}
							<span>{problem}</span>
						{/each}
					</span>
				</Alert>
			{/if}
			{#if !editable}<Alert tone="info"
					>Your roles let you look at login flows, not change them.</Alert
				>{/if}
		</div>
	{/if}

	<div class="workspace">
		<Palette {kinds} steps={draft.steps} {editable} waiting={waiting !== null} onAdd={add} />

		<FlowCanvas
			steps={draft.steps}
			{kinds}
			{summaryOf}
			startDetail={draft.name || 'Untitled flow'}
			endDetail={`session: ${lifetimeText(draft.session_lifetime_hours)}`}
			{selected}
			{editable}
			{waiting}
			onSelect={(id) => (selected = id)}
			onReorder={(steps) => (draft.steps = steps)}
			onInsert={insert}
			onRemove={remove}
			onWait={(index) => (waiting = index)}
		/>

		<Inspector
			bind:draft
			{kinds}
			{selected}
			{editable}
			{creating}
			savedDefault={flow?.is_default ?? false}
			applications={flow?.applications ?? 0}
			onRemove={remove}
		/>
	</div>
</SvelteFlowProvider>

<style>
	.heading {
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}

	.messages {
		display: flex;
		flex-direction: column;
		gap: var(--space-2);
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}

	.problems {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}

	.confirm {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	/* The editor takes the rest of the screen: the canvas is the page, and a
	   canvas that scrolls with the page is one that fights every drag. */
	.workspace {
		display: grid;
		grid-template-columns: 260px minmax(0, 1fr) 320px;
		height: calc(100dvh - var(--header-height) - 120px);
		min-height: 560px;
		border-top: 1px solid var(--color-border);
		border-bottom: 1px solid var(--color-border);
	}

	@media (max-width: 64rem) {
		.workspace {
			grid-template-columns: minmax(0, 1fr);
			grid-template-rows: auto 520px auto;
			height: auto;
		}

		.workspace :global(.palette),
		.workspace :global(.inspector) {
			border: none;
			border-bottom: 1px solid var(--color-border);
		}
	}
</style>
