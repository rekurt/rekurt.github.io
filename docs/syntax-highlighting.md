# Syntax highlighting

Project pages and the portfolio ship Highlight.js 11.12.0 locally. No CDN request or runtime dependency installation is needed. Existing code fonts, wrapping and backgrounds are preserved.

- README code uses its fenced language label. Unknown languages and plain-text reports remain readable without guessed syntax.
- Product examples use their explicit language. Terminal commands use a small CLI grammar for program names, flags, versions and package paths.
- Light project surfaces and dark examples use separate token palettes, tested against the actual code backgrounds for a minimum 4.5:1 contrast ratio.
- Highlighting preserves text exactly, so existing copy buttons keep using the original command. Escaped markup inside examples remains text.
- Already highlighted Astro/Shiki blocks are left alone. The initializer is safe to run again.

## Vendored source

Official CDN distribution: https://github.com/highlightjs/cdn-release/tree/dce7a3dab8f3fd586138ba6c5f29ecc19c02db9f

The local bundle combines unchanged `build/highlight.min.js` with `build/languages/dockerfile.min.js` from that commit. Original BSD-3-Clause notice is included as `highlight.LICENSE.txt`. Copies under `site/public/syntax/` and `internal/projectsite/assets/` must remain identical; CI verifies this.

## Verification

`node --test scripts/test-family-ui.mjs scripts/test-syntax.mjs` verifies language grammars, escaping, source preservation, identical assets and contrast. Browser QA verifies portfolio integration, copy behavior, idempotence and narrow-screen layout in light and dark themes.
