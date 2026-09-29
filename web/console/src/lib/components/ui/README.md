# The design system

Everything the panel is built from. The point of this folder is that a page
never writes a colour, a height or a hover state — it names one.

```svelte
<Button>Save changes</Button>
<Button variant="subtle">Cancel</Button>
<Button colorPalette="danger" icon={RiDeleteBinLine}>Delete</Button>
<Button size="sm" loading={saving}>Saving…</Button>

<IconButton icon={RiRefreshLine} label="Refresh the data" />
<IconLink href="https://loginer.org/docs" icon={RiBookOpenLine} label="Documentation" />

<Input label="email" bind:value={email} type="email" required />
<Select label="Type" bind:value={type} options={['text', 'number', 'bool']} />
<Switch label="Verified" bind:checked={verified} />
<SearchInput label="Search users" bind:value={search} onsubmit={apply} oninput={debounced} />
```

## Menus and centred dialogs

A button in the header that opens a card of rows is a `DropdownMenu`, and its
rows are the four pieces below — the help menu and the account menu are both
built this way, so they read as one family:

```svelte
<DropdownMenu label="Help and resources">
	{#snippet trigger()}<Icon icon={RiQuestionLine} />{/snippet}

	<MenuGroup label="Resources">
		<MenuItem
			value="docs"
			icon={RiBookOpenLine}
			label="Documentation"
			description="Guides"
			href={docsUrl}
		/>
	</MenuGroup>
	<MenuSeparator />
	<MenuInfo icon={RiMailLine} label={email} />
	<MenuItem value="sign-out" icon={RiLogoutBoxRLine} label="Sign out" danger onSelect={ask} />
</DropdownMenu>
```

`MenuItem` does something (`onSelect`) or goes somewhere off this site
(`href`, which opens a tab and marks itself); `MenuInfo` only tells, and takes
no part in the arrow keys. `shape="avatar"` makes the button round.

A record, or anything else with a form to fill in, opens in a
`FullscreenDialog`; everything else that needs a dialog — a question, a
short choice — is a `Modal`, a card in the middle of
the window in three widths (`sm`, `md`, `lg`). Its `footer` holds the buttons,
`closable={false}` keeps it open while something it started is under way, and
a large one keeps one height, so tabs inside it do not make it jump.

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
select for `name` and form submission. The panel is portalled, so a dialog
that scrolls or a cell that hides its overflow cannot clip it.

`hint`, `error`, `required`, `disabled` and `readOnly` behave as they do on
`Input`: a read-only select shows its value and drops the chevron rather than
disappearing.

## DatePicker

A day, in the same box as every other field, with a calendar button at its
right end. The value is an ISO date — `"2026-03-12"`, or `''` for none — the
same string a native date input holds, so a form sends what it always did:

```svelte
<DatePicker label="Hired on" bind:value={hired} min="2020-01-01" max={today} clearable />
```

The day is shown and typed as `dd/mm/yyyy` in every locale: the server
renders the field too, and a date written two ways would be two different
pages. An ISO date pasted in is understood as well. The calendar's title steps
out to months and then years, for a day far from this one. `compact` is the
toolbar's version, as on `Input` and `Select`.

Pass `todayIso()` rather than `new Date().toISOString()` for today: the
second is the day in UTC, which is tomorrow or yesterday for part of every
day almost everywhere.

The calendar's styles are in `styles/ark.css`. The buttons at a field's
right end — a clear button, a select's arrow, a calendar, a password's eye —
share one size and one set of slots (`styles/fields.css`, "the right-hand
end"), so every field lines them up alike.

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

The values — every colour, light and dark, the scales, the sizes, the
typefaces — are in `lib/theme/theme.ts`. `bun run theme` renders it into
`styles/tokens.css` and `styles/fonts.css`, which are generated and checked
by `bun run check`; everything below refers to the tokens they declare.

| Looking for                                                          | It is in                |
| -------------------------------------------------------------------- | ----------------------- |
| a colour, a size, a radius, a font                                   | `lib/theme/theme.ts`    |
| what `danger` or `success` means                                     | `styles/palettes.css`   |
| buttons, icon buttons, link buttons                                  | `styles/controls.css`   |
| inputs, text areas, passwords, selects, dates                        | `styles/fields.css`     |
| switches, checkboxes, dialogs, tooltips, menus, dropdowns, calendars | `styles/ark.css`        |
| one component's own layout                                           | its own `<style>` block |

`ark.css` styles Ark UI's parts through the `data-scope` / `data-part`
attributes they render, so a menu or a switch looks the same wherever it is
used. Fields are the same idea in `fields.css`: every `Input`, `Textarea`,
`PasswordInput`, `Select` and `DatePicker` is a `.field-root` holding a `.field-box`, and
styling those once is what makes them match without any of them saying so.
Anything else shaped like a field — `SearchInput`, the command palette's
button — wears `.field-box` too and gets the same border, hover and focus.
The reusable field controls give their native inputs stable ids when the
caller does not provide one; explicit names are preserved where a control
supports native form values. This keeps browser autofill and form diagnostics
able to identify the fields.

## Corners

Three kinds of corner, and a field, a control or a surface names one of them
rather than a size:

| Token              | What wears it                                              |
| ------------------ | ---------------------------------------------------------- |
| `--radius-field`   | a box to type in — inputs, text areas, selects, dates      |
| `--radius-control` | something to press — buttons, chips, tags, filters: a pill |
| `--radius-surface` | something that holds — cards, panels, tables, dialogs      |

A field is the one thing with a tight corner, so a box to type in never reads
as a button. The values are in `theme.ts`; changing how round the panel is,
is changing those three. The sizes (`--radius-sm`, `-md`, `-lg`) are for the
small parts inside them — a menu's row, a checkbox, a `Code` — which follow
the thing they sit in rather than a kind of their own.

## Full-window dialogs

A record opens in a `FullscreenDialog`: the whole window, a few pixels in
from its edges and rounded, over a dimmed page. A bar across the top holds
the close button, the title and a tag beside it (`meta`, for an id or an
address), and the actions at its right; the body scrolls under it, in one
centred column.

```svelte
<FullscreenDialog bind:open title={user.name} meta={user.email} onsubmit={save}>
	<FormSection title="Account" description="Who they are.">…</FormSection>

	<DangerZone
		title="Delete this user"
		description="This cannot be undone."
		label="Delete"
		onclick={ask}
	/>

	{#snippet actions()}
		<Button variant="subtle" onclick={() => (open = false)}>Cancel</Button>
		<Button type="submit" loading={saving}>Save changes</Button>
	{/snippet}
</FullscreenDialog>
```

`onsubmit` makes the body a form, so the submit button among the actions
belongs to it and Enter saves. Without it the dialog is a plain container.
The actions are Cancel and the button that saves, and nothing else: a dialog
that only shows something has none — the close button, and Escape, are how it
is left — and the one thing that cannot be taken back is a `DangerZone` at the
foot of the form, never a button beside Save. `status` is a few words before
the buttons — what is unsaved, what is still missing. Pass either only when
there is something in it: an empty one still draws its bar. On a phone the
actions move to a bar along the bottom.

`FormSection` is what the body is built from, and what a settings page is
built from too — there it takes an `icon` beside its title and `meta` tags
under its description, saying where the values show. It lays itself out by its own
width: in a dialog or on a settings page its title and description sit on the
left and its fields in a card beside them; somewhere narrow — the flow
editor's inspector — it is a title over its fields. Give every section a
description; the left column is how a long form is scanned.

It is built when it opens and taken down once it has finished leaving, so a
dialog that opens always starts fresh — nothing scrolled, no tab still on the
one it was left on, no half-finished form from last time. A component inside
one may assume it is mounted for a single visit.

Arriving and leaving are 200ms, a rise of 12px and a fade. Both are in
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

## How a page is laid out

The dashboard layout draws every page in one frame: centred, no wider than
`--content-max-width` (1400px), and inset by `--page-gutter`, which grows from
16px on a phone to 32px on a wide monitor. It also stacks the page's children
with one gap. So a page writes no padding, no margins between its parts and no
maximum width of its own — it lists its parts in order:

```svelte
<PageHeader
	crumbs={['Dashboard', 'Users']}
	count={total}
	description="Everyone with an account here."
>
	{#snippet secondary()}<IconButton icon={RiRefreshLine} label="Refresh" />{/snippet}
	{#snippet actions()}<Button>New user</Button>{/snippet}
</PageHeader>

<Toolbar>
	<SearchInput label="Search users" bind:value={search} />
	<SegmentedControl label="Filter by verified" {options} value={verified} onChange={apply} />
	{#if role}<FilterChip title="Clear the role filter" onclear={clear}
			>role: <strong>{role}</strong></FilterChip
		>{/if}
</Toolbar>

<UserTable … />
```

A form is the exception to the full width: wrap it in `PageContainer`, which
holds it to a readable column (`--form-max-width`) that starts at the frame's
left edge, under the title — never centred on its own, so the title and the
first field line up and nothing moves from page to page. A page with tabs
keeps the tabs full width and puts the container inside the tab.

| Component          | What it is                                                                                                                                                                                    |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PageHeader`       | The trail, the title as the h1 with a `count` beside it, `meta` tags about the record, a `description`, and the buttons on the right.                                                         |
| `Toolbar`          | The row over a table: search first, held to a readable width, filters beside it, and an `end` snippet for the far side.                                                                       |
| compact fields     | `Input compact`, `Select compact`: a filter in a toolbar, at a control's height with its label as an inline prefix, so it lines up with the search box and the buttons.                       |
| `SegmentedControl` | A few mutually exclusive filters as one control at a field's height. The chosen one is filled in the brand colour; an option marked `reset` — "All" — is ruled off from the values beside it. |
| `FilterChip`       | A filter that came with the address, shown so it is not forgotten; clicking it clears it.                                                                                                     |
| `DataTable`        | A framed card: a tinted header of sentence-case labels, 52px rows, and a sideways scroll inside the frame when it needs one.                                                                  |

## Page building blocks

The pieces the Activity page and the account dialog are made of, after PocketBase's
settings and logs pages. Reach for these before writing a box of your own.

```svelte
<PageContainer>
	<!-- a readable column for a form: sm, md (default), lg -->

	<StatCard
		label="Users"
		value={42}
		icon={RiGroupLine}
		note="40 active · +3 this week"
		href={usersPage}
	/>

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

| Component       | What it is                                                                                                                                                 |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PageContainer` | A readable column for a form, from the frame's left edge.                                                                                                  |
| `PageHeader`    | Every page's heading: the trail above, the title, a count and a description, and the controls and actions on the title's row.                              |
| `Panel`         | A titled block: header strip with an icon and `meta`, then content. `flush` for lists.                                                                     |
| `StatCard`      | A metric: a number, the name it counts, and a one-line `note` beside it. `href` makes it a link; `tone` colours the note for the one that is an exception. |
| `List`          | Rows ruled off one under another. `bordered` makes it a box of its own.                                                                                    |
| `ListItem`      | A row: `lead` (usually a `Thumb`), a title and description or any content, and `end`.                                                                      |
| `Thumb`         | A small square holding an icon or a few letters.                                                                                                           |
| `Tag`           | A label in any palette; `dot` marks it with a coloured dot instead, `small` and `strong` size it.                                                          |
| `Thumb`         | Also a picture — `src`, falling back to the icon or letters if it fails — `circle` for a person, `tone="accent"` for the brand mark.                       |
| `Code`          | An id, a key, a code in the monospace face on a quiet tint; `tone="quiet"`, `truncate`.                                                                    |
| `CodeBlock`     | A block of code — JSON — highlighted, with a copy button; `title` gives it a head. Colours are the `--syntax-*` tokens in `theme.ts`.                      |
| `CodeEditor`    | A field for JSON: a text area over a highlighted copy of itself, so it edits like any field and reads like code.                                           |
| `Kbd`           | A key on the keyboard, for writing a shortcut down.                                                                                                        |
| `Note`          | The quiet sentence under a list or a form; brings no margin of its own.                                                                                    |
| `DangerZone`    | The one action on a form that cannot be taken back, at its foot, ruled in the danger colour, with what it costs beside the button.                         |
| `SaveBar`       | The Discard and Save buttons of a settings page that saves as a whole: shown once something changed, floating at the foot of the window.                   |
| `FieldGrid`     | Fields side by side, one column when it is narrow (by its own width); a child with `class="full"` spans the row.                                           |
| `Alert`         | A note about the page itself: what it shows or cannot show. Not the outcome of an action — that is `notify`.                                               |
| `notify`        | What an action came to: `notify.success(title)`, `notify.error(err, fallback)`. A toast in the corner, drawn by `<Toaster />` in the root layout.          |
