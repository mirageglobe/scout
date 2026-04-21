# Scout

*when you need a rapid intelligence overview of your environment, you call in a Scout.*

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

![Scout Demo](demo.gif)

Scout is a fast, elegant, terminal-native file explorer designed for immediate situational awareness. It combines a high-performance dual-pane layout with real-time Git integration and rich previews to help you navigate your codebase with speed and precision.

---

## Key Features

- **▸ Navigation**: fully keyboard-driven with instant directory entry, parent-navigation, and top/bottom jumps.
- **▸ Rich Previews**: real-time file previews with Chroma syntax highlighting, directory metadata, and intelligent binary detection.
- **▸ Git Integration**: integrated git status badges (`M`, `+`, `?`, `!`) and branch name in the status bar.
- **▸ Time-Aware Themes**: seven color themes auto-selected by time of day, manually cycled with `t`.
- **▸ Help Overlay**: full keybinding and symbol reference available at any time with `?`.
- **▸ System Stats**: live CPU usage, memory consumption, directory size, and clock in the header bar.
- **▸ Editor Handoff**: seamlessly launch into `vim` with a single keystroke; TUI suspends and resumes cleanly.
- **▸ Collapsible Pane**: compress the file list to 8 chars with `tab` for maximum preview space.

---

## Keybindings

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

## Getting Started

**via homebrew (recommended):**

```bash
brew tap mirageglobe/tap
brew install mirageglobe/tap/scout
```

**from source:**

```bash
git clone https://github.com/mirageglobe/scout.git
cd scout
make build
./scout
```

---

## Roadmap

### near term
- [x] ls all files in current directory
- [x] syntax highlighting
- [x] time-aware color themes
- [x] help overlay
- [x] system stats in header (CPU, memory, clock)
- [x] git branch display in status bar
- [x] collapsible file list pane
- [ ] respect `$EDITOR` environment variable for editor handoff

### future ideas
- [ ] preview images
- [ ] fuzzy file search

---

## Known Issues

- **TUI Viewport Overflow**: in some environments (notably `tmux`), the preview pane can occasionally extend beyond the bottom of the screen when viewing long files, causing the status bar to disappear. this is likely due to complex ANSI width calculations or terminal height reporting discrepancies.
