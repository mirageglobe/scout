# Scout - Agent Instructions

This document is intended for AI coding assistants working in the `scout` directory.

## Commands
- **Build**: `make build` or `go build -o scout cmd/scout/main.go`
- **Run**: `make run` or `./scout`
- **Format**: `make fmt` or `go fmt ./...`

## Project Context
- **Description**: A dual-pane terminal UI file manager and previewer.
- **Language**: Go
- **Core Libraries**: `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`
- **Architecture**: Modular structure (`cmd/`, `internal/`) using Model-View-Update (MVU). See `SPEC.md` for full architectural details.

## Coding Conventions
- Do not import external "UI components" or "bubbles" packages without explicit user permission. The current file list and preview panes are custom-built for explicit layout and scrolling control.
- Ensure any file/directory modifications properly update the git status badge logic.
- Process execution (like launching `vim`) must use `tea.ExecProcess`.
- Run `go build` after making changes to verify compilation. Use standard `go fmt` rules.
- **Git Commits**: Do not commit code with co-authors.
- **Aesthetics**: Prefer clean ASCII and Unicode characters over emoji for icons and UI elements to maintain a professional, high-contrast look.
- **Diagrams**: Structure all architectural or logic diagrams using clean ASCII or Unicode characters.
- **Tables**: Ensure all Markdown tables are highly readable by maintaining perfect column alignment and spacing.
- **Documentation**: Prioritise code comments where possible; they are essential for readability and context for both humans and AI.
