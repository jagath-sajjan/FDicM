package main

import (
	"fdicm/internal/ui"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var version = "v1.0.0-dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Printf("FDicM Terminal Interface Engine %s\n", version)
			os.Exit(0)
		case "--help", "-h", "help":
			fmt.Println("FDicM - High-Fidelity TUI Split-Pane Dictionary Client")
			fmt.Println("\nUsage:\n  fdicm             Launch the interface dashboard")
			fmt.Println("  fdicm --version   Print engine build metadata details")
			os.Exit(0)
		}
	}

	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, the TUI has crashed: %v\n", err)
		os.Exit(1)
	}
}
