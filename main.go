package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse command line flags
	var filePath string
	flag.StringVar(&filePath, "file", "", "Path to todo list file")
	flag.StringVar(&filePath, "f", "", "Path to todo list file (shorthand)")
	flag.Parse()

	// Resolve storage path
	storagePath, err := ResolveStoragePath(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Create storage
	storage := NewStorage(storagePath)

	// Load todo list
	list, err := storage.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading todos: %v\n", err)
		fmt.Fprintf(os.Stderr, "Starting with empty list...\n")
		list = NewTodoList()
	}

	// Create and run the TUI
	model := NewModel(list, storage)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
