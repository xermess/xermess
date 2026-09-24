<script lang="ts">
	/**
	 * Everywhere the panel can go, in one list, over the page.
	 *
	 * The sidebar is the map; this is the shortcut for somebody who already
	 * knows where they are going. It is built from the same `sections`
	 * catalog, so a page added there is in here with its permission check and
	 * its translated name already done, and there is no second list to keep
	 * in step.
	 *
	 * The matching is deliberately loose — the letters of the query in order,
	 * not in a row — so "adro" finds "Admin roles". That is what every
	 * palette does, and it is what makes typing three letters enough.
	 */
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import type { ComponentType } from 'svelte';
	import { Dialog } from '@ark-ui/svelte/dialog';
	import { Portal } from '@ark-ui/svelte/portal';
	import {
		RiBookOpenLine,
		RiCornerDownLeftLine,
		RiExternalLinkLine,
		RiGithubFill,
		RiSearchLine
	} from 'svelte-remixicon';
	import type { Admin } from '$lib/api';
	import { Icon } from '$lib/components/ui';
	import { DOCS_URL, GITHUB_URL } from '$lib/constants';
	import { allSections } from './sidebar/sections';

	type Command = {
		id: string;
		label: string;
		group: string;
		icon: ComponentType;
		href: string;
		/** Off this site: it opens in a tab of its own and the row says so. */
		external?: boolean;
	};

	let open = $state(false);
	let query = $state('');
	let active = $state(0);
	/** The rows, for scrolling the active one back into view. */
	let list = $state<HTMLElement | null>(null);

	const admin = $derived(page.data.admin as Admin | undefined);

	/** Every page this administrator may open, then the two that are not
	    pages of this installation at all. */
	const commands = $derived<Command[]>([
		...allSections(admin).flatMap((section) =>
			section.items.map((item) => ({
				id: item.route,
				label: item.label,
				group: section.group,
				icon: item.icon,
				href: resolve(item.route)
			}))
		),
		{
			id: 'docs',
			label: 'Documentation',
			group: 'Help',
			icon: RiBookOpenLine,
			href: DOCS_URL,
			external: true
		},
		{
			id: 'github',
			label: 'GitHub',
			group: 'Help',
			icon: RiGithubFill,
			href: GITHUB_URL,
			external: true
		}
	]);

	const matches = $derived(commands.filter((command) => fuzzy(command, query)));

	/** The matches under their group headings, in the order the groups first
	    appear, so filtering never reshuffles the list. */
	const grouped = $derived(
		matches.reduce<{ name: string; items: Command[] }[]>((groups, command) => {
			const last = groups.at(-1);
			if (last?.name === command.group) last.items.push(command);
			else groups.push({ name: command.group, items: [command] });

			return groups;
		}, [])
	);

	/** Whether the query's letters appear in the name, in order. The group's
	    name counts too, so "settings" brings up everything filed under it. */
	function fuzzy(command: Command, text: string): boolean {
		const needle = text.trim().toLowerCase();
		if (!needle) return true;

		const haystack = `${command.label} ${command.group}`.toLowerCase();
		let at = 0;

		for (const letter of needle) {
			if (letter === ' ') continue;

			at = haystack.indexOf(letter, at);
			if (at === -1) return false;
			at += 1;
		}

		return true;
	}

	function show() {
		query = '';
		active = 0;
		open = true;
	}

	/** ⌘K, or Ctrl+K where there is no ⌘. The one shortcut every palette has,
	    so it is worth taking from the page wherever the reader is. */
	function onKeydown(event: KeyboardEvent) {
		if (event.key.toLowerCase() !== 'k' || !(event.metaKey || event.ctrlKey)) return;

		event.preventDefault();

		if (open) open = false;
		else show();
	}

	/** Arrows move the highlight, Enter follows it. The input keeps the
	    focus throughout, so typing and choosing are the same gesture. */
	function onInputKeydown(event: KeyboardEvent) {
		if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
			event.preventDefault();

			const step = event.key === 'ArrowDown' ? 1 : -1;
			const count = matches.length;
			if (count === 0) return;

			// Wrapping means Up from the top reaches the last row, which is
			// quicker than holding Down through the whole list.
			active = (active + step + count) % count;
			scrollIntoView();

			return;
		}

		if (event.key === 'Enter') {
			event.preventDefault();
			run(matches[active]);
		}
	}

	function scrollIntoView() {
		// After the highlight has been drawn, not before.
		requestAnimationFrame(() => {
			list?.querySelector('[data-active="true"]')?.scrollIntoView({ block: 'nearest' });
		});
	}

	async function run(command: Command | undefined) {
		if (!command) return;

		open = false;

		if (command.external) {
			window.open(command.href, '_blank', 'noopener,noreferrer');
			return;
		}

		// Already resolved, back where the list was built from the routes
		// themselves — by here it is an address, not a route to look up.
		// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto(command.href);
	}
</script>

<svelte:window onkeydown={onKeydown} />

<!-- The trigger reads as the field it opens rather than as a button: that is
     what says "type here" without a word of instruction, and the shortcut
     beside it teaches the faster way. Narrow screens get the magnifier
     alone. -->
<button type="button" class="trigger" onclick={show}>
	<Icon icon={RiSearchLine} size="1rem" />
	<span class="placeholder">Search…</span>
	<kbd>⌘K</kbd>
</button>

<Dialog.Root bind:open lazyMount unmountOnExit>
	<Portal>
		<Dialog.Backdrop class="palette-backdrop" />
		<Dialog.Positioner class="palette-positioner">
			<Dialog.Content class="palette-content" aria-label="Search…">
				<div class="field">
					<Icon icon={RiSearchLine} size="1.125rem" />
					<!-- svelte-ignore a11y_autofocus -->
					<input
						autofocus
						type="text"
						autocomplete="off"
						spellcheck="false"
						placeholder="Go to a page…"
						bind:value={query}
						oninput={() => (active = 0)}
						onkeydown={onInputKeydown}
					/>
				</div>

				<div class="results" bind:this={list}>
					{#each grouped as group (group.name)}
						<div class="group">
							<h2>{group.name}</h2>

							{#each group.items as command (command.id)}
								{@const index = matches.indexOf(command)}
								<button
									type="button"
									class="row"
									data-active={index === active}
									onclick={() => run(command)}
									onmousemove={() => (active = index)}
								>
									<Icon icon={command.icon} />
									<span class="label">{command.label}</span>

									{#if command.external}
										<Icon icon={RiExternalLinkLine} size="0.875rem" />
									{:else if index === active}
										<Icon icon={RiCornerDownLeftLine} size="0.875rem" />
									{/if}
								</button>
							{/each}
						</div>
					{:else}
						<p class="empty">{`Nothing matches “${query}”`}</p>
					{/each}
				</div>

				<footer>
					<span><kbd>↑</kbd><kbd>↓</kbd> move</span>
					<span><kbd>↵</kbd> open</span>
					<span><kbd>esc</kbd> close</span>
				</footer>
			</Dialog.Content>
		</Dialog.Positioner>
	</Portal>
</Dialog.Root>

<style>
	/* The trigger: an input's colours at a control's height, which is the
	   shape everything else in the bar already has. */
	.trigger {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		width: 18rem;
		max-width: 100%;
		height: var(--control-height-sm);
		padding: 0 var(--space-2);
		border: 1px solid transparent;
		border-radius: var(--radius-sm);
		background: var(--color-input);
		color: var(--color-text-hint);
		font: inherit;
		font-size: var(--text-base);
		text-align: left;
		cursor: pointer;
		transition:
			background-color var(--speed-fast),
			border-color var(--speed-fast);
	}

	.trigger:hover {
		background: var(--color-input-focus);
	}

	.trigger:focus-visible {
		outline: none;
		border-color: var(--color-accent);
	}

	.placeholder {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	kbd {
		padding: 1px 5px;
		border: 1px solid var(--color-border);
		border-radius: 4px;
		background: var(--color-surface);
		color: var(--color-text-hint);
		font-family: var(--font-sans);
		font-size: var(--text-xs);
		line-height: 1.4;
	}

	/* The dialog's anatomy is styled once in styles/ark.css, and it is styled
	   for the drawer: a full-height panel against the right edge. A palette
	   is the other kind of dialog, so these three override it by being one
	   selector more specific. */
	:global(.palette-positioner[data-part='positioner']) {
		display: flex;
		justify-content: center;
		align-items: flex-start;
		padding: var(--space-4);
		/* Below the top of the window rather than centred: the list grows
		   downwards, so a centred palette would move while being typed in. */
		padding-top: 12vh;
		overflow-y: auto;
	}

	:global(.palette-backdrop[data-part='backdrop']) {
		backdrop-filter: blur(2px);
	}

	:global(.palette-content[data-part='content']) {
		width: min(36rem, 100%);
		height: auto;
		max-height: min(26rem, 70vh);
		border: 1px solid var(--color-border);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-md);
		overflow: hidden;
		animation: palette-in var(--speed) ease;
	}

	:global(.palette-content[data-part='content'][data-state='closed']) {
		animation: palette-out var(--speed) ease forwards;
	}

	/* It arrives a shade small and settles, which reads as the palette coming
	   forward rather than a box appearing. Two names, because Ark waits for a
	   leaving animation only when the name changes — the note in ark.css
	   beside the drawer's keyframes explains why. */
	@keyframes palette-in {
		from {
			opacity: 0;
			transform: scale(0.98) translateY(-4px);
		}
	}

	@keyframes palette-out {
		to {
			opacity: 0;
			transform: scale(0.98) translateY(-4px);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		:global(.palette-content[data-part='content']),
		:global(.palette-content[data-part='content'][data-state='closed']) {
			animation: none;
		}
	}

	.field {
		display: flex;
		flex-shrink: 0;
		align-items: center;
		gap: var(--space-2);
		padding: 0 var(--space-3);
		height: var(--control-height-lg);
		border-bottom: 1px solid var(--color-border);
		color: var(--color-text-hint);
	}

	/* The field is the whole strip, so the input itself is just the text in
	   it: no box of its own inside another box. */
	.field input {
		flex: 1;
		min-width: 0;
		border: none;
		background: none;
		color: var(--color-text);
		font: inherit;
		font-size: var(--text-lg);
	}

	.field input:focus {
		outline: none;
	}

	.results {
		flex: 1;
		min-height: 0;
		padding: var(--space-1);
		overflow-y: auto;
	}

	.group h2 {
		margin: 0;
		padding: var(--space-2) var(--space-2) var(--space-1);
		color: var(--color-text-hint);
		font-size: var(--text-xs);
		font-weight: 600;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}

	.row {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		width: 100%;
		padding: var(--space-2);
		border: none;
		border-radius: var(--radius-sm);
		background: none;
		color: var(--color-text);
		font: inherit;
		font-size: var(--text-base);
		text-align: left;
		cursor: pointer;
	}

	/* One row is highlighted at a time and the keyboard owns which: the mouse
	   moves the highlight rather than drawing a second one of its own, so
	   there is never a question which row Enter will follow. */
	.row[data-active='true'] {
		background: var(--color-secondary);
	}

	.row .label {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.row :global(svg:last-child) {
		color: var(--color-text-hint);
	}

	.empty {
		margin: 0;
		padding: var(--space-5) var(--space-2);
		color: var(--color-text-hint);
		font-size: var(--text-base);
		text-align: center;
	}

	footer {
		display: flex;
		flex-shrink: 0;
		gap: var(--space-3);
		padding: var(--space-2) var(--space-3);
		border-top: 1px solid var(--color-border);
		background: var(--color-surface-alt);
		color: var(--color-text-hint);
		font-size: var(--text-xs);
	}

	footer span {
		display: flex;
		align-items: center;
		gap: var(--space-1);
	}

	footer kbd {
		background: var(--color-surface);
	}

	/* In a narrow bar the trigger is the magnifier and nothing else — the
	   shortcut it teaches needs a keyboard anyway. */
	@media (max-width: 55rem) {
		.trigger {
			width: var(--control-height-sm);
			justify-content: center;
			padding: 0;
		}

		.placeholder,
		.trigger kbd {
			display: none;
		}

		footer {
			display: none;
		}
	}
</style>
