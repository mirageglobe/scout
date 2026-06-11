# Scout — Specification & Architecture

> a terminal file browser built with Go and the Charm library suite (Bubble Tea, Lip Gloss).

---

## 1. Overview

**Scout** is a two-pane terminal UI (TUI) file manager that lets you browse the filesystem, preview file contents, check git status at a glance, and hand off to an editor — all without leaving the terminal.

### Design Philosophy

**non-blocking, read-only by default.** scout never locks, writes to, or modifies the filesystem it browses. all directory reads and git queries are issued as async `tea.Cmd` values — they complete in the background and deliver results as messages, leaving the UI responsive at all times. this constraint keeps the codebase simple: no mutexes, no write paths, no risk of data loss.

### Goals

| goal                                                        | status |
| ----------------------------------------------------------- | ------ |
| two-pane layout (file list + preview)                       | [x]    |
| keyboard navigation (j/k/h/l/g/G)                           | [x]    |
| editor hand-off via `$EDITOR` + `tea.ExecProcess`           | [x]    |
| git status badges and branch display                        | [x]    |
| styled borders via Lip Gloss                                | [x]    |
| modular architecture (`cmd/` + `internal/`)                 | [x]    |
| chroma syntax highlighting in preview                       | [x]    |
| time-aware color themes (9 themes, manual cycle with `t`)   | [x]    |
| help overlay (`?`)                                          | [x]    |
| live system stats (CPU, memory, clock)                      | [x]    |
| hidden file toggle (`i`)                                    | [x]    |
| resizable explorer pane width cycle (`tab`)                 | [x]    |
| open with system default application (`o`)                  | [x]    |
| scrollable preview pane (nav with `l`, scroll with `j`/`k`) | [x]    |
| search in explorer and preview pane (`/`, `n`/`N`)          | [x]    |
| root-lock mode: lock navigation to launch directory (`l`)   | [x]    |
| persistent `scout ›` status prompt with loading spinner     | [x]    |
| cursor restores to previous folder on parent navigation     | [x]    |
| context-bounded async commands (no goroutine pile-up)       | [x]    |

---

## 2. Technology Stack

| dependency                                                   | version | purpose                              |
| ------------------------------------------------------------ | ------- | ------------------------------------ |
| `charm.land/bubbletea/v2`                                    | v2.0.6  | TUI runtime, MVU event loop          |
| `charm.land/lipgloss/v2`                                     | v2.0.3  | terminal styling and layout          |
| `github.com/alecthomas/chroma/v2`                            | v2.x    | syntax highlighting for file preview |
| Go stdlib (`os`, `os/exec`, `path/filepath`, `runtime`, ...) | —       | I/O, process execution, system stats |

> **no external bubbles components are used.** the file list is hand-rolled to give precise control over scrolling, padding, and git badge rendering.

---

## 3. Architecture

Scout follows the **Model-Update-View (MVU)** pattern enforced by Bubble Tea.

```
┌──────────────────────────────────────────────────────────────┐
│                          tea.Program                         │
│  ┌───────────┐   Msg   ┌──────────────┐   string   ┌───────┐ │
│  │   Init()  │────────▶│   Update()   │───────────▶│View() │ │
│  └───────────┘         └──────────────┘            └───────┘ │
│                                │                             │
│                                │ tea.Cmd                     │
│                                ▼                             │
│                     ┌──────────────────────┐                 │
│                     │    Async Commands    │                 │
│                     │  - LoadDir()         │                 │
│                     │  - RefreshGit()      │                 │
│                     │  - GetStats()        │                 │
│                     │  - DoTick()          │                 │
│                     │  - tea.ExecProcess() │                 │
│                     └──────────────────────┘                 │
└──────────────────────────────────────────────────────────────┘
```

### 3.1 Model

```go
type Model struct {
    Cwd                  string            // current working directory (absolute)
    Entries              []filesystem.Entry // sorted list of directory entries
    Cursor               int               // index of selected entry
    Width                int               // terminal width (from WindowSizeMsg)
    Height               int               // terminal height (from WindowSizeMsg)
    Preview              string            // pre-computed preview string for right pane
    PreviewScroll        int               // scroll offset for preview pane
    FocusRight           bool              // true when preview pane has keyboard focus
    ShowHelp             bool              // true when help overlay is visible
    ThemeIdx             int               // index into Themes slice
    GitStatus            map[string]string // filename → git status code ("M", "+", "?", "!")
    GitBranch            string            // current git branch name
    ShowHidden           bool              // whether hidden (dot) files are shown
    ExplorerWidthMode    int               // explorer pane width: 0/3 default, 1 sliver, 2 narrow, 4 wide
    Stats                filesystem.Stats  // live CPU, memory, and directory size
    StatusMsg            string            // transient status message shown in scout › prompt
    Err                  error             // last error to display in-pane
    SearchActive         bool              // true while user is typing a preview search query
    SearchQuery          string            // committed preview search term
    SearchInput          string            // in-progress buffer while SearchActive
    SearchMatches        []int             // preview line indices containing the query
    SearchMatchIdx       int               // current match index within SearchMatches
    ExplorerSearchActive bool              // true while user is typing an explorer search query
    ExplorerSearchInput  string            // current explorer search input
    RootLock             bool              // restrict navigation to RootPath
    RootPath             string            // the directory scout was launched from
    Loading              bool              // true while a LoadDir command is in-flight
    SpinnerFrame         int               // animation frame (0–2) for the scout › spinner
    PendingCursor        string            // entry name to restore cursor to after next load
}
```

`NewModel` sets `ThemeIdx` via `ThemeForHour(time.Now().Hour())` (or the saved config), and enables `RootLock` by default so navigation is locked to the launch directory until toggled with `l`.

### 3.2 Messages (Msg)

| message                      | source                | purpose                                                |
| ---------------------------- | --------------------- | ------------------------------------------------------ |
| `tea.WindowSizeMsg`          | Bubble Tea runtime    | captures terminal dimensions for layout                |
| `tea.KeyPressMsg`            | keyboard              | all navigation, actions, and quit signals              |
| `filesystem.DirLoadedMsg`    | `LoadDir` cmd         | delivers fresh entry list, git status, and branch      |
| `filesystem.DirWatchMsg`     | `WatchDir` cmd        | background poll result; updates entries without resetting cursor or scroll |
| `filesystem.GitRefreshMsg`   | `RefreshGit` cmd      | periodic git status and branch refresh                 |
| `filesystem.TickMsg`         | `DoTick` cmd          | 2-second heartbeat; triggers stats, git, and watch     |
| `filesystem.StatsMsg`        | `GetStats` cmd        | delivers live CPU, memory, and directory size          |
| `ui.EditorFinishedMsg`       | `tea.ExecProcess` cb  | signals editor has exited; triggers directory reload   |
| `ui.SpinnerTickMsg`          | `DoSpinnerTick` cmd   | 200ms tick that advances the scout › loading animation |

### 3.3 Commands (Cmd)

#### `LoadDir(path string) tea.Cmd`
runs asynchronously with a 10-second context timeout. reads the directory via `ReadDirContext`, sorts entries (directories first, then alphabetical), fetches git status and branch. stores the current directory name in `PendingCursor` before navigating to parent so the cursor is restored on load. returns `DirLoadedMsg`.

#### `WatchDir(path string) tea.Cmd`
background directory poll with a 5-second timeout. like `LoadDir` but returns `DirWatchMsg`, which is handled without resetting cursor or scroll — used by the 2-second tick to detect external filesystem changes.

#### `RefreshGit() tea.Cmd`
re-fetches git status and branch with a 5-second timeout. returns `GitRefreshMsg`.

#### `GetStats(path string) tea.Cmd`
reads allocated memory via `runtime.ReadMemStats`, directory size via `ReadDirContext`, and CPU via `ps`, all within a 5-second timeout. returns `StatsMsg`.

#### `DoTick() tea.Cmd`
fires a `TickMsg` after a 2-second delay to drive the heartbeat for stats, git refresh, and directory watch.

#### `DoSpinnerTick() tea.Cmd`
fires a `SpinnerTickMsg` after 200ms. only scheduled while `Loading` is true; each tick advances `SpinnerFrame` and reschedules itself.

#### `tea.ExecProcess(cmd, callback)`
suspends the TUI, forks `$EDITOR <file>`, and resumes on exit. the callback wraps the error in `EditorFinishedMsg`.

### 3.4 View

`View()` is a pure function of `Model` that produces a string. It:

1. if `ShowHelp` is true, renders the full-screen help overlay and returns early.
2. computes `leftWidth` (40 % or 8 chars if collapsed) and `rightWidth` (remaining space).
3. renders the **header bar**: app name, version, current time, CPU, and memory.
4. renders the **left pane**: path header → visible entry rows with scroll offset, git badges, and directory indicators.
5. renders the **right pane**: pre-computed `m.Preview` string (syntax-highlighted content or dir listing).
6. joins panes horizontally with `lipgloss.JoinHorizontal`.
7. renders the **`scout ›` status line** (always visible): shows loading spinner, search input/results, status messages, or a dim idle prompt.
8. renders the **hint bar**: git branch (`⎇ name`) and keybinding hints; active toggles (`i:hidden`, `l:root-lock`, `tab:explorer`) render bold+accent.

### 3.5 Theming

Nine themes are defined in a `Themes` slice. Each theme carries a name, accent, dim, text, and selected colours:

| index | name            | accent    | auto-active hours |
| ----- | --------------- | --------- | ----------------- |
| 0     | Classic Amber   | `#FFAF00` | 09:00 – 12:00     |
| 1     | Safety Orange   | `#FF8700` | 17:00 – 20:00     |
| 2     | Mono            | `#FFFFFF` | manual only       |
| 3     | Electric Cyan   | `#00AFFF` | 12:00 – 17:00     |
| 4     | Dawn            | `#FF8787` | 05:00 – 09:00     |
| 5     | Midnight        | `#875FFF` | 00:00 – 05:00     |
| 6     | Evening         | `#FF5FAF` | 20:00 – 24:00     |
| 7     | Solarized Dark  | `#268BD2` | manual only       |
| 8     | Solarized Light | `#268BD2` | manual only       |

`ThemeForHour(h int)` returns the correct index for the given hour. pressing `t` cycles forward through the slice with wrap-around.

---

## 4. Layout

```
┌─ scout v0.1.0 ──────────────────────────── 14:32  cpu 3%  mem 12MB ─┐
│                                                                     │
├─────────────────────────┬───────────────────────────────────────────┤
│  ~/projects/scout       │  · file: main.go                          │
│  ──────────────────     │  ──────────────────────────               │
│  M cmd/                 │  size:     16.0 KB                        │
│  · internal/            │  modified: 2026-04-18 17:00               │
│  · go.mod               │  mode:     -rw-r--r--                     │
│  · go.sum               │  ──────────────────────────               │
│  · README.md            │    1 │ package main                       │
│  · SPEC.md              │    2 │                                    │
│                         │    3 │ import (                           │
│                         │    …                                      │
├─────────────────────────┴───────────────────────────────────────────┤
│  6/8 items  · 14.2 KB  ⎇ main  │  q:quit  ?:help  j/k:nav  t:theme  │
└─────────────────────────────────────────────────────────────────────┘
```

- **header bar** — full-width, shows app name/version, clock, CPU, and memory.
- **left pane** — 40 % of terminal width (or 8 chars when collapsed), rounded border, theme accent.
- **right pane** — remaining terminal width, rounded border, same accent; dimmed border when unfocused.
- **status bar** — single line; item count, file size, git branch, and key hints.

---

## 5. Key Bindings

| key            | action                                             |
| :------------- | :------------------------------------------------- |
| `↓` / `↑`      | move cursor down / up                              |
| `←` / `⌫`      | nav to parent directory (or nav back from preview) |
| `→` / `enter`  | enter directory or nav to preview pane             |
| `g` / `G`      | jump to top / bottom of active pane                |
| `e`            | open file in editor                                |
| `o`            | open file with system default application          |
| `y`            | copy selected path to clipboard                    |
| `i`            | toggle hidden files                                |
| `l`            | toggle root-lock mode                              |
| `tab`          | cycle explorer pane width                          |
| `r`            | refresh preview                                    |
| `w`            | toggle word wrap in preview                        |
| `t`            | cycle color theme                                  |
| `T`            | toggle dark / light mode                           |
| `/`            | search active pane (`n` / `N`: next / prev)        |
| `esc`          | clear search                                       |
| `?`            | show / hide help overlay                           |
| `q` / `ctrl+c` | quit                                               |

---

## 6. Git Status Integration

`git.GetStatus(dir)` runs `git status --porcelain` and returns a `map[string]string`:

| porcelain code | badge | color  | meaning                  |
| -------------- | ----- | ------ | ------------------------ |
| `??`           | `?`   | accent | untracked                |
| `A` / ` A`     | `+`   | accent | added / staged           |
| `M` / ` M`     | `M`   | accent | modified                 |
| other non-space| `!`   | accent | other change             |

- nested paths (e.g. `subdir/file.go`) attribute the change to the top-level entry (`subdir`).
- renamed paths (`R  old -> new`) use the new (destination) name.
- if `git` is unavailable or the directory is not a repo, the map is empty and no badges are shown.

`git.GetBranch(dir)` runs `git rev-parse --abbrev-ref HEAD` and returns the branch name string.

---

## 7. Preview Logic

| selected entry  | preview content                                                             |
| --------------- | --------------------------------------------------------------------------- |
| **directory**   | icon, modified time, mode, child count, list of up to 20 children           |
| **text file**   | icon, size, modified time, mode, syntax-highlighted content with line numbers (first ~1000 lines or 32 KB) |
| **binary file** | icon, metadata, `(binary file – no preview)` message                       |

binary detection: any null byte (`0x00`) in the first 4 KB marks the file as binary.

syntax highlighting uses Chroma with the Dracula theme. the lexer is selected by file extension; falls back to plain text if unknown.

preview is regenerated whenever the cursor moves, a directory is loaded, or the window is resized. it is stored in `Model.Preview` as a pre-rendered string to keep `View()` allocation-light. when `FocusRight` is true, `j`/`k` scroll `PreviewScroll` instead of moving the cursor.

---

## 8. File Structure

```
scout/
├── cmd/
│   └── scout/
│       └── main.go                    # entry point
├── internal/
│   ├── filesystem/                    # file I/O, config, stats, tick, entry types
│   │   ├── config.go                  # theme config load/save (~/.config/scout/config)
│   │   ├── operations.go              # ReadDir, ReadDirContext, GetStats, DoTick, OpenWithSystem
│   │   ├── types.go                   # Entry, Stats, and Msg types
│   │   ├── utils.go                   # IsBinary, HumanSize, Truncate, VisibleLen
│   │   └── utils_test.go              # unit tests: IsBinary, HumanSize, Truncate
│   ├── git/
│   │   └── status.go                  # GetStatus (porcelain parser), GetBranch (context-aware)
│   └── ui/                            # MVU model, update, view, preview, themes
│       ├── header.go                  # RenderHeader
│       ├── help.go                    # RenderHelp overlay
│       ├── model.go                   # Model, Init, LoadDir, WatchDir, RefreshGit, DoSpinnerTick
│       ├── preview.go                 # BuildPreview (syntax highlight, dir listing)
│       ├── themes.go                  # Theme type, Themes slice, ThemeForHour
│       ├── themes_test.go             # unit tests: ThemeForHour
│       ├── update.go                  # Update (all state transitions)
│       ├── update_test.go             # unit tests: computeSearchMatches, dirEntriesChanged, clampedScrollFor
│       ├── version.go                 # Version constant (injected at build time)
│       └── view.go                    # View, RenderStatusLine
├── .github/workflows/
│   └── release.yml                    # goreleaser CI trigger on tag push
├── .goreleaser.yaml                   # cross-platform build + homebrew-tap config
├── go.mod
├── go.sum
├── AGENT.md                           # AI assistant guidelines (CLAUDE.md symlinks here)
├── CHANGELOG.md                       # hand-curated release history (keep-a-changelog format)
├── Makefile
├── README.md
└── SPEC.md
```

---

## 9. Build & Run

```bash
# build
make build
# or: go build -o scout cmd/scout/main.go

# run in current directory
./scout
```

### prerequisites
- Go 1.22+
- `vim` on `$PATH` for file opening
- `git` on `$PATH` for status badges (optional)

---

## 10. Releasing

> **for AI agents:** always ask the user which release method to use before proceeding (default: CI). present every command as a manual step for the user to run — do NOT execute `make bump-*`, `make push-tags`, `make release`, or `make update` autonomously. these commands affect shared git history and remote state. guide one phase at a time and wait for confirmation before continuing.

two release methods are available. **CI goreleaser is the default and preferred method.**

| method          | when to use                                              |
| :-------------- | :------------------------------------------------------- |
| CI goreleaser   | default; clean environment; audit trail in GitHub Actions |
| local goreleaser | CI is broken; no internet on runner; faster iteration    |

### prerequisites

- homebrew tap repo checked out locally alongside this repo: `../homebrew-tap`
- CI method: `GITHUB_TOKEN` in repo secrets (provided automatically by GitHub Actions)
- local method: `goreleaser` installed locally; `GITHUB_TOKEN` exported in shell

### version bump guide

| change type                                   | bump  | example          |
| :-------------------------------------------- | :---- | :--------------- |
| bug fixes only                                | patch | v0.3.0 → v0.3.1  |
| new user-facing features, no breaking changes | minor | v0.3.0 → v0.4.0  |
| breaking changes to behaviour or config       | major | v0.3.0 → v1.0.0  |

### phase 1 — prepare changelog (on feature branch)

shared by both methods. do not commit changelog or tag directly on main.

```bash
# 1. decide the target version using the bump guide above

# 2. update CHANGELOG.md:
#    - move all [unreleased] items under a new heading: ## [vX.Y.Z] — YYYY-MM-DD
#    - add a fresh empty [unreleased] section at the top

# 3. commit and push
git add CHANGELOG.md && git commit -m "docs: finalize changelog for vX.Y.Z" && git push

# 4. open a PR and merge into main
```

### phase 2 — tag (on main, after PR is merged)

shared by both methods. always use `make bump-*` — do NOT use `git tag` directly.

```bash
# 5. sync local main
git checkout main && git pull

# 6. tag the next version
make bump-patch   # bug fixes only         e.g. v0.3.0 -> v0.3.1
make bump-minor   # new features           e.g. v0.3.0 -> v0.4.0
make bump-major   # breaking changes       e.g. v0.3.0 -> v1.0.0
```

### phase 2a — publish via CI goreleaser (default)

```bash
# 7. push the tag — triggers GitHub Actions goreleaser
make push-tags
```

verify CI completes before proceeding: check https://github.com/mirageglobe/scout/actions

### phase 2b — publish via local goreleaser (alternative)

```bash
# 7. publish directly using local goreleaser (requires GITHUB_TOKEN exported)
make release
```

### phase 3 — update homebrew tap (after goreleaser completes)

shared by both methods. the tap lives at `../homebrew-tap`. run manually — do not automate.

```bash
# 8. switch to the tap repo and run the update target
cd ../homebrew-tap
gmake update FORMULA=scout VERSION=X.Y.Z   # VERSION without the v prefix, e.g. 0.8.0
# note: gmake required — macOS ships with GNU make 3.81 which lacks .ONESHELL support
```

the `update` target:
- fetches `scout_X.Y.Z_checksums.txt` from the GitHub release
- patches `Formula/scout.rb` — version string, download urls (including tag path), and all sha256 values
- commits with `feat: update scout formula to vX.Y.Z`
- pushes to origin main

users then upgrade via:
```bash
brew upgrade mirageglobe/tap/scout
```

### local validation (optional)

```bash
make release-dry   # dry-run goreleaser: builds binaries and archives locally, no publish
```

### troubleshooting

**release fails with `422 Validation Failed — tag_name already_exists`**

this happens when a previous goreleaser run partially created a GitHub release for the same tag. goreleaser cannot overwrite an existing release.

fix: delete the partial release and retrigger — run these commands manually:

```bash
make release-reset   # deletes any existing GitHub release for the current tag
make push-tags       # CI method: retriggers goreleaser via tag push
# make release       # local method: run goreleaser directly instead
```

---

## 11. Roadmap

### bugs

- [x] `[explorer]` auto-refresh not working — file changes on disk are not reflected in the file list or preview pane without manual navigation  [medium]

### near term

- [x] `[site]` add github pages website — astro source in `/site`, ci builds and deploys to github pages environment (no branch); workflow triggers on `/site/**` changes  [easy]
- [ ] `[ai]` detect locally running ollama instance and connect for in-app chat — probe `http://localhost:11434` on startup; if available, expose a chat panel keybinding to open a conversational interface backed by the detected model  [hard]
- [x] `[explorer]` consider showing in file pane, the number of changed files  [easy]
- [x] `[explorer]` update naming of command `root-focus` to `root-lock`  [easy]
- [x] `[explorer]` ls all files in current directory  [easy]
- [x] `[preview]` syntax highlighting  [medium]
- [x] `[ui]` time-aware color themes  [medium]
- [x] `[ui]` help overlay  [easy]
- [x] `[ui]` system stats in header (CPU, memory, clock)  [medium]
- [x] `[git]` git branch display in status bar  [easy]
- [x] `[explorer]` collapsible file list pane  [medium]
- [x] `[explorer]` identify symlinks in file list (e.g. with @ or ↳ symbol)  [easy]
- [x] `[explorer]` respect `$EDITOR` environment variable for editor handoff  [easy]
- [x] `[preview]` preview auto-refresh or manual refresh key to reload files changed by external processes  [medium]
- [x] `[config]` create saved local configs to support theme save  [medium]
- [x] `[explorer]` focus command: restrict navigation to root directory where scout was launched (no escaping to parent)  [medium]
- [x] `[ui]` visible status/activity indicator above the hint bar (`scout ›` persistent prompt with spinner and state-aware messages)  [medium]
- [x] `[explorer]` navigating to parent directory should restore cursor focus to the folder you came from  [medium]
- [x] `[ui]` toggle state indicators in the hint bar (bold accent on i:hidden, l:root-lock, tab:explorer when active)  [easy]
- [x] `[explorer]` add context.Context with timeout to WatchDir, LoadDir, RefreshGit, and GetStats to prevent goroutine pile-up on slow or hung mounts  [medium]
- [x] `[preview]` preview pane text wrapping — long lines truncated at pane boundary with a dim-styled `…` indicator; horizontal scroll deferred (use `e` to open in `$EDITOR`)  [easy]
- [x] `[preview]` stale preview notification — preview auto-refreshes on file change via dirEntriesChanged ModTime check; no separate notification needed  [easy]
- [x] `[ui]` rotating hint bar tips — normal bar shown at rest; after 10s idle, cycles once through 12 friendly tips (5s each) then returns to normal; any keypress cancels and resets  [medium]
- [x] `[ui]` consistent message bar styling — uniform dim style for all messages; bracketed tag prefix `[error]`, `[ok]`, `[info]` distinguishes type; no colour emphasis on body or tag  [easy]
- [x] `[preview]` increase truncation for text files to 1200 lines (currently ~1000 lines or 32 KB) [easy]
- [x] `[explorer]` mouse click to select and navigate files in the explorer pane  [medium]
- [x] `[preview]` scrollbar indicator in the preview pane showing scroll position  [easy]
- [x] `[explorer]` mouse wheel scroll in the file explorer pane  [easy]

### ideas

- [x] `[explorer]` copy file path to clipboard — single keypress copies the full path of the selected entry to the system clipboard (`pbcopy`/`xclip`)  [easy]
- [ ] `[explorer]` fuzzy file search  [hard]
- [ ] `[ui]` ambiguous-width Unicode rendering in CJK locales — characters like `›`, `⎇`, `▸` may render as 2-cell wide in terminals with `RUNEWIDTH_EASTASIAN=1`, causing column misalignment; add `SCOUT_UNICODE_SAFE=1` env var that swaps the symbol set to narrow-safe ASCII alternatives at startup  [medium]
- [ ] `[git]` git diff preview — when selected file has an `M` badge, show `git diff` output in the preview pane  [medium]
- [ ] `[git]` git log preview — when selecting a file, offer a keypress to show `git log --oneline` for that file in the preview pane  [medium]
- [ ] `[preview]` mouse drag text selection in preview viewport — click-drag highlights lines; releasing the mouse copies the selected text to the system clipboard  [medium]
- [x] `[install]` curl binary install/upgrade script — provide a one-liner script that detects OS/arch, downloads the correct tarball from the GitHub release, and places the binary in `~/.local/bin` or `/usr/local/bin`; re-running the script upgrades to the latest release; alternative to Homebrew for non-Mac or Homebrew-free environments  [medium]
- [x] `[explorer]` four-width explorer pane — `tab` from default (~40 cols) enters a sub-cycle: sliver (5 cols) → narrow (13 cols) → wide (50%) → sliver; default is an entry point only and is never revisited via tab; replaces the binary collapse toggle; `tab:explorer` hint bar indicator activates when not in default mode  [medium]
- [x] `[explorer]` file size column in the file list — show human-readable size for files alongside the name (data already available via `Entry.Info`)  [easy]
- [x] `[ui]` dark / light mode — detect terminal background via OSC 11 query (`tea.BackgroundColorMsg`); auto-select a light theme when on a light background, dark when dark; `t` continues to cycle within the active mode  [medium]
- [x] `[ui]` manual dark/light mode toggle — `T` (shift+t) switches between dark and light mode pools and selects the first theme in the new pool; `t` continues to cycle within the active pool  [easy]
- [x] `[preview]` theme-aware chroma syntax highlighting — map each scout theme to a named chroma style (e.g. `dracula` for dark, `github` for light) so syntax colours complement the active palette; switch style when theme changes  [medium]
- [x] `[ui]` more light themes — add 2–3 light-background palettes (e.g. light mono, light warm, Github Light) so light-mode users have themes to cycle through  [easy]
- [x] `[ui]` context-aware help overlay — filter displayed keybindings to only those relevant to the active pane; explorer-only keys (e, o, i, l) hidden when preview is focused, preview-only keys (r, n/N) hidden when explorer is focused  [easy]
- [x] `[preview]` word-wrap toggle — keypress (e.g. `w`) wraps long lines in the preview pane to fit the pane width instead of truncating with `…`; wrap state persists across file navigation until toggled off  [easy]
- [x] `[cli]` update check — `scout --version` compares the running version against the latest GitHub release tag via the API and prints a notice if an upgrade is available  [easy]

---

## 12. Design Decisions

| decision                             | rationale                                                                                  |
| ------------------------------------ | ------------------------------------------------------------------------------------------ |
| pre-computed `Preview` string        | avoids re-allocating on every `View()` call; recomputes only on state changes              |
| `AltScreen = true`                   | uses the secondary terminal buffer so shell history is not polluted                        |
| `tea.ExecProcess` for vim            | idiomatic Bubble Tea way to suspend TUI, hand off stdin/stdout, and resume cleanly         |
| no `bubbles/list` component          | gives full control over git badge rendering, scrolling, and padding behaviour              |
| directories-first sort               | standard filesystem browser convention; reduces cognitive load                             |
| 128 KB / 2500-line preview cap       | prevents large files from blocking the UI during preview generation                       |
| wrap-aware scroll via `previewDisplayLineCount` | `view.go` expands raw lines into display lines (wrap); `update.go` must compute the same count to bound `PreviewScroll` correctly — `previewDisplayLineCount` approximates it using visible rune count / pane width without re-running `lipgloss.Wrap` on every keypress |
| time-based theme auto-selection      | reduces manual configuration; theme still switchable at runtime with `t`                  |
| 2-second tick for stats and git      | low enough overhead to feel live; high enough to avoid hammering the filesystem            |
| `runtime.ReadMemStats` for memory    | zero-dependency way to surface allocated heap without external tooling                     |
