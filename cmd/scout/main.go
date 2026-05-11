package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"github.com/mirageglobe/scout/internal/ui"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("scout v%s\n", ui.Version)
			return
		case "--help", "-h":
			fmt.Println("usage: scout [--version] [--help] [path]")
			fmt.Println()
			fmt.Println("  path    directory to open (default: current directory)")
			return
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "scout: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		target := os.Args[1]
		if !filepath.IsAbs(target) {
			target = filepath.Join(cwd, target)
		}
		info, err := os.Stat(target)
		if err != nil || !info.IsDir() {
			fmt.Fprintf(os.Stderr, "scout: not a directory: %s\n", os.Args[1])
			os.Exit(1)
		}
		cwd = target
	}

	p := tea.NewProgram(ui.NewModel(cwd))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "scout: %v\n", err)
		os.Exit(1)
	}
}
