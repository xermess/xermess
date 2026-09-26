<script lang="ts">
	import { RiBookOpenLine, RiGithubFill, RiKeyboardLine, RiQuestionLine } from 'svelte-remixicon';
	import { BRAND } from '$lib/brand';
	import {
		DropdownMenu,
		Icon,
		Kbd,
		MenuGroup,
		MenuInfo,
		MenuItem,
		MenuSeparator
	} from '$lib/components/ui';
</script>

<!-- Where to read about the product, behind one button: the bar carries one
     quiet icon instead of a row of them, and the menu has room to say what
     each link is and that it leaves the panel. It is the same on every width,
     so the links are never somewhere else on a phone. -->
<DropdownMenu label="Help and resources">
	{#snippet trigger()}<Icon icon={RiQuestionLine} size="1.125rem" />{/snippet}

	<MenuGroup label="Resources">
		<MenuItem
			value="documentation"
			icon={RiBookOpenLine}
			label="Documentation"
			description="Guides and the API reference"
			href={BRAND.docsUrl}
		/>
		<MenuItem
			value="github"
			icon={RiGithubFill}
			label="GitHub"
			description="Source, issues and releases"
			href={BRAND.githubUrl}
		/>
	</MenuGroup>

	<MenuSeparator />

	<!-- Not a link: the one shortcut worth knowing, written down where
	     somebody looking for help will find it. -->
	<MenuInfo icon={RiKeyboardLine} label="Search the panel">
		{#snippet end()}<Kbd>/</Kbd>{/snippet}
	</MenuInfo>
</DropdownMenu>
