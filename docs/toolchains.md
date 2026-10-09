# Compiler and analysis toolchains

TypeScript 7 is installed as `typescript-native` and is invoked explicitly by the typecheck command. TypeScript 6 remains under the `typescript` package name for tools using the JavaScript compiler API (Astro Check, TypeDoc, tsup, typescript-eslint). This preserves their supported peer ranges; no peer overrides or forced installs are needed.

`npm run check` runs Astro Check with its supported compiler API and then the native TypeScript 7 compiler. Astro diagnostics and Browser QA remain required.
