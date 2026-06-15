package cmd

import (
	"fmt"
	"log"

	"github.com/fumorin/gdl-manager/internal/downloader"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all of the download job",
	RunE: func(cmd *cobra.Command, args []string) error {
		m := downloader.NewDownloadManager()
		if err := m.LoadState("download.json"); err != nil {
			log.Fatalln("No download state found")
		}

		fmt.Printf("%-5s | %-30s | %s\n", "ID", "Filename", "Status")
		for id, state := range m.State {
			fmt.Printf("%-5s | %-30s | %s\n", id, state.Filename, state.Status.String())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
