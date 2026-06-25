package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var renameDownload = &cobra.Command{
	Use: "rename [id] [filename]",
	Short: "Rename existing download filename",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		newFilename := args[1]
		m := app.Manager
		err := m.RenameDownload(id, newFilename)
		if err != nil {
			return fmt.Errorf("error renaming download: %w", err)
		}

		fmt.Printf("Successfully renaming download: %s\n", newFilename)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(renameDownload)
}
