package cmd

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var filename string

var downloadCmd = &cobra.Command{
	Use:   "download [url]",
	Short: "Download a file from the given url",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m := app.Manager

		if filename != "" && len(args) > 1 {
			return fmt.Errorf("error: Flag output is only available for single download")
		}

		for _, url := range args {
			ID := uuid.NewString()
			m.StartDownload(ID, url, filename)
		}
		m.Close()

		return RunDownloadSession(m)
	},
}

func init() {
	downloadCmd.Flags().StringVarP(
		&filename,
		"output",
		"o",
		"",
		"The output filename provided by user (Only for a single download)",
	)
	rootCmd.AddCommand(downloadCmd)
}
