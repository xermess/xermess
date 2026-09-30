# Fonts

Three families are bundled. `lib/styles/fonts.css` declares them, from the
faces in `lib/theme/theme.ts`; nothing else in the console refers to these
files.

```
chirp/             primary text              400, 500 and 700, Latin and Latin Extended
google-sans-code/  code: ids, keys, JSON     400, 500 and 700, Latin and Latin Extended
poppins/           display: the wordmark     600, Latin
```

Poppins is `--font-display`, for the header's wordmark and nothing that runs
to a sentence, so one weight in one alphabet is enough. It is an open font,
under the SIL Open Font License 1.1 (`poppins/OFL.txt`); the file is from the
`@fontsource/poppins` release.

Google Sans Code is `--font-mono`, so everything the panel sets in the code
face — a `Code`, a read-only id, a `CodeBlock` — is drawn in it.

One file is kept per alphabet and weight:

```
<family>-<alphabet>-<weight>-<style>.woff2
chirp-latin-500-normal.woff2
```

Each `@font-face` names its alphabet in `unicode-range`, so the browser
downloads only what the page shows: an English page never fetches the Latin
Extended file. Chirp is static; the panel draws its 600 weight from 700 and
uses the system face for scripts and styles Chirp does not cover.

Google Sans Code is an open font, under the SIL Open Font License 1.1
(`google-sans-code/OFL.txt`); the files are from the
`@fontsource/google-sans-code` release.

Chirp is X's typeface, not an open font. It is kept here for the console's
primary typeface; confirm the appropriate license before shipping the console
publicly.
