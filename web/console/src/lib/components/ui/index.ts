// The design system. One import path for every building block:
//
//   import { Button, Input, Select } from '$lib/components/ui';
//
// How they are styled — sizes, variants, palettes — is in README.md beside
// this file, and in styles/controls.css and styles/palettes.css.

// Controls
export { default as Button } from './Button.svelte';
export { default as IconButton } from './IconButton.svelte';
export { default as IconLink } from './IconLink.svelte';
export { default as CopyButton } from './CopyButton.svelte';
export { default as LinkButton } from './LinkButton.svelte';
export type { ColorPalette, ControlProps, Size, Variant } from './control';

// Form fields
export { default as Field } from './Field.svelte';
export { default as Input } from './Input.svelte';
export { default as PasswordInput } from './PasswordInput.svelte';
export { default as Select } from './Select.svelte';
export type { SelectOption } from './select';
export { default as Switch } from './Switch.svelte';
export { default as SwitchField } from './SwitchField.svelte';
export { default as Textarea } from './Textarea.svelte';
export { default as Checkbox } from './Checkbox.svelte';

// Surfaces and feedback
export { default as Alert } from './Alert.svelte';
export { default as Badge } from './Badge.svelte';
export { default as Card } from './Card.svelte';
export { default as Drawer } from './Drawer.svelte';
export { default as Panel } from './Panel.svelte';
export { default as StatCard } from './StatCard.svelte';
export { default as Tag } from './Tag.svelte';
export { default as Thumb } from './Thumb.svelte';
export { default as FormSection } from './FormSection.svelte';
export { default as SelectionBar } from './SelectionBar.svelte';
export { default as Tabs } from './Tabs.svelte';
export { default as Tooltip } from './Tooltip.svelte';

// Data and page furniture
export { default as DataTable } from './DataTable.svelte';
export { default as List } from './List.svelte';
export { default as ListItem } from './ListItem.svelte';
export type { Column } from './table';
export { default as Icon } from './Icon.svelte';
export { default as PageContainer } from './PageContainer.svelte';
export { default as PageHeader } from './PageHeader.svelte';
export { default as FontPicker } from './FontPicker.svelte';
export { default as ThemeToggle } from './ThemeToggle.svelte';
