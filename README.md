# Scout

*when you need a rapid intelligence overview of your environment, you call in a Scout.*

![Scout Demo](demo.gif)

Scout is a fast, elegant, terminal-native file explorer designed for immediate situational awareness. It combines a high-performance dual-pane layout with real-time Git integration and rich previews to help you navigate your codebase with speed and precision.

---

### ◆ Key Features

- **▸ Navigation**: fully keyboard-driven with instant directory entry, parent-navigation, and top/bottom jumps.
- **▸ Rich Previews**: real-time file previews with Chroma syntax highlighting, directory metadata, and intelligent binary detection.
- **▸ Git Integration**: integrated git status badges (`M`, `+`, `?`, `!`) and branch name in the status bar.
- **▸ Time-Aware Themes**: seven color themes (Midnight, Dawn, Classic Amber, Electric Cyan, Safety Orange, Evening, Mono) auto-selected by time of day, manually cycled with `t`.
- **▸ Help Overlay**: full keybinding and symbol reference available at any time with `?`.
- **▸ System Stats**: live CPU usage, memory consumption, directory size, and clock in the header bar.
- **▸ Editor Handoff**: seamlessly launch into `vim` with a single keystroke; TUI suspends and resumes cleanly.
- **▸ Collapsible Pane**: compress the file list to 8 chars with `tab` for maximum preview space.

---

### ◆ Keybindings

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

### ◆ Getting Started

Ensure you have [Go](https://go.dev/) installed, then:

```bash
# clone and build
git clone https://github.com/mirageglobe/scout.git
cd scout
make build

# run Scout
./scout
```

*to regenerate the demo GIF, ensure you have [vhs](https://github.com/charmbracelet/vhs) installed and run `make demo`.*

---

## Releasing

Scout uses [goreleaser](https://goreleaser.com) to build cross-platform binaries, publish a GitHub release, and auto-update the [homebrew-tap](https://github.com/mirageglobe/homebrew-tap) formula.

### prerequisites

- [`goreleaser`](https://goreleaser.com/install/) installed (`brew install goreleaser`)
- two GitHub personal access tokens (classic or fine-grained, with repo write scope):
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
git tag v0.2.0
git push origin v0.2.0

# 4. release (builds binaries, publishes GitHub release, updates homebrew formula)
make release
```

to test the release pipeline locally without publishing:

```bash
make release-dry
```

---

## Architecture

Scout follows standard Go conventions for scalable, modular CLIs:

```text
scout/
├── cmd/
│   └── scout/
│       └── main.go             # entry point
│
├── internal/
│   ├── filesystem/             # file ops, stats, and utilities
│   ├── git/
│   │   └── git.go              # subprocess logic for fetching git statuses
│   └── ui/
│       ├── model.go            # Bubble Tea model definition, Msg structs
│       ├── update.go           # Init() and Update() functions
│       ├── view.go             # View() function and lipgloss styles
│       └── preview.go          # preview text building logic
│
├── go.mod
├── go.sum
├── AGENT.md
├── CLAUDE.md
├── README.md
└── SPEC.md
```

---

## Roadmap

### Near Term
- [x] ls all files in current directory
- [x] syntax highlighting
- [x] time-aware color themes
- [x] help overlay
- [x] system stats in header (CPU, memory, clock)
- [x] git branch display in status bar
- [x] collapsible file list pane
- [ ] respect `$EDITOR` environment variable for editor handoff

### Future Ideas
- [ ] preview images
- [ ] fuzzy file search

---

## Known Issues

- **TUI Viewport Overflow**: in some environments (notably `tmux`), the preview pane can occasionally extend beyond the bottom of the screen when viewing long files, causing the status bar to disappear. this is likely due to complex ANSI width calculations or terminal height reporting discrepancies.
