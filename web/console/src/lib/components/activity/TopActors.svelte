<script lang="ts">
	import type { Overview } from '$lib/api';
	import { List, ListItem, Tag, Thumb } from '$lib/components/ui';
	import { initials } from '$lib/utils/format';

	type Props = {
		actors: Overview['top_actors'];
	};

	let { actors }: Props = $props();

	const most = $derived(Math.max(1, ...actors.map((actor) => actor.events)));

	/** Two letters for the thumb, from the address before the @, read as
	    words where dots, dashes or underscores part it. */
	function actorInitials(actor: string): string {
		return initials((actor.split('@')[0] ?? actor).replace(/[._-]+/g, ' '));
	}
</script>

{#if actors.length === 0}
	<p class="empty">No changes made this week.</p>
{:else}
	<List label="Most active administrators">
		{#each actors as actor (actor.actor)}
			<ListItem>
				{#snippet lead()}<Thumb text={actorInitials(actor.actor)} />{/snippet}

				<span class="name" title={actor.actor}>{actor.actor}</span>
				<span class="track">
					<span class="fill" style:width="{(actor.events / most) * 100}%"></span>
				</span>

				{#snippet end()}<Tag strong>{actor.events.toLocaleString()}</Tag>{/snippet}
			</ListItem>
		{/each}
	</List>
{/if}

<style>
	.empty {
		margin: 0;
		padding: var(--space-4);
		color: var(--color-text-hint);
	}

	.name {
		overflow: hidden;
		font-size: var(--text-sm);
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.track {
		display: block;
		height: 4px;
		margin-top: 5px;
		overflow: hidden;
		border-radius: var(--radius-sm);
		background: var(--color-secondary);
	}

	.fill {
		display: block;
		height: 100%;
		background: var(--color-border);
	}
</style>
