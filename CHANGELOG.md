# Changelog

<!-- AI agents: add entries under the ## [Unreleased] header. Do NOT add version numbers or dates. Do NOT duplicate headings. The ## Known Bugs section must always stay pinned above ## [Unreleased]. Group entries under ### Added, ### Changed, ### Fixed, or ### Removed. Combine or update items refined within the same session. If the file exceeds 2000 lines, truncate the oldest releases. -->

## Known Bugs

- Typing directly in the freedom text box misbehaves; pasting works fine

## [Unreleased]

### Changed

- Pinned all GitHub Actions to full commit SHAs and bumped to their latest major versions (checkout v7, setup-go v6, setup-node v6, cache v6, upload-artifact v7, download-artifact v8, action-gh-release v3)
- Updated Go dependencies to latest stable: Wails v2.12.0 (now matching the CLI), chroma v2.27.0, glamour v2.0.1, mcp-go v0.55.1
- Upgraded glamour to v2 (module path is now `charm.land/glamour/v2`); replaced the removed `WithAutoStyle` with `WithEnvironmentConfig`, which honours `GLAMOUR_STYLE` and defaults to the dark theme
- Updated frontend and VSCode extension npm dependencies to latest stable; VSCode extension `npm audit` vulnerabilities reduced from 9 to 0 (serialize-javascript and diff resolved via overrides pending upstream mocha)
- Updated vscode-extension dependencies (@types/vscode, vscode engine, typescript-eslint, serialize-javascript)
- Removed an ineffective dynamic import in `frontend/src/App.jsx` (module was already statically imported)

### Added

- `make install`: installs M2E.app to /Applications (clearing quarantine attributes with `xattr -c`) and the m2e CLI to GOPATH/bin
- Around 730 new dictionary mappings imported from [tmgldn/en-mappings](https://github.com/tmgldn/en-mappings), kindly offered by its author in [issue #29](https://github.com/sammcj/m2e/issues/29). The import tooling and curated exclusion blocklist live in `scripts/import-en-mappings`
- Dictionary hygiene test (`tests/dictionary_hygiene_test.go`) enforcing invariants: lowercase single-token keys, no self-mappings, and no conversion target that is also a conversion source (prevents double-conversion chains and converting valid British English)
- Regression tests covering dictionary fixes, imported entries, blocklisted exclusions, British-text stability, and conversion idempotency

### Fixed

- Dictionary entries that produced misspellings or wrong inflections: `edema` now converts to `oedema` (was `edoema`), `pummeled` to `pummelled` (was `pummelling`), `yogurt` to `yoghurt` (was the archaic `yoghourt`), the `colorize` family to `colourise` (was `colourize`), and `diarization` to `diarisation` (was a self-mapping)
- Removed entries that converted correct British English into misspellings or American forms: `licensing` no longer becomes `licencing`, `bussing` no longer becomes `busing`
- Removed 39 entries with archaic or wrong targets, including the `gram`->`gramme` and `jail`->`gaol` families, `reflection`->`reflexion`, `siphon`->`syphon`, `ankle`->`ancle`, `lathe`->`laith`, `mocha`->`moka`, `slough`->`sleugh` and `stoichiometry`->`stoicheiometry`
- Removed dead entries that could never match at runtime: capitalised keys (`Americanization` etc., now lowercased so they convert), the multi-word key `pickup truck`, and four trailing-hyphen prefix keys
- `make build` no longer ships without the `m2e` CLI binary: `wails build` wipes `build/bin/` after the test phase had already built the CLI there, and Make's prerequisite de-duplication meant it was never rebuilt; the Go binaries now build after the Wails step
