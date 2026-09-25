<script lang="ts">
	import { resolve } from '$app/paths';
	import { Menu } from '@ark-ui/svelte/menu';
	import { Portal } from '@ark-ui/svelte/portal';
	import { RiArrowRightSLine } from 'svelte-remixicon';
	import { Icon } from '$lib/components/ui';
	import SidebarLink from './SidebarLink.svelte';
	import type { Section, SidebarBranch } from './sections';

	type Props = {
		branch: SidebarBranch;
		/** Showing its pages. Ignored while the column is folded, which has no
		    room for them. */
		open: boolean;
		onToggle: () => void;
		/** Which page is being read, so the branch holding it can say so. */
		isCurrent: (route: Section) => boolean;
		/** The column is folded to its icons. */
		collapsed: boolean;
	};

	let { branch, open, onToggle, isCurrent, collapsed }: Props = $props();

	/** Whether the page being read is one of this branch's. A folded branch
	    says so with a dot, so the column still answers "where am I" without
	    being unfolded. */
	const holdsCurrent = $derived(branch.items.some((item) => isCurrent(item.route)));
</script>

{#if collapsed}
	<!-- Folded, there is no room to open underneath: the pages come out
	     beside the icon instead, as the menu they already look like. -->
	<!-- No tooltip on the trigger: the flyout names the section in its own
	     heading, and a tooltip and a menu on one button fight over the same
	     handlers. -->
	<Menu.Root positioning={{ placement: 'right-start', gutter: 8 }}>
		<Menu.Trigger>
			{#snippet asChild(menu)}
				<button {...menu()} type="button" class="row folded" aria-label={branch.label}>
					<span class="icon"><Icon icon={branch.icon} /></span>
					<!-- Named on a narrow screen, where the column is a row and
					     there is width for words but no height for a tree. -->
					<span class="name">{branch.label}</span>
					{#if holdsCurrent}<span class="here" aria-hidden="true"></span>{/if}
				</button>
			{/snippet}
		</Menu.Trigger>

		<Portal>
			<Menu.Positioner>
				<Menu.Content class="branch-menu">
					<!-- The label has to be inside the group it names: Ark reads
					     the group's context to tie the two together. -->
					<Menu.ItemGroup>
						<Menu.ItemGroupLabel>{branch.label}</Menu.ItemGroupLabel>

						{#each branch.items as item (item.route)}
							<Menu.Item value={item.route}>
								{#snippet asChild(menuItem)}
									<a
										{...menuItem()}
										href={resolve(item.route)}
										aria-current={isCurrent(item.route) ? 'page' : undefined}
										class:current={isCurrent(item.route)}
									>
										<Icon icon={item.icon} />
										{item.label}
									</a>
								{/snippet}
							</Menu.Item>
						{/each}
					</Menu.ItemGroup>
				</Menu.Content>
			</Menu.Positioner>
		</Portal>
	</Menu.Root>
{:else}
	<button
		type="button"
		class="row"
		aria-expanded={open}
		aria-controls="branch-{branch.id}"
		onclick={onToggle}
	>
		<span class="icon"><Icon icon={branch.icon} /></span>
		<span class="label">{branch.label}</span>
		<span class="chevron" class:open aria-hidden="true">
			<Icon icon={RiArrowRightSLine} size="1rem" />
		</span>
	</button>

	<!-- Kept in the markup while closed rather than removed, so opening one is
	     a smooth height change rather than a mount. The wrapper is inert and
	     hidden from assistive technology until it is open. -->
	<div id="branch-{branch.id}" class="pages" class:open aria-hidden={!open} inert={!open}>
		<div class="pages-inner">
			<ul>
				{#each branch.items as item (item.route)}
					<li>
						<SidebarLink
							route={item.route}
							label={item.label}
							icon={item.icon}
							current={isCurrent(item.route)}
							collapsed={false}
							nested
						/>
					</li>
				{/each}
			</ul>
		</div>
	</div>
{/if}

<style>
	/* The parent row is laid out exactly like a link, so the column reads as
	   one list whether a row leads somewhere or opens. */
	.row {
		position: relative;
		display: flex;
		align-items: center;
		gap: var(--space-2);
		width: 100%;
		height: var(--nav-item-height);
		padding: 0 8px 0 10px;
		overflow: hidden;
		border: none;
		border-radius: var(--radius-sm);
		background: transparent;
		color: var(--nav-text);
		font: inherit;
		font-size: var(--text-base);
		white-space: nowrap;
		cursor: pointer;
		transition:
			background-color var(--speed-fast),
			color var(--speed-fast);
	}

	/* A branch names a subject rather than a page, so it is set a step
	   smaller and heavier than the rows under it: the column reads as
	   headings with their pages, not as one long list. */
	.row:not(.folded) {
		font-size: var(--text-sm);
		font-weight: 600;
	}

	.row:hover {
		background: var(--nav-hover);
		color: var(--color-text);
	}

	.row:focus-visible {
		outline: 2px solid var(--color-accent);
		outline-offset: -2px;
	}

	.icon {
		display: inline-flex;
		flex: none;
		width: 16px;
		height: 16px;
		align-items: center;
		justify-content: center;
		color: inherit;
	}

	.label {
		flex: 1 1 0;
		min-width: 0;
		overflow: hidden;
		text-align: left;
		text-overflow: ellipsis;
	}

	/* The chevron points along the row when the branch is closed and down
	   when it is open, which is the one mark saying a row opens rather than
	   leads. It is quiet until the row is under the pointer. */
	.chevron {
		display: inline-flex;
		flex: none;
		color: var(--color-text-hint);
		transition:
			transform var(--speed),
			color var(--speed-fast);
	}

	.row:hover .chevron {
		color: var(--color-text);
	}

	.chevron.open {
		transform: rotate(90deg);
	}

	/* The pages sit under their branch with a rule down the left: the one
	   line that makes the column a tree rather than a list with gaps. The
	   branch opens as a height, not a jump, and without a fade: the rows are
	   solid the whole way. The inner element is the grid item that can
	   actually shrink; the outer one owns the transition and the tree's left
	   rule. */
	.pages {
		display: grid;
		grid-template-rows: 0fr;
		margin: 0 0 0 17px;
		visibility: hidden;
		transition:
			grid-template-rows var(--speed) ease,
			margin var(--speed) ease,
			visibility 0s linear var(--speed);
	}

	.pages.open {
		grid-template-rows: 1fr;
		margin-top: 1px;
		visibility: visible;
		transition:
			grid-template-rows var(--speed) ease,
			margin var(--speed) ease;
	}

	.pages-inner {
		min-height: 0;
		overflow: hidden;
		padding-left: var(--nav-branch-inset);
		border-left: 1px solid var(--color-border);
	}

	.pages ul {
		display: flex;
		flex-direction: column;
		gap: 1px;
		margin: 0;
		padding: 0;
		list-style: none;
	}

	/* Folded, the row is the icon alone. */
	.row.folded {
		justify-content: flex-start;
		padding-right: 10px;
	}

	.row.folded .name {
		display: none;
	}

	/* Where you are, while the names are away: a dot on the branch holding
	   the page being read. */
	.here {
		position: absolute;
		top: 50%;
		right: 7px;
		width: 5px;
		height: 5px;
		border-radius: var(--radius-pill);
		background: var(--nav-mark);
		transform: translateY(-50%);
	}

	/* The flyout is an ordinary menu; these are the two things it needs that
	   a menu of actions does not — rows that are links, and a mark on the one
	   being read. */
	:global(.branch-menu) {
		min-width: 13rem;
	}

	:global(.branch-menu a[data-part='item']) {
		color: var(--color-text);
		text-decoration: none;
	}

	:global(.branch-menu a[data-part='item'].current) {
		background: var(--color-brand);
		color: var(--color-brand-text);
		font-weight: 600;
	}

	@media (prefers-reduced-motion: reduce) {
		.row,
		.pages,
		.pages.open,
		.chevron {
			transition: none;
		}
	}
</style>
