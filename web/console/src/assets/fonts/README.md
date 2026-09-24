# Fonts

Every face the panel is set in, bundled rather than fetched, so a page looks
the same on every machine. `lib/styles/fonts.css` declares them; nothing else
refers to these files.

```
roboto/          text, the default         variable, 100–900
roboto-mono/     code, the default         variable, 100–700
ibm-plex-sans/   text, in the font picker  variable, 100–700
ibm-plex-mono/   code, with IBM Plex Sans  400, 500, 600, 700, and 400 italic
product-sans/    text, in the font picker  400, 500, 700, and 400 italic
```

One folder per family, named in kebab-case, and one file per alphabet:

```
<family>-<alphabet>-<weight>-<style>.woff2
roboto-latin-wght-normal.woff2          wght: a variable font, every weight
product-sans-cyrillic-500-normal.woff2  a static font, one weight
```

The alphabets are `latin`, `latin-ext`, `cyrillic` and `cyrillic-ext`, and
each `@font-face` names its own in `unicode-range`, so the browser downloads
only what the page shows: an English page never fetches a Cyrillic file.

Static families carry only the weights the panel sets — 400, 500, 600 and
700 — and a browser draws a missing 600 from the 700. To split a new font the
same way, [fonttools](https://github.com/fonttools/fonttools):

```sh
pyftsubset Font-Regular.ttf --flavor=woff2 --layout-features='*' \
  --unicodes="<the range from fonts.css>" \
  --output-file=family/family-latin-400-normal.woff2
```

The open families keep their `LICENSE` (SIL Open Font License) beside them.
Product Sans is Google's own and not openly licensed: it is here to try in
the font picker, and needs clearing before it ships in a release.
