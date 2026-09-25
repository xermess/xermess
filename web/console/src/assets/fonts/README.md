# Fonts

Chirp is the console's only bundled family. `lib/styles/fonts.css` declares
it; nothing else in the console refers to these files.

```
chirp/  primary text  400, 500 and 700, Latin and Latin Extended
```

One file is kept per alphabet and weight:

```
<family>-<alphabet>-<weight>-<style>.woff2
chirp-latin-500-normal.woff2
```

Each `@font-face` names its alphabet in `unicode-range`, so the browser
downloads only what the page shows: an English page never fetches the Latin
Extended file. Chirp is static; the panel draws its 600 weight from 700 and
uses the system face for scripts and styles Chirp does not cover.

Chirp is X's typeface, not an open font. It is kept here for the console's
primary typeface; confirm the appropriate license before shipping the console
publicly.
