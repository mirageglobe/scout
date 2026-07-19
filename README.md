# Scout

*when you need a rapid intelligence overview of your environment, you call in a Scout.*

[![License](https://img.shields.io/badge/license-BUSL--1.1-blue.svg)](LICENSE.md)

![Scout Demo](demo.gif)

Scout is a fast, keyboard-driven file explorer for the terminal. Files on the left, a live syntax-highlighted preview on the right, with Git status, diffs, and system stats always in view, so you can read and move through a codebase at speed without opening an editor.

Built for developers who live in the terminal: it stays responsive on large files and directories, shows what changed at a glance, and hands off to your `$EDITOR` the moment you want to edit.

---

## Key Features

- **Navigation**: fully keyboard-driven with instant directory entry, parent-navigation, and top/bottom jumps.
- **Live Previews**: syntax-highlighted file previews (Chroma), directory metadata, and binary detection; large files highlight in the background so navigation never stalls.
- **Git Integration**: integrated git status badges (`M`, `+`, `?`, `!`) and branch name in the status bar; `d` cycles the preview between the file, its `git diff`, and `git log`.
- **Time-Aware Themes**: nine color themes auto-selected by time of day, manually cycled with `t`; `T` toggles dark / light mode.
- **Help Overlay**: full keybinding and symbol reference available at any time with `?`.
- **System Stats**: live CPU usage, memory consumption, directory size, and clock in the header bar.
- **Editor Handoff**: open the selected file in your `$EDITOR` with a single keystroke (`e`); the TUI suspends and resumes cleanly.
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

Set `SCOUT_UNICODE_SAFE=1` to swap the UI marker glyphs (`▸`, `⎇`, `›`, arrows, and friends) for narrow-safe ASCII equivalents. This helps in terminals that render ambiguous-width characters as two cells (for example with `RUNEWIDTH_EASTASIAN=1`), which would otherwise misalign the columns.

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
| `d`            | cycle git preview (file / diff / log)              |
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
