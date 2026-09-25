# The design system

Everything the panel is built from. The point of this folder is that a page
never writes a colour, a height or a hover state — it names one.

```svelte
<Button>Save changes</Button>
<Button variant="subtle">Cancel</Button>
<Button colorPalette="danger" icon={RiDeleteBinLine}>Delete</Button>
<Button size="sm" loading={saving}>Saving…</Button>

<IconButton icon={RiRefreshLine} label="Refresh the data" />
<IconLink href="https://xermess.org/docs" icon={RiBookOpenLine} label="Documentation" />

<Input label="email" bind:value={email} type="email" required />
<Select label="Type" bind:value={type} options={['text', 'number', 'bool']} />
<Switch label="Verified" bind:checked={verified} />
<SearchInput label="Search users" bind:value={search} onsubmit={apply} oninput={debounced} />
```

## Buttons that are links

`LinkButton` and `IconLink` are `Button` and `IconButton` with an anchor
inside instead of a button: somewhere to go rather than something to do, so
they can be opened in a new tab and the browser says where they lead. They
take the same `size`, `variant` and `colorPalette`, and `IconLink` carries the
tooltip `IconButton` does — an icon with no words needs the label either way.

An address off this site is the usual reason to reach for one, and those need
no `resolve()`; both components say so to the linting rule once, so the pages
using them do not have to.

## Select

`options` takes bare strings, and that is all a short list of keywords needs.
Give it objects instead when the rows should carry more:

```svelte
<Select
	label="Type"
	bind:value={type}
	options={[
		{ value: 'text', label: 'Plain text', icon: RiText },
		{ value: 'bool', label: 'Bool', icon: RiToggleLine, description: 'True or false' },
		{ value: 'json', label: 'JSON', disabled: true }
	]}
	placeholder="Pick one…"
	clearable
/>
```

It is Ark UI's Select underneath, so it comes with the keyboard — arrows,
home/end, and typing a few letters to jump to a row — and with a hidden native
select for `name` and form submission. The panel is portalled, so a drawer
that scrolls or a cell that hides its overflow cannot clip it.

`hint`, `error`, `required`, `disabled` and `readOnly` behave as they do on
`Input`: a read-only select shows its value and drops the chevron rather than
disappearing.

## Three props, and what they do

Every control takes the same three, and they combine: any size, in any
variant, in any palette.

| Prop           | Values                                        | What it decides                   |
| -------------- | --------------------------------------------- | --------------------------------- |
| `size`         | `sm` `md` `lg`                                | how tall it is — 35px, 45px, 52px |
| `variant`      | `solid` `subtle` `outline` `ghost` `plain`    | how much of the palette it uses   |
| `colorPalette` | `neutral` `danger` `success` `info` `warning` | which colours those are           |

A control sets them as data attributes and stops there:

```svelte
<button class="control" data-size={size} data-variant={variant} data-palette={colorPalette}>
```

`styles/controls.css` styles `.control` by those attributes, and it never
names a colour either — only `--palette-solid`, `--palette-subtle`,
`--palette-fg` and friends. `styles/palettes.css` fills those in per palette.

That indirection is the whole trick, and it is why:

- **changing how every button looks** is one rule in `controls.css`;
- **changing what "danger" means** is one block in `palettes.css`;
- **adding a palette** is copying a block — no component changes;
- `colorPalette="danger"` recolours solid, outline and ghost correctly,
  because each variant asks the palette rather than hard-coding red.

## Where each kind of styling lives

| Looking for                                               | It is in                |
| --------------------------------------------------------- | ----------------------- |
| a colour, a size, a radius, a font                        | `styles/tokens.css`     |
| what `danger` or `success` means                          | `styles/palettes.css`   |
| buttons, icon buttons, link buttons                       | `styles/controls.css`   |
| inputs, text areas, passwords, selects                    | `styles/fields.css`     |
| switches, checkboxes, drawers, tooltips, menus, dropdowns | `styles/ark.css`        |
| one component's own layout                                | its own `<style>` block |

`ark.css` styles Ark UI's parts through the `data-scope` / `data-part`
attributes they render, so a menu or a switch looks the same wherever it is
used. Fields are the same idea in `fields.css`: every `Input`, `Textarea`,
`PasswordInput` and `Select` is a `.field-root` holding a `.field-box`, and
styling those once is what makes them match without any of them saying so.
Anything else shaped like a field — `SearchInput`, the command palette's
button — wears `.field-box` too and gets the same border, hover and focus.
The reusable field controls give their native inputs stable ids when the
caller does not provide one; explicit names are preserved where a control
supports native form values. This keeps browser autofill and form diagnostics
able to identify the fields.

## Drawers

A record opens in a drawer, the way PocketBase opens one: a full-height panel
against the right edge, over a washed-out page.

```svelte
<Drawer bind:open title="Edit user" meta={user.id} onsubmit={save}>
	<FormSection title="Account">…</FormSection>

	{#snippet footer()}
		<Button variant="subtle" onclick={() => (open = false)}>Cancel</Button>
		<Button type="submit" loading={saving}>Save changes</Button>
	{/snippet}
</Drawer>
```

`onsubmit` makes the body a form, so the footer's submit button belongs to it
and Enter saves. Without it the drawer is a plain container.

It is built when it opens and taken down once it has finished leaving, so a
drawer that opens always starts fresh — nothing scrolled, no tab still on the
one it was left on, no half-finished form from last time. A component inside
one may assume it is mounted for a single visit.

Arriving and leaving are 200ms, and the panel travels the last 30px rather
than its whole width: PocketBase's motion, to the millisecond. Both are in
`styles/ark.css`, and they are two named animations rather than one reversed
on purpose — Ark waits for an exit animation only when it can see the
animation _name_ change, and a reversed one never changes it.

## Adding a component

1. If it is shaped like a button, give it `class="control"` and pass the three
   props through. Do not write colours.
2. If it is a form field, build it on `Field.svelte` — it already lays out the
   outlined box, the floating label, the required mark, and the hint or error
   under the box — and pass `filled` when the control holds a value, so the
   label rises out of its way. A field that brings its own anatomy, as
   `Select` and `PasswordInput` do, puts `class="field-root"` on its root, a
   `.field-box` around the label and the control, and `FieldText` inside the
   label, and `fields.css` does the rest.
3. If it is neither, its own `<style>` block is the right place, and it should
   use tokens rather than literal values.
4. Export it from `index.ts`.

## Page building blocks

The pieces the Activity page and account drawer are made of, after PocketBase's
settings and logs pages. Reach for these before writing a box of your own.

```svelte
<PageContainer>
	<!-- a centred column: sm, md (default), lg -->
	<PageHeader crumbs={['Account', 'Profile']}>
		<!-- breadcrumbs; the last is the h1 -->
		{#snippet secondary()}<IconButton icon={RiRefreshLine} label="Refresh" />{/snippet}
		{#snippet actions()}<Button>Save</Button>{/snippet}
	</PageHeader>

	<StatCard label="Users" value={42} icon={RiGroupLine} href={usersPage}>
		<Tag small>40 active</Tag>
	</StatCard>

	<Panel title="Sessions" icon={RiComputerLine} flush>
		{#snippet meta()}<Tag tone="success" dot>2 active</Tag>{/snippet}

		<List>
			<ListItem title="Chrome on macOS" description="127.0.0.1">
				{#snippet lead()}<Thumb icon={RiComputerLine} />{/snippet}
				{#snippet end()}<Tag tone="success" dot strong>Active</Tag>{/snippet}
			</ListItem>
		</List>
	</Panel>

	<Alert tone="warning">One administrator is locked out.</Alert>
</PageContainer>
```

| Component       | What it is                                                                                                                                                |
| --------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PageContainer` | A centred column for a page read top to bottom.                                                                                                           |
| `PageHeader`    | Every page's heading: `Dashboard / <section>`, with a link crumb back on a detail page, the count and controls beside it, and the actions at the far end. |
| `Panel`         | A titled block: header strip with an icon and `meta`, then content. `flush` for lists.                                                                    |
| `StatCard`      | A number with its name, an icon, tags under it, and a link when `href` is given.                                                                          |
| `List`          | Rows ruled off one under another. `bordered` makes it a box of its own.                                                                                   |
| `ListItem`      | A row: `lead` (usually a `Thumb`), a title and description or any content, and `end`.                                                                     |
| `Thumb`         | A small square holding an icon or a few letters.                                                                                                          |
| `Tag`           | A label in any palette; `dot` marks it with a coloured dot instead, `small` and `strong` size it.                                                         |
| `Alert`         | A message: `danger` (the default) is announced as an error; `warning`, `info`, `success` are notes.                                                       |
