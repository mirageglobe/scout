# Scout — Specification & Architecture

> a terminal file browser built with Go and the Charm library suite (Bubble Tea, Lip Gloss).

---

## 1. Overview

**Scout** is a two-pane terminal UI (TUI) file manager that lets you browse the filesystem, preview file contents, check git status at a glance, and hand off to an editor — all without leaving the terminal.

### Goals

| goal                                                        | status |
| ----------------------------------------------------------- | ------ |
| two-pane layout (file list + preview)                       | [x]    |
| keyboard navigation (j/k/h/l/g/G)                          | [x]    |
| editor hand-off via `$EDITOR` + `tea.ExecProcess`           | [x]    |
| git status badges and branch display                        | [x]    |
| styled borders via Lip Gloss                                | [x]    |
| modular architecture (`cmd/` + `internal/`)                 | [x]    |
| chroma syntax highlighting in preview                       | [x]    |
| time-aware color themes (9 themes, manual cycle with `t`)   | [x]    |
| help overlay (`?`)                                          | [x]    |
| live system stats (CPU, memory, clock)                      | [x]    |
| hidden file toggle (`i`)                                    | [x]    |
| collapsible file list pane (`tab`)                          | [x]    |
| open with system default application (`o`)                  | [x]    |
| scrollable preview pane (nav with `l`, scroll with `j`/`k`) | [x]    |
| search in explorer and preview pane (`/`, `n`/`N`)          | [x]    |
| root-focus mode: lock navigation to launch directory (`f`)  | [x]    |
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
    ExplorerCollapsed    bool              // whether file list pane is collapsed to 8 chars
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
    RootFocus            bool              // restrict navigation to RootPath
    RootPath             string            // the directory scout was launched from
    Loading              bool              // true while a LoadDir command is in-flight
    SpinnerFrame         int               // animation frame (0–2) for the scout › spinner
    PendingCursor        string            // entry name to restore cursor to after next load
}
```

`NewModel` sets `ThemeIdx` via `ThemeForHour(time.Now().Hour())` (or the saved config), and enables `RootFocus` by default so navigation is locked to the launch directory until toggled with `f`.

### 3.2 Messages (Msg)

| message                      | source                | purpose                                              |
| ---------------------------- | --------------------- | ---------------------------------------------------- |
| `tea.WindowSizeMsg`          | Bubble Tea runtime    | captures terminal dimensions for layout              |
| `tea.KeyPressMsg`            | keyboard              | all navigation, actions, and quit signals            |
| `filesystem.DirLoadedMsg`    | `LoadDir` cmd         | delivers fresh entry list, git status, and branch    |
| `filesystem.DirWatchMsg`     | `WatchDir` cmd        | background poll result; updates entries without resetting cursor or scroll |
| `filesystem.GitRefreshMsg`   | `RefreshGit` cmd      | periodic git status and branch refresh               |
| `filesystem.TickMsg`         | `DoTick` cmd          | 2-second heartbeat; triggers stats, git, and watch   |
| `filesystem.StatsMsg`        | `GetStats` cmd        | delivers live CPU, memory, and directory size        |
| `ui.EditorFinishedMsg`       | `tea.ExecProcess` cb  | signals editor has exited; triggers directory reload |
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
8. renders the **hint bar**: git branch (`⎇ name`) and keybinding hints; active toggles (`i:hidden`, `f:root-focus`, `tab:explorer`) render bold+accent.

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
│                                                                       │
├─────────────────────────┬─────────────────────────────────────────── ┤
│  ~/projects/scout       │  · file: main.go                           │
│  ──────────────────     │  ──────────────────────────                │
│  M cmd/                 │  size:     16.0 KB                         │
│  · internal/            │  modified: 2026-04-18 17:00                │
│  · go.mod               │  mode:     -rw-r--r--                      │
│  · go.sum               │  ──────────────────────────                │
│  · README.md            │    1 │ package main                        │
│  · SPEC.md              │    2 │                                     │
│                         │    3 │ import (                            │
│                         │    …                                       │
├─────────────────────────┴─────────────────────────────────────────── ┤
│  6/8 items  · 14.2 KB  ⎇ main  │  q:quit  ?:help  j/k:nav  t:theme  │
└───────────────────────────────────────────────────────────────────── ┘
```

- **header bar** — full-width, shows app name/version, clock, CPU, and memory.
- **left pane** — 40 % of terminal width (or 8 chars when collapsed), rounded border, theme accent.
- **right pane** — remaining terminal width, rounded border, same accent; dimmed border when unfocused.
- **status bar** — single line; item count, file size, git branch, and key hints.

---

## 5. Key Bindings

| key              | action                                             |
| ---------------- | -------------------------------------------------- |
| `j` / `↓`        | move cursor down                                   |
| `k` / `↑`        | move cursor up                                     |
| `h` / `←` / `⌫`  | nav to parent directory (or nav back from preview) |
| `l` / `→`        | enter directory or nav to preview pane             |
| `enter`          | enter directory or open file in editor             |
| `e`              | open file in editor                                |
| `o`              | open file with system default application          |
| `g`              | jump to top of list                                |
| `G`              | jump to bottom of list                             |
| `i`              | toggle hidden files                                |
| `f`              | toggle root-focus mode                             |
| `tab`            | collapse / expand file list pane                   |
| `t`              | cycle color theme                                  |
| `?`              | show / hide help overlay                           |
| `q` / `ctrl+c`   | quit                                               |

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

Scout uses [goreleaser](https://goreleaser.com) to build cross-platform binaries, publish a GitHub release, and auto-update the [homebrew-tap](https://github.com/mirageglobe/homebrew-tap) formula.

### prerequisites

- [`goreleaser`](https://goreleaser.com/install/) installed (`brew install goreleaser`)
- two GitHub personal access tokens (classic or fine-grained, with `contents: write` scope):
  - `GITHUB_TOKEN` — write access to this repo (scout)
  - `HOMEBREW_TAP_GITHUB_TOKEN` — write access to the homebrew-tap repo

### version bump guide

| change type                                      | bump    | example         |
| ------------------------------------------------ | ------- | --------------- |
| bug fixes only                                   | patch   | v0.3.0 → v0.3.1 |
| new user-facing features, no breaking changes    | minor   | v0.3.0 → v0.4.0 |
| breaking changes to behaviour or config format   | major   | v0.3.0 → v1.0.0 |

### steps

```bash
# 1. merge the release branch into main via PR, then pull
git checkout main && git pull

# 2. export tokens (add these to ~/.zshrc or ~/.bashrc to avoid repeating)
export GITHUB_TOKEN=<your-scout-token>
export HOMEBREW_TAP_GITHUB_TOKEN=<your-tap-token>

# 3. update CHANGELOG.md — move [unreleased] items under the new version heading
#    e.g. ## [v0.4.0] — 2026-05-01
#    add a fresh empty [unreleased] section at the top for the next cycle

# 4. commit the changelog update
git add CHANGELOG.md && git commit -m "docs: update changelog for vX.Y.Z"

# 5. tag the next version (patch / minor / major — see guide above)
make bump-minor

# 6. push the tag to origin
make push-tags

# 7. build binaries, publish github release, and update homebrew formula
make release
```

to test the release pipeline locally without publishing:

```bash
make release-dry
```

### troubleshooting

**release fails with `422 Validation Failed — tag_name already_exists`**

this happens when a previous goreleaser run partially created a GitHub release for the same tag (e.g. interrupted mid-upload). goreleaser cannot overwrite an existing release.

fix: delete the partial release(s) and re-run:

```bash
make release-reset   # deletes any existing GitHub release for the current tag
make release
```

---

## 11. Roadmap

### bugs

- [x] `[explorer]` auto-refresh not working — file changes on disk are not reflected in the file list or preview pane without manual navigation  [medium]

### near term

- [x] ls all files in current directory
- [x] syntax highlighting
- [x] time-aware color themes
- [x] help overlay
- [x] system stats in header (CPU, memory, clock)
- [x] git branch display in status bar
- [x] collapsible file list pane
- [x] identify symlinks in file list (e.g. with @ or ↳ symbol)
- [x] respect `$EDITOR` environment variable for editor handoff
- [x] preview auto-refresh or manual refresh key to reload files changed by external processes
- [x] create saved local configs to support theme save
- [x] focus command: restrict navigation to root directory where scout was launched (no escaping to parent)
- [x] visible status/activity indicator above the hint bar (`scout ›` persistent prompt with spinner and state-aware messages)
- [x] navigating to parent directory should restore cursor focus to the folder you came from
- [x] toggle state indicators in the hint bar (bold accent on i:hidden, f:root-focus, tab:explorer when active)
- [x] add context.Context with timeout to WatchDir, LoadDir, RefreshGit, and GetStats to prevent goroutine pile-up on slow or hung mounts
- [x] preview pane text wrapping — long lines truncated at pane boundary with a dim-styled `…` indicator; horizontal scroll deferred (use `e` to open in `$EDITOR`)
- [ ] stale preview notification — when the previewed file is modified on disk, display an inline prompt asking the user to refresh rather than silently showing stale content
- [x] rotating hint bar tips — normal bar shown at rest; after 10s idle, cycles once through 12 friendly tips (5s each) then returns to normal; any keypress cancels and resets
- [ ] consistent message bar styling — remove highlight/colour emphasis from messages; use a uniform text style with a leading icon or Unicode glyph to distinguish message types (e.g. `!` error, `i` hint, `·` status)
- [ ] ambiguous-width Unicode rendering in CJK locales — characters like `›`, `⎇`, `▸` may render as 2-cell wide in terminals with `RUNEWIDTH_EASTASIAN=1`, causing column misalignment; investigate per-terminal detection or a layout tolerance mode

### ideas

- [ ] `[explorer]` file size column in the file list — show human-readable size for files alongside the name (data already available via `Entry.Info`)  [easy]
- [ ] `[explorer]` copy file path to clipboard — single keypress copies the full path of the selected entry to the system clipboard (`pbcopy`/`xclip`)  [easy]
- [ ] fuzzy file search
- [ ] preview images

---

## 12. Design Decisions

| decision                             | rationale                                                                                  |
| ------------------------------------ | ------------------------------------------------------------------------------------------ |
| pre-computed `Preview` string        | avoids re-allocating on every `View()` call; recomputes only on state changes              |
| `AltScreen = true`                   | uses the secondary terminal buffer so shell history is not polluted                        |
| `tea.ExecProcess` for vim            | idiomatic Bubble Tea way to suspend TUI, hand off stdin/stdout, and resume cleanly         |
| no `bubbles/list` component          | gives full control over git badge rendering, scrolling, and padding behaviour              |
| directories-first sort               | standard filesystem browser convention; reduces cognitive load                             |
| 32 KB / 1000-line preview cap        | prevents large files from blocking the UI during preview generation                       |
| time-based theme auto-selection      | reduces manual configuration; theme still switchable at runtime with `t`                  |
| 2-second tick for stats and git      | low enough overhead to feel live; high enough to avoid hammering the filesystem            |
| `runtime.ReadMemStats` for memory    | zero-dependency way to surface allocated heap without external tooling                     |
