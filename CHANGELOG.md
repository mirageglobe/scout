# Changelog

all notable changes to scout are documented here.
format follows [keep a changelog](https://keepachangelog.com/en/1.1.0/).
versions follow [semantic versioning](https://semver.org/).

---

## [unreleased]

### fixed
- `[explorer]` auto-refresh not working — file changes on disk are not reflected in the file list or preview pane without manual navigation

### docs
- adjust roadmap for file size column in `SPEC.md`

---

## [v0.4.0] — 2026-04-24

### added
- persistent `scout ›` prompt in status line with state-aware messages (idle, loading, search, errors)
- animated loading spinner (`scout › ·/··/···`) during directory navigation
- toggle state indicators in hint bar — `i:hidden`, `f:root-focus`, `tab:explorer` render bold+accent when active
- cursor restores to previous folder when navigating to parent directory
- root-focus mode: restricts navigation to the launch directory (`f` to toggle)
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
