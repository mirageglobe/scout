# Scout — Specification & Architecture

> A terminal file browser built with Go and the Charm library suite (Bubble Tea, Lip Gloss).

---

## 1. Overview

**Scout** is a two-pane terminal UI (TUI) file manager that lets you browse the filesystem, preview file contents, check git status at a glance, and hand off to an editor — all without leaving the terminal.

### Goals

| Goal | Status |
|---|---|
| Two-pane layout (file list + preview) | ✅ |
| Keyboard navigation (j/k/h/enter/g/G) | ✅ |
| Editor hand-off via `vim` + `tea.ExecProcess` | ✅ |
| Git status integration (`[M]` / `[?]`) | ✅ |
| Styled borders via Lip Gloss | ✅ |
| Single-file architecture (`main.go`) | ✅ |

---

## 2. Technology Stack

| Dependency | Version | Purpose |
|---|---|---|
| `charm.land/bubbletea/v2` | v2.0.6 | TUI runtime, MVU event loop |
| `charm.land/lipgloss/v2` | v2.0.3 | Terminal styling & layout |
| Go stdlib (`os`, `os/exec`, `path/filepath`, `sort`, `strings`, `fmt`) | — | I/O, process execution, text |

> **No external bubbles components are used.** The file list is hand-rolled to give precise control over scrolling, padding, and git badge rendering.

---

## 3. Architecture

Scout follows the **Model-Update-View (MVU)** pattern enforced by Bubble Tea.

```
┌──────────────────────────────────────────────────────────┐
│                        tea.Program                        │
│  ┌───────────┐   Msg   ┌──────────┐   tea.View   ┌──────┐│
│  │   Init()  │────────▶│ Update() │─────────────▶│View()││
│  └───────────┘         └──────────┘              └──────┘│
│                             │                            │
│                             │ tea.Cmd                    │
│                             ▼                            │
│                    ┌─────────────────┐                   │
│                    │  Async Commands │                   │
│                    │  - loadDir()    │                   │
│                    │  - ExecProcess()│                   │
│                    └─────────────────┘                   │
└──────────────────────────────────────────────────────────┘
```

### 3.1 Model

```go
type model struct {
    cwd       string            // Current working directory (absolute)
    entries   []entry           // Sorted list of directory entries
    cursor    int               // Index of selected entry
    width     int               // Terminal width (from WindowSizeMsg)
    height    int               // Terminal height (from WindowSizeMsg)
    preview   string            // Pre-computed preview string for right pane
    gitStatus map[string]string // filename → git status code ("M", "?", …)
    err       error             // Last error to display in-pane
}
```

The `entry` struct wraps `os.FileInfo` alongside the name and a boolean `isDir` flag, providing everything needed for both rendering and navigation without additional stat calls.

### 3.2 Messages (Msg)

| Message | Source | Purpose |
|---|---|---|
| `tea.WindowSizeMsg` | Bubble Tea runtime | Captures terminal dimensions for layout |
| `tea.KeyPressMsg` | Keyboard | All navigation and quit signals |
| `dirLoadedMsg` | `loadDir` cmd | Delivers fresh entry list + git map |
| `editorFinishedMsg` | `tea.ExecProcess` callback | Signals vim has exited; triggers reload |

### 3.3 Commands (Cmd)

#### `loadDir(path string) tea.Cmd`
Runs asynchronously. Reads the directory with `os.ReadDir`, sorts entries (directories first, then alphabetical), and calls `getGitStatus` in the same goroutine. Returns `dirLoadedMsg`.

#### `tea.ExecProcess(cmd, callback)`
Suspends the TUI, forks `vim <file>`, and resumes on exit. The callback wraps the error in `editorFinishedMsg`.

### 3.4 View

`View()` is a pure function of `model` that produces a `tea.View`. It:

1. Computes `leftWidth` (40 % of usable width) and `rightWidth` (60 %).
2. Renders the **left pane**: path header → optional error → visible entry rows (with scroll offset, git badges, directory indicators).
3. Renders the **right pane**: pre-computed `m.preview` string (file content or dir listing).
4. Joins panes horizontally with `lipgloss.JoinHorizontal`.
5. Appends a **status bar** with item count, position, and key hints.
6. Sets `AltScreen = true` so the TUI uses the secondary terminal buffer.

---

## 4. Layout

```
┌─────────────────────────┬────────────────────────────────┐
│  ~/projects/scout       │  📄 File: main.go              │
│  ──────────────────     │  ──────────────────────────    │
│  [M] main.go            │  Size:     16.0 KB             │
│  [?] SPEC.md            │  Modified: 2026-04-18 17:00    │
│  ▶  go.mod              │  Mode:     -rw-r--r--          │
│     go.sum              │  ──────────────────────────    │
│     README.md           │    1 │ package main            │
│     scout               │    2 │                         │
│                         │    3 │ import (                │
│                         │    …                           │
└─────────────────────────┴────────────────────────────────┘
 5 items  1/5  │  q:quit  j/k:navigate  h:up  enter:open  g/G:top/bottom
```

- **Left pane** — 40 % of terminal width, rounded border, purple accent.
- **Right pane** — 60 % of terminal width, rounded border, same accent.
- **Status bar** — single line below the panes; dim colour.

---

## 5. Key Bindings

| Key | Action |
|---|---|
| `j` / `↓` | Move cursor down |
| `k` / `↑` | Move cursor up |
| `h` / `←` | Go to parent directory |
| `enter` | Enter directory or open file in vim |
| `g` | Jump to top of list |
| `G` / `shift+G` | Jump to bottom of list |
| `q` / `ctrl+c` | Quit |

---

## 6. Git Status Integration

`getGitStatus(dir)` runs:
```
git status --porcelain
```
in `dir` and produces a `map[string]string`:

- Output line format: `XY filename` (2-char status + space + path)
- `??` → badge `[?]` (untracked, green)
- Any other non-space XY → badge `[M]` (modified/staged/etc., orange)
- Nested paths (e.g. `subdir/file.go`) attribute the change to the top-level entry (`subdir`)
- Renamed paths (`R  old -> new`) use the new name

If `git` is unavailable or the directory is not a repo, the map is empty and no badges are shown.

---

## 7. Preview Logic

| Selected entry | Preview content |
|---|---|
| **Directory** | Icon, modified time, mode, child count, list of up to 20 children |
| **Text file** | Icon, size, modified time, mode, first 40 lines with line numbers |
| **Binary file** | Icon, metadata, `(binary file – no preview)` message |

Binary detection: any null byte (`0x00`) in the first 4 KB marks the file as binary.

Preview is regenerated eagerly whenever the cursor moves, a directory is loaded, or the window is resized. It is stored in `model.preview` as a pre-rendered string to keep `View()` allocation-light.

---

## 8. File Structure

```
scout/
├── main.go       # Entire application — Model, Update, View + helpers
├── go.mod        # Module: github.com/mirageglobe/scout
├── go.sum        # Dependency lock
├── README.md     # Project overview
├── SPEC.md       # This document
└── scout         # Compiled binary (gitignored)
```

---

## 9. Build & Run

```bash
# Build
go build -o scout .

# Run in current directory
./scout

# Run in a specific directory
cd /some/path && /path/to/scout
```

### Prerequisites
- Go 1.22+ (module `go 1.26.2` toolchain)
- `vim` on `$PATH` for file opening
- `git` on `$PATH` for status badges (optional)

---

## 10. Design Decisions

| Decision | Rationale |
|---|---|
| Single `main.go` | Keeps the project approachable and easy to audit in one read |
| Pre-computed `preview` string | Avoids re-allocating on every `View()` call; only recomputes on state changes |
| `AltScreen = true` | Uses the secondary terminal buffer so the shell history is not polluted |
| `tea.ExecProcess` for vim | The idiomatic Bubble Tea way to suspend the TUI, hand off stdin/stdout, and resume cleanly |
| No `bubbles/list` component | Gives full control over git badge rendering, scrolling, and padding behaviour |
| Directories first sort | Standard filesystem browser convention; reduces cognitive load |
| 4 KB preview cap | Prevents large files from blocking the UI thread during preview generation |
