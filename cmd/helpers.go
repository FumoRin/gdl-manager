package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fumorin/gdl-manager/internal/downloader"
)

func RunDownloadSession(m *downloader.DownloadManager) error {
	progressDone := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Print Download Progress
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
		close(progressDone)
	}()

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal %v. Shutting down gracefully...\n", sig)
		m.Cancel()
	}()

	m.Wg.Wait()
	close(m.Progress)
	<-progressDone

	return nil
}
