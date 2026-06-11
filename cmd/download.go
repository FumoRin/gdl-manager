package cmd

import (
	"github.com/fumorin/gdl-manager/internal/downloader"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use: "download [url]",
	Short: "Download a file from the given url",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]
		
		opts := downloader.DownloadOptions{
			
		}

		return downloader.Download(url, opts)
	},	
}

func init() {
	rootCmd.AddCommand(downloadCmd)
}
