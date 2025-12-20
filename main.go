package main

import (
	"fmt"
	"log"
	"os"

	"github.com/abhimanyu003/pttr/internal/common"
	"github.com/abhimanyu003/pttr/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	startWithProcesses bool
	startWithIPs       bool
	version            = "dev"
)

var rootCmd = &cobra.Command{
	Use:     "portmanager",
	Version: version,
	Short:   "Cross-platform port and process management tool",
	Long:    `Port and Process Manager is a cross-platform TUI application for managing ports and processes`,
	Run: func(cmd *cobra.Command, args []string) {
		// Start the main application
		initialMode := common.PortsMode
		if startWithProcesses {
			initialMode = common.ProcessesMode
		} else if startWithIPs {
			initialMode = common.IPsMode
		}

		model := ui.NewModel(initialMode)
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.Flags().BoolVarP(&startWithProcesses, "processes", "p", false, "Start with processes view instead of ports")
	rootCmd.Flags().BoolVarP(&startWithIPs, "ips", "i", false, "Start with network interfaces view")
	rootCmd.CompletionOptions.DisableDefaultCmd = false
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
