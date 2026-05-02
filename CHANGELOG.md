# Changelog

all notable changes to scout are documented here.
format follows [keep a changelog](https://keepachangelog.com/en/1.1.0/).
versions follow [semantic versioning](https://semver.org/).

---

## [unreleased]

---

## [v0.6.0] — 2026-05-02

### added
- `± N` changed-file count badge in explorer stat line when git status is dirty
- renamed `root-focus` to `root-lock`; toggle key remapped from `f` → `l`
- removed vim-style navigation keys `j`/`k`/`h`/`l` — arrow keys are now the sole navigation bindings
- right-aligned human-readable file size column in file list (e.g. `1.2 KB`), matching dir child-count layout

### fixed
- `?` help overlay now dismisses on `?` keypress (previously any keypress dismissed it)

---

## [v0.5.0] — 2026-04-26

### added
- status message notification when the selected file or previewed file changes on disk
- rotating hint bar tips that cycle through helpful shortcuts during idle periods
- horizontal truncation for long preview lines with a dim ellipsis indicator
- widened collapsed explorer view for better visibility
- increased hint bar idle timeout from 10s to 60s for better readability
- consistent message bar styling with bracketed status tags (`[ok]`, `[error]`, `[info]`)

### fixed
- untracked files now highlighted with accent colour in file explorer
- auto-refresh not working — file changes on disk now reflected in real-time in file list and preview pane

---

## [v0.4.0] — 2026-04-24

### added
- persistent `scout ›` prompt in status line with state-aware messages (idle, loading, search, errors)
- animated loading spinner (`scout › ·/··/···`) during directory navigation
- toggle state indicators in hint bar — `i:hidden`, `f:root-lock`, `tab:explorer` render bold+accent when active
- cursor restores to previous folder when navigating to parent directory
- root-lock mode: restricts navigation to the launch directory (`f` to toggle)
- theme preference persisted to `~/.config/scout/config`
- homebrew release workflow via goreleaser (`make bump-patch`, `make push-tags`, `make release`)
- `context.Context` with timeout on all blocking operations (WatchDir 5s, LoadDir 10s, RefreshGit 5s, GetStats 5s)
- unit tests for filesystem utils, themes, and ui logic
- `CHANGELOG.md` with backfilled history

### fixed
- panic on `makeslice: len out of range` when preview scroll exceeded line count of newly loaded file

---

## [v0.3.0] — 2026-04-22

### added
- search in both explorer (`/` in left pane) and preview pane (`/` in right pane) with `n`/`N` navigation
- live directory watch: file list updates in the background without resetting cursor or scroll
- solarized dark and solarized light themes (indices 7 and 8)
- `make bump-patch` / `bump-minor` / `bump-major` targets for version tagging

---

## [v0.2.0] — 2026-04-20

### added
- goreleaser cross-platform build and homebrew-tap auto-update on tag push
- apache 2.0 license

---

## [v0.1.0] — 2026-04-20

### added
- dual-pane layout: file list (left) + file preview (right)
- keyboard navigation: `j`/`k` move, `h`/`l` enter/parent, `g`/`G` top/bottom
- chroma syntax highlighting in preview pane
- time-aware color themes (7 themes, auto-selected by hour, cycle with `t`)
- git status badges (`M`, `+`, `?`, `!`) and branch name in status bar
- live system stats: CPU, memory, directory size, clock in header
- help overlay (`?`) with full keybinding reference
- hidden file toggle (`i`)
- collapsible file list pane (`tab`)
- open with system default application (`o`)
- editor handoff via `$EDITOR` with `tea.ExecProcess` (TUI suspends and resumes cleanly)
- scrollable preview pane (focus with `l`, scroll with `j`/`k`)
- symlink indicator (`↳`) in file list
