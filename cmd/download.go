package cmd

import (
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
		m := &downloader.DownloadManager{
			Job:      make(chan downloader.DownloadJob),
			Progress: make(chan downloader.Progress),
		}
		done := make(chan bool)
		totalWorkers := 3
		if filename != "" && len(args) > 1 {
			log.Fatal("Error: Flag output is only available for single download")
		}

		m.StartWorker(totalWorkers)

		for _, url := range args {
			m.StartDownload(url, filename)
		}
		close(m.Job)

		go func() {
			nextIndex := 0
			lineMap := make(map[string]int)
			for p := range m.Progress {
				if _, ok := lineMap[p.Filename]; !ok {
					lineMap[p.Filename] = nextIndex
					nextIndex++
					fmt.Println()
				}

				dist := nextIndex - lineMap[p.Filename]
				fmt.Printf("\033[%dA", dist)
				fmt.Print("\r")
				fmt.Printf(
					"\033[K%-30s | %.2f%% | %-7s/%-7s |  %s/s | ETA %s",
					downloader.Truncate(p.Filename, 30),
					p.Percentage,
					downloader.FormatBytes(p.CurrentSize),
					downloader.FormatBytes(p.TotalSize),
					downloader.FormatBytes(int64(p.Speed)),
					downloader.FormatTime(p.ETA),
				)
				fmt.Printf("\033[%dB", dist)
			}

			fmt.Println()
			done <- true
		}()

		m.Wg.Wait()

		close(m.Progress)
		<-done

		return nil
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
