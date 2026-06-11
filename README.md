# Scout

*when you need a rapid intelligence overview of your environment, you call in a Scout.*

[![License](https://img.shields.io/badge/license-BUSL--1.1-blue.svg)](LICENSE.md)

![Scout Demo](demo.gif)

Scout is a fast, elegant, terminal-native file explorer designed for immediate situational awareness. It combines a high-performance dual-pane layout with real-time Git integration and rich previews to help you navigate your codebase with speed and precision.

---

## Key Features

- **Navigation**: fully keyboard-driven with instant directory entry, parent-navigation, and top/bottom jumps.
- **Rich Previews**: real-time file previews with Chroma syntax highlighting, directory metadata, and intelligent binary detection.
- **Git Integration**: integrated git status badges (`M`, `+`, `?`, `!`) and branch name in the status bar.
- **Time-Aware Themes**: nine color themes auto-selected by time of day, manually cycled with `t`; `T` toggles dark / light mode.
- **Help Overlay**: full keybinding and symbol reference available at any time with `?`.
- **System Stats**: live CPU usage, memory consumption, directory size, and clock in the header bar.
- **Editor Handoff**: seamlessly launch into `vim` with a single keystroke; TUI suspends and resumes cleanly.
- **Resizable Explorer**: cycle the file list pane through sliver, narrow, and wide widths with `tab` to trade list space for preview space.
- **Search & Clipboard**: search the active pane with `/` (`n`/`N` to step matches); copy the selected path with `y`.

---

## Installation

**via homebrew (recommended):**

```bash
brew tap mirageglobe/tap
brew install mirageglobe/tap/scout
brew upgrade mirageglobe/tap/scout
```

**via install script (no homebrew):**

```bash
curl -fsSL https://raw.githubusercontent.com/mirageglobe/scout/main/install.sh | sh
```

detects your os/arch, downloads the matching release binary, verifies its checksum, and installs to `~/.local/bin`. re-run any time to upgrade to the latest release. set `SCOUT_VERSION` to pin a version or `SCOUT_BIN_DIR` to choose the install directory.

**from source:**

```bash
git clone https://github.com/mirageglobe/scout.git
cd scout
make build
./scout
```

---

## Configuration

Scout stores your theme preferences in `~/.config/scout/config`. This file is automatically created and updated when you cycle through themes using `t`.

---

## Keybindings

| key            | action                                             |
| :------------- | :------------------------------------------------- |
| `↓` / `↑`      | move cursor down / up                              |
| `←` / `⌫`      | nav to parent directory (or back from preview)     |
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

> for architecture, design decisions, roadmap, and release process see [SPEC.md](SPEC.md).
