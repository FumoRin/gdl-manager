package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteFiles bool

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a download from the manager",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		m := app.Manager
		state, getErr := m.GetDownload(id)
		if getErr != nil {
			return fmt.Errorf("error getting download detail: %w", getErr)
		}
		if state == nil {
			return fmt.Errorf("download with ID %s not found", id)
		}

		err := m.DeleteDownload(id, deleteFiles)
		if err != nil {
			return fmt.Errorf("error deleting download: %w", err)
		}

		if deleteFiles {
			fmt.Printf("Successfully deleting download and its files: %s\n", state.Filename)
		} else {
			fmt.Printf("Successfully deleting download entry: %s\n", state.Filename)
		}
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(
		&deleteFiles,
		"files",
		"f",
		false,
		"Also delete the downloaded files from disk",
	)
	rootCmd.AddCommand(deleteCmd)
}
