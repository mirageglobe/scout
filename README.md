# Scout

*When you need a rapid intelligence overview of your environment, you call in a Scout.*

Scout was born out of a desire for a fast, elegant, terminal-native file explorer that doesn't just list files, but gives you immediate situational awareness. Designed for power users, it uses a dual-pane layout: the left panel presents a structured list of files/directories interwoven with live Git status markers, while the right panel provides a robust real-time preview revealing file metadata, text contents, or inner directory structures.

It is lightweight, completely keyboard-driven (`j`/`k`/`h`/`enter`), and seamlessly hands off to your editor (`vim`) when you need to act on your reconnaissance. Built purely in Go using the [Charm](https://charm.sh) stack (Bubble Tea & Lip Gloss), Scout combines rich visual aesthetics with strict UNIX philosophy.

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

### Ideas and Issues
- preview images
- syntax highlighting