# Scout - Agent Instructions

This document is intended for AI coding assistants working in the `scout` directory.

## Commands
- **Build**: `go build -o scout .`
- **Run**: `./scout`
- **Format**: `go fmt ./...`

## Project Context
- **Description**: A dual-pane terminal UI file manager and previewer.
- **Language**: Go
- **Core Libraries**: `charm.land/bubbletea/v2`, `charm.land/lipgloss/v2`
- **Architecture**: Single-file (`main.go`) Model-View-Update (MVU). See `SPEC.md` for full architectural details.

## Coding Conventions
- Do not import external "UI components" or "bubbles" packages without explicit user permission. The current file list and preview panes are custom-built for explicit layout and scrolling control.
- Ensure any file/directory modifications properly update the git status badge logic.
- Process execution (like launching `vim`) must use `tea.ExecProcess`.
- Run `go build` after making changes to verify compilation. Use standard `go fmt` rules.
- **Git Commits**: Do not commit code with co-authors.
