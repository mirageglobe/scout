# Scout

*When you need a rapid intelligence overview of your environment, you call in a Scout.*

![Scout Demo](demo.gif)

Scout is a fast, elegant, terminal-native file explorer designed for immediate situational awareness. It combines a high-performance dual-pane layout with real-time Git integration and rich previews to help you navigate your codebase with speed and precision.

### ◆ Key Features
- **▸ Navigation**: Fully keyboard-driven (`j`/`k`/`h`/`l`) with instant directory entry and parent-navigation.
- **▸ Rich Previews**: Real-time file previews with **Chroma syntax highlighting**, directory metadata, and intelligent binary detection.
- **▸ Git Integration**: Integrated git status indicators (`●`, `○`) show you exactly which files have changed.
- **▸ Pane Focus**: Effortlessly switch focus between the file list and the preview pane for scrolling.
- **▸ Editor Handoff**: Seamlessly launch into `vim` with a single keystroke.
- **▸ Aesthetics**: Built with the [Charm](https://charm.sh) stack, featuring a polished UI and minimalist symbols (`▸`, `•`).

---

### ◆ Getting Started
Ensure you have [Go](https://go.dev/) installed, then:

```bash
# Clone and build
git clone https://github.com/mirageglobe/scout.git
cd scout
make build

# Run Scout
./scout
```

*To regenerate the demo GIF, ensure you have [vhs](https://github.com/charmbracelet/vhs) installed and run `make demo`.*

---

## Target Go Architecture
While initially prototyped as a standalone `main.go`, the target reference structure for this project follows standard Go conventions for scalable, modular CLIs:

```text
scout/
├── cmd/
│   └── scout/
│       └── main.go       (Entry point: executes the program)
│
├── internal/
│   ├── git/
│   │   └── git.go        (Subprocess logic for fetching git statuses)
│   ├── ui/
│   │   ├── model.go      (Bubble Tea model definition, Msg structs)
│   │   ├── update.go     (Init() and Update() functions)
│   │   ├── view.go       (View() function and lipgloss styles)
│   │   └── preview.go    (Preview text building logic)
│
├── go.mod
├── go.sum
├── AGENT.md
├── CLAUDE.md (symlink)
├── README.md
└── SPEC.md
```

## Roadmap

### Near Term
- [x] ls all files in current directory
- [ ] allow vim as editor by reading editor in env var
- [ ] Implement target Go folder refactor (`cmd/` + `internal/`)
- [x] syntax highlighting

### Future Ideas
- [ ] preview images

## Known Issues
- **TUI Viewport Overflow**: In some environments (notably `tmux`), the preview pane can occasionally extend beyond the bottom of the screen when viewing long files, causing the status bar to disappear. This is likely due to complex ANSI/Emoji width calculations or terminal height reporting discrepancies.