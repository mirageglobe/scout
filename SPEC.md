# Scout — Specification & Architecture

> a terminal file browser built with Go and the Charm library suite (Bubble Tea, Lip Gloss).

---

## 1. Overview

**Scout** is a two-pane terminal UI (TUI) file manager that lets you browse the filesystem, preview file contents, check git status at a glance, and hand off to an editor — all without leaving the terminal.

### Goals

| goal                                             | status |
| ------------------------------------------------ | ------ |
| two-pane layout (file list + preview)            | [x]    |
| keyboard navigation (j/k/h/l/g/G)                | [x]    |
| editor hand-off via `vim` + `tea.ExecProcess`    | [x]    |
| git status badges and branch display             | [x]    |
| styled borders via Lip Gloss                     | [x]    |
| modular architecture (`cmd/` + `internal/`)      | [x]    |
| chroma syntax highlighting in preview            | [x]    |
| time-aware color themes (7 themes, manual cycle) | [x]    |
| help overlay (`?`)                               | [x]    |
| live system stats (CPU, memory, clock)           | [x]    |
| hidden file toggle (`i`)                         | [x]    |
| collapsible file list pane (`tab`)               | [x]    |
| open with system default application (`o`)       | [x]    |
| scrollable preview pane (focus + j/k)            | [x]    |

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
│                          tea.Program                          │
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
    Cwd               string            // current working directory (absolute)
    Entries           []filesystem.Entry // sorted list of directory entries
    Cursor            int               // index of selected entry
    Width             int               // terminal width (from WindowSizeMsg)
    Height            int               // terminal height (from WindowSizeMsg)
    Preview           string            // pre-computed preview string for right pane
    PreviewScroll     int               // scroll offset for preview pane
    FocusRight        bool              // true when preview pane has keyboard focus
    ShowHelp          bool              // true when help overlay is visible
    ThemeIdx          int               // index into themes slice
    GitStatus         map[string]string // filename → git status code ("M", "+", "?", "!")
    GitBranch         string            // current git branch name
    ShowHidden        bool              // whether hidden (dot) files are shown
    ExplorerCollapsed bool              // whether file list pane is collapsed to 8 chars
    Stats             filesystem.Stats  // live CPU, memory, and directory size
    StatusMsg         string            // transient status or error message
    Err               error             // last error to display in-pane
}
```

`NewModel` sets the initial `ThemeIdx` by calling `ThemeForHour(time.Now().Hour())` so the theme is automatically chosen based on time of day.

### 3.2 Messages (Msg)

| message                      | source                | purpose                                              |
| ---------------------------- | --------------------- | ---------------------------------------------------- |
| `tea.WindowSizeMsg`          | Bubble Tea runtime    | captures terminal dimensions for layout              |
| `tea.KeyPressMsg`            | keyboard              | all navigation, actions, and quit signals            |
| `filesystem.DirLoadedMsg`    | `LoadDir` cmd         | delivers fresh entry list, git status, and branch    |
| `filesystem.GitRefreshMsg`   | `RefreshGit` cmd      | periodic git status and branch refresh               |
| `filesystem.TickMsg`         | `DoTick` cmd          | 2-second heartbeat; triggers stats and git refresh   |
| `filesystem.StatsMsg`        | `GetStats` cmd        | delivers live CPU, memory, and directory size        |
| `ui.EditorFinishedMsg`       | `tea.ExecProcess` cb  | signals vim has exited; triggers directory reload    |

### 3.3 Commands (Cmd)

#### `LoadDir(path string) tea.Cmd`
runs asynchronously. reads the directory with `os.ReadDir`, sorts entries (directories first, then alphabetical), fetches git status and branch in the same goroutine. returns `DirLoadedMsg`.

#### `RefreshGit() tea.Cmd`
re-fetches git status and branch for the current directory without reloading entries. returns `GitRefreshMsg`.

#### `GetStats(path string) tea.Cmd`
reads CPU usage via `runtime.ReadMemStats`, allocated memory, and directory size by walking the tree. returns `StatsMsg`.

#### `DoTick() tea.Cmd`
fires a `TickMsg` after a 2-second delay to drive the heartbeat for stats and git refresh.

#### `tea.ExecProcess(cmd, callback)`
suspends the TUI, forks `vim <file>`, and resumes on exit. the callback wraps the error in `EditorFinishedMsg`.

### 3.4 View

`View()` is a pure function of `Model` that produces a string. It:

1. if `ShowHelp` is true, renders the full-screen help overlay and returns early.
2. computes `leftWidth` (40 % or 8 chars if collapsed) and `rightWidth` (remaining space).
3. renders the **header bar**: app name, version, current time, CPU, and memory.
4. renders the **left pane**: path header → optional error → visible entry rows with scroll offset, git badges, and directory indicators.
5. renders the **right pane**: pre-computed `m.Preview` string (syntax-highlighted content or dir listing).
6. joins panes horizontally with `lipgloss.JoinHorizontal`.
7. appends a **status bar** with item count, position, git branch (`⎇ name`), and key hints.

### 3.5 Theming

Seven themes are defined in a `themes` slice. Each theme carries a name and an accent color:

| index | name           | accent    | auto-active hours |
| ----- | -------------- | --------- | ----------------- |
| 0     | Midnight       | `#875FFF` | 00:00 – 05:00     |
| 1     | Dawn           | `#FF8787` | 05:00 – 09:00     |
| 2     | Classic Amber  | `#FFAF00` | 09:00 – 12:00     |
| 3     | Electric Cyan  | `#00AFFF` | 12:00 – 17:00     |
| 4     | Safety Orange  | `#FF8700` | 17:00 – 20:00     |
| 5     | Evening        | `#FF5FAF` | 20:00 – 24:00     |
| 6     | Mono           | `#808080` | manual only       |

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

| key              | action                                         |
| ---------------- | ---------------------------------------------- |
| `j` / `↓`        | move cursor down                               |
| `k` / `↑`        | move cursor up                                 |
| `h` / `←` / `⌫`  | go to parent directory (or unfocus preview)    |
| `l` / `→`        | enter directory or focus preview pane          |
| `enter`          | enter directory or open file in vim            |
| `v`              | open file in vim                               |
| `o`              | open file with system default application      |
| `g`              | jump to top of list                            |
| `G`              | jump to bottom of list                         |
| `i`              | toggle hidden files                            |
| `tab`            | collapse / expand file list pane               |
| `t`              | cycle color theme                              |
| `?`              | show / hide help overlay                       |
| `q` / `ctrl+c`   | quit                                           |

---

## 6. Git Status Integration

`git.GetStatus(dir)` runs `git status --porcelain` and returns a `map[string]string`:

| porcelain code | badge | color  | meaning                  |
| -------------- | ----- | ------ | ------------------------ |
| `??`           | `?`   | dim    | untracked                |
| `A` / ` A`     | `+`   | green  | added / staged           |
| `M` / ` M`     | `M`   | orange | modified                 |
| other non-space| `!`   | red    | other change             |

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
│       └── main.go             # entry point
├── internal/
│   ├── filesystem/             # file ops, stats, tick, and entry types
│   ├── git/                    # git status and branch logic
│   └── ui/                     # MVU model, update, view, preview, themes
├── go.mod
├── go.sum
├── AGENT.md
├── CLAUDE.md
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

### steps

```bash
# 1. ensure you are on main and everything is merged + clean
git checkout main && git pull

# 2. export tokens (add these to ~/.zshrc or ~/.bashrc to avoid repeating)
export GITHUB_TOKEN=<your-scout-token>
export HOMEBREW_TAP_GITHUB_TOKEN=<your-tap-token>

# 3. tag the release (use semantic versioning)
git tag v0.3.0
git push origin v0.3.0

# 4. release (builds binaries, publishes GitHub release, updates homebrew formula)
make release
```

to test the release pipeline locally without publishing:

```bash
make release-dry
```

---

## 11. Design Decisions

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
