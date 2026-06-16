package cmd

import (
	"crypto/sha256"
	"fmt"
	"log"

	"github.com/fumorin/gdl-manager/internal/downloader"
	"github.com/spf13/cobra"
)

var filename string

var downloadCmd = &cobra.Command{
	Use:   "download [url]",
	Short: "Download a file from the given url",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m := downloader.NewDownloadManager()
		totalWorkers := 3

		if filename != "" && len(args) > 1 {
			log.Fatal("Error: Flag output is only available for single download")
		}
		
		m.Wg.Add(len(args))

		m.StartWorker(totalWorkers)

		go func ()  {
			for _, url := range args {
				hash := sha256.Sum256([]byte(url))
				ID := fmt.Sprintf("%x", hash)
				m.StartDownload(ID, url, filename)
			}
			close(m.Job)	
		}()
		
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
