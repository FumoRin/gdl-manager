package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all of the download job",
	RunE: func(cmd *cobra.Command, args []string) error {
		m := app.Manager
		states, err := m.GetAllDownload()
		if err != nil {
			return fmt.Errorf("failed to retrieve downloads: %v", err)
		}

		fmt.Printf("%-64s | %-30s | %s\n", "ID", "Filename", "Status")
		for _, state := range states {
			fmt.Printf("%-64s | %-30s | %s\n", state.ID, state.Filename, state.Status.String())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
