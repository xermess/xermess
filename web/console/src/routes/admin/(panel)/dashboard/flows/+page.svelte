<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import { RiAddLine, RiCloseLine, RiRefreshLine, RiUpload2Line } from 'svelte-remixicon';
	import { createQuery, useQueryClient } from '@tanstack/svelte-query';
	import { flowsApi, messageOf, type LoginFlow } from '$lib/api';
	import { BRAND } from '$lib/brand';
	import { Alert, Button, Icon, IconButton, PageHeader, SearchInput } from '$lib/components/ui';
	import FlowTable from '$lib/components/flows/FlowTable.svelte';
	import { TEMPLATES, freeSlug, fromFile } from '$lib/components/flows/steps';
	import { can } from '$lib/permissions';
	import { keys, loginFlowsOptions } from '$lib/query';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const queryClient = useQueryClient();

	// The list is a query seeded with what the server rendered: a save in the
	// editor or the refresh button refills the cache rather than reloading.
	const flows = createQuery(() =>
		loginFlowsOptions({ flows: data.flows, step_kinds: data.stepKinds })
	);

	const canWrite = $derived(can(data.admin, 'login_flows.write'));

	// Typing updates it, and it follows the URL again whenever that changes,
	// so the back button and a shared link both put the right term in the box.
	let search = $derived(data.search);

	/** The rows the search leaves. There are as many flows as somebody has
	    written — a handful — so narrowing them here is cheaper than asking
	    again, and a shared link arrives already filtered. */
	const visible = $derived.by(() => {
		const term = data.search.trim().toLowerCase();
		if (term === '') return flows.data.flows;

		return flows.data.flows.filter((flow) =>
			[flow.name, flow.slug, flow.description].some((value) => value.toLowerCase().includes(term))
		);
	});

	let choosing = $state(false);
	let error = $state('');
	let importing = $state(false);
	let file = $state<HTMLInputElement>();

	async function apply(changes: { search?: string }) {
		const params = new SvelteURLSearchParams(page.url.searchParams);

		for (const [key, value] of Object.entries(changes)) {
			if (value) params.set(key, value);
			else params.delete(key);
		}

		const query = params.toString();
		const path = resolve('/admin/(panel)/dashboard/flows');

		// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto(query ? `${path}?${query}` : path, { keepFocus: true, noScroll: true });
	}

	let timer: ReturnType<typeof setTimeout>;

	function debounced() {
		clearTimeout(timer);
		timer = setTimeout(() => apply({ search }), 250);
	}

	function open(flow: LoginFlow) {
		goto(resolve('/admin/(panel)/dashboard/flows/[id]', { id: flow.id }));
	}

	/** A flow exported from here or another installation, made a flow of this
	    one under an identifier nobody has, and opened. */
	async function importFile(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const chosen = input.files?.[0];
		input.value = '';
		if (!chosen) return;

		error = '';
		importing = true;

		try {
			let draft;
			try {
				draft = fromFile(await chosen.text(), flows.data.step_kinds);
			} catch (err) {
				error = (err as Error).message;
				return;
			}

			const taken = flows.data.flows.map((flow) => flow.slug);
			const result = await flowsApi.create({
				...draft,
				slug: freeSlug(draft.slug, taken),
				is_default: false
			});
			await queryClient.invalidateQueries({ queryKey: keys.flows.all });
			await goto(resolve('/admin/(panel)/dashboard/flows/[id]', { id: result.flow.id }));
		} catch (err) {
			error = messageOf(err);
		} finally {
			importing = false;
		}
	}

	let refreshing = $state(false);

	async function refresh() {
		refreshing = true;

		try {
			await Promise.all([
				queryClient.invalidateQueries({ queryKey: keys.flows.all }),
				new Promise((done) => setTimeout(done, 400))
			]);
		} finally {
			refreshing = false;
		}
	}
</script>

<svelte:head><title>Login flows · {BRAND.name}</title></svelte:head>

<div class="heading">
	<PageHeader crumbs={['Dashboard', 'Login flows']}>
		{#snippet secondary()}
			<span class="total">{`${flows.data.flows.length} total`}</span>

			<IconButton
				icon={RiRefreshLine}
				label="Refresh the data"
				onclick={refresh}
				loading={refreshing}
				disabled={refreshing}
			/>
		{/snippet}

		{#snippet actions()}
			{#if canWrite}
				<input
					id="flow-import-file"
					name="flow-import-file"
					bind:this={file}
					type="file"
					accept="application/json,.json"
					hidden
					onchange={importFile}
				/>
				<Button variant="subtle" loading={importing} onclick={() => file?.click()}>
					<Icon icon={RiUpload2Line} />
					Import
				</Button>
				<Button onclick={() => (choosing = !choosing)}>
					<Icon icon={choosing ? RiCloseLine : RiAddLine} />
					New flow
				</Button>
			{/if}
		{/snippet}
	</PageHeader>

	<p class="lead">
		How people sign in, drawn as the steps they go through. Every application follows its own flow,
		or the default: the password, the other accounts, a verified address and how long a session
		lasts are all decided here.
	</p>
</div>

{#if error}
	<div class="gutter banner"><Alert>{error}</Alert></div>
{/if}

{#if choosing}
	<section class="gutter templates" aria-label="Start from a template">
		<h2>Start from a template</h2>
		<div class="cards">
			{#each TEMPLATES as template (template.id)}
				<!-- The path is resolved; only the query is added to it. -->
				<!-- eslint-disable svelte/no-navigation-without-resolve -->
				<a
					class="card"
					href={`${resolve('/admin/(panel)/dashboard/flows/new')}?template=${template.id}`}
				>
					<span class="mark"><Icon icon={template.icon} size="1.2rem" /></span>
					<strong>{template.name}</strong>
					<span>{template.hint}</span>
				</a>
				<!-- eslint-enable svelte/no-navigation-without-resolve -->
			{/each}
		</div>
	</section>
{/if}

<div class="toolbar">
	<SearchInput
		label="Search login flows"
		placeholder="Search name, identifier or description…"
		bind:value={search}
		onsubmit={() => apply({ search })}
		oninput={debounced}
	/>
</div>

<FlowTable
	flows={visible}
	kinds={flows.data.step_kinds}
	onOpen={open}
	empty={flows.data.flows.length === 0 ? 'No flows yet.' : 'No flows match this.'}
/>

<style>
	.heading {
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}

	.lead {
		max-width: 90ch;
		margin: var(--space-2) 0 0;
		color: var(--color-text-hint);
		font-size: var(--text-base);
	}

	.gutter {
		padding-inline: var(--page-gutter);
	}

	.banner {
		margin-bottom: var(--space-3);
	}

	.templates {
		margin-bottom: var(--space-4);
	}

	.templates h2 {
		margin: 0 0 var(--space-2);
		font-size: var(--text-base);
	}

	.cards {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
		gap: var(--space-3);
	}

	.card {
		display: flex;
		flex-direction: column;
		gap: var(--space-1);
		padding: var(--space-4);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-md);
		background: var(--color-surface);
		color: var(--color-text);
		text-decoration: none;
		transition:
			border-color var(--speed-fast),
			background-color var(--speed-fast);
	}

	.card:hover {
		border-color: var(--color-info);
		background: var(--surface-info);
	}

	.card span:last-child {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.mark {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 36px;
		height: 36px;
		margin-bottom: var(--space-1);
		border-radius: var(--radius-sm);
		background: var(--color-secondary-alt);
	}

	.toolbar {
		display: flex;
		gap: var(--space-2);
		margin-bottom: var(--space-3);
		padding-inline: var(--page-gutter);
	}
</style>
