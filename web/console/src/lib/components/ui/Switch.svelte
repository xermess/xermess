<script lang="ts">
	import { Switch } from '@ark-ui/svelte/switch';
	import { onlyTheSwitch } from './switch';

	type Props = {
		label: string;
		checked: boolean;
		onChange?: (checked: boolean) => void;
		disabled?: boolean;
	};

	let { label, checked = $bindable(false), onChange, disabled }: Props = $props();
</script>

<!-- A boolean is the toggle itself: no filled box around it, the way
     PocketBase shows a record's true/false fields. The parts are styled once
     in styles/ark.css. -->
<Switch.Root
	{checked}
	{disabled}
	onclick={onlyTheSwitch}
	onCheckedChange={(details) => {
		checked = details.checked;
		onChange?.(details.checked);
	}}
>
	<Switch.Control><Switch.Thumb /></Switch.Control>
	<Switch.Label>{label}</Switch.Label>
	<Switch.HiddenInput />
</Switch.Root>
