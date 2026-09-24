<script lang="ts">
	import { PasswordInput } from '@ark-ui/svelte/password-input';
	import { RiEyeLine, RiEyeOffLine } from 'svelte-remixicon';
	import FieldText from './FieldText.svelte';
	import Icon from './Icon.svelte';

	type Props = {
		label: string;
		value: string;
		disabled?: boolean;
		required?: boolean;
		autocomplete?: 'current-password' | 'new-password';
	};

	let {
		label,
		value = $bindable(''),
		disabled,
		required,
		autocomplete = 'current-password'
	}: Props = $props();
</script>

<!-- The same outlined box and floating label as every other field
     (styles/fields.css), with the show/hide button inside it. -->
<PasswordInput.Root class="field-root" {disabled} {required}>
	<div class="field-box" data-float={value !== '' || undefined}>
		<PasswordInput.Label><FieldText {label} {required} /></PasswordInput.Label>
		<PasswordInput.Control>
			<PasswordInput.Input bind:value {autocomplete} />
			<PasswordInput.VisibilityTrigger aria-label="Show password">
				<!-- The indicator renders `children` while the password is
				     visible and `fallback` while it is hidden. -->
				<PasswordInput.Indicator>
					{#snippet fallback()}
						<Icon icon={RiEyeLine} size="1.0625rem" />
					{/snippet}
					<Icon icon={RiEyeOffLine} size="1.0625rem" />
				</PasswordInput.Indicator>
			</PasswordInput.VisibilityTrigger>
		</PasswordInput.Control>
	</div>
</PasswordInput.Root>
