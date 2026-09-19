<script lang="ts">
	import { RiGlobalLine, RiHashtag, RiToggleLine, RiTranslate2 } from 'svelte-remixicon';
	import type { Language, LocaleApp } from '$lib/api';
	import { Badge, type Column, DataTable, Tag, Tooltip } from '$lib/components/ui';
	import { useTranslator } from '$lib/i18n';

	type Props = {
		languages: Language[];
		/** The apps a language is counted against, named by the server. */
		apps: LocaleApp[];
		/** Called with the language whose row was chosen. */
		onOpen: (language: Language) => void;
		empty: string;
	};

	let { languages, apps, onOpen, empty }: Props = $props();

	const t = useTranslator();

	/** The DataTable identifies rows by `id`; a language's code is what
	    identifies it. */
	type Row = Language & { id: string };

	const rows = $derived<Row[]>(languages.map((language) => ({ ...language, id: language.code })));

	const columns = $derived<Column[]>([
		{ key: 'language', label: t('languages.column_language'), icon: RiGlobalLine, min: '14rem' },
		{ key: 'code', label: t('languages.column_code'), icon: RiHashtag, min: '7rem' },
		{
			key: 'translated',
			label: t('languages.column_translated'),
			icon: RiTranslate2,
			min: '18rem'
		},
		{ key: 'status', label: t('languages.column_status'), icon: RiToggleLine, min: '11rem' }
	]);

	/** What each app is called in the coverage bars. */
	function appName(app: LocaleApp): string {
		return app === 'id' ? t('languages.sign_in_pages') : t('languages.admin_panel');
	}
</script>

<DataTable {columns} {rows} {empty} onOpen={(row) => onOpen(row)} label={(row) => row.name}>
	{#snippet row(language)}
		<td>
			<span class="names">
				<strong lang={language.code}>{language.native}</strong>
				<span class="english">{language.name}</span>
			</span>
		</td>

		<td><span class="chip">{language.code}</span></td>

		<td>
			<span class="bars">
				{#each apps as app (app)}
					{@const percent = language.coverage[app] ?? 0}
					{@const missing = language.missing[app] ?? 0}
					{#if language.apps.includes(app)}
						<Tooltip
							label={missing > 0
								? `${appName(app)} · ${t('languages.missing', { count: missing })}`
								: `${appName(app)} · ${t('languages.complete')}`}
						>
							{#snippet children(trigger)}
								<span class="bar" {...trigger()}>
									<span class="track">
										<span class="fill" class:whole={percent === 100} style="--filled: {percent}%"
										></span>
									</span>
									<span class="percent">{percent}%</span>
								</span>
							{/snippet}
						</Tooltip>
					{:else}
						<!-- The panel is not shown in this language, so there is
						     nothing to count; the slot stays so the rows line up. -->
						<Tooltip label={`${appName(app)} · ${t('languages.panel_only')}`}>
							{#snippet children(trigger)}
								<span class="bar none" {...trigger()}>—</span>
							{/snippet}
						</Tooltip>
					{/if}
				{/each}
			</span>
		</td>

		<td>
			<span class="status">
				{#if language.is_default}
					<Tag tone="info" strong>{t('languages.default')}</Tag>
				{:else if language.enabled}
					<Badge tone="success">{t('languages.offered')}</Badge>
				{:else}
					<Badge>{t('languages.off')}</Badge>
				{/if}
			</span>
		</td>
	{/snippet}
</DataTable>

<style>
	.names {
		display: flex;
		flex-direction: column;
		min-width: 0;
		line-height: 1.3;
	}

	.english {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
	}

	.chip {
		display: inline-block;
		padding: 3px 6px;
		border-radius: var(--radius-sm);
		background: var(--color-secondary-alt);
		color: var(--color-text-hint);
		font-family: var(--font-mono);
		font-size: var(--text-sm);
	}

	/* One bar per app, side by side: a language may be finished for the
	   sign-in pages and half-done for the panel, and that is the thing
	   somebody deciding whether to offer it needs to see. */
	.bars {
		display: flex;
		align-items: center;
		gap: var(--space-3);
	}

	.bar {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
	}

	.track {
		display: block;
		width: 4.5rem;
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
	}

	.fill.whole {
		background: var(--color-success);
	}

	/* As wide as a bar and its number, so the columns stay in line. */
	.bar.none {
		justify-content: center;
		width: calc(4.5rem + 2.75rem);
		color: var(--color-text-hint);
	}

	.percent {
		color: var(--color-text-hint);
		font-size: var(--text-sm);
		font-variant-numeric: tabular-nums;
	}

	.status {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
	}
</style>
