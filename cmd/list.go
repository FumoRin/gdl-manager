package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all of the download job",
	RunE: func(cmd *cobra.Command, args []string) error {
		m := app.Manager
		states, err := m.Repo.GetAllDownloads()
		if err != nil {
			log.Fatalln("Failed to retrieve downloads:", err)
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
