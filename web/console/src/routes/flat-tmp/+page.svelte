<script lang="ts">
	import { RiGlobalLine, RiSmartphoneLine } from 'svelte-remixicon';
	import { Drawer, Input, Panel, Select } from '$lib/components/ui';
	let open = $state(true);
	let name = $state('Marketplace API');
	let kind = $state<'' | 'web' | 'native'>('web');
</script>

<div style="min-height: 100vh; padding: 24px; background: var(--color-body); color: var(--color-text); max-width: 700px">
	<Panel title="Applications"><p style="margin: 0">The page behind.</p></Panel>
</div>

<Drawer bind:open title="Edit application">
	<div style="display: flex; flex-direction: column; gap: 14px">
		<Input label="Name" bind:value={name} />
		<Select label="Type" bind:value={kind} options={[{ value: 'web', label: 'Web application', icon: RiGlobalLine, description: 'A server that keeps a secret.' }, { value: 'native', label: 'Native', icon: RiSmartphoneLine }]} />
	</div>
</Drawer>
