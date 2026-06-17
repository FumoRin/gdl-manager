package cmd

import (
	"fmt"
	"log"

	"github.com/fumorin/gdl-manager/internal/downloader"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume [id]",
	Short: "Resume a paused download",
	RunE: func(cmd *cobra.Command, args []string) error {
		totalWorkers := 3
		m := downloader.NewDownloadManager()
		if err := m.LoadState("download.json"); err != nil {
			log.Fatalln("No download state found")
		}

		m.StartWorker(totalWorkers)

		if len(args) > 0 {
			// If args provided
			targetID := args[0]

			state, exist := m.State[targetID]
			if !exist {
				log.Fatalf("Download with provided ID not found: %s\n", targetID)
			}

			if state.Status != downloader.StateCompleted {
				fmt.Printf("Resuming specific download: %s\n", state.Filename)
				m.Wg.Add(1)
				m.StartDownload(targetID, state.URL, state.Filename)
			} else {
				fmt.Printf("Download %s is already completed", state.Filename)
			}
		} else {
			// If no Args provided
			var toResume []string
			for id, state := range m.State {
				if state.Status != downloader.StateCompleted {
					toResume = append(toResume, id)
				}

				m.Wg.Add(len(toResume))

				for _, id := range toResume {
					state := m.State[id]
					m.StartDownload(id, state.URL, state.Filename)
				}
			}
		}

		close(m.Job)
		return RunDownloadSession(m)
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)
}
