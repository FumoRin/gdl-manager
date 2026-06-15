package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fumorin/gdl-manager/internal/downloader"
)

func RunDownloadSession(m *downloader.DownloadManager) error {
	done := make(chan bool)
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
		done <- true
	}()

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal %v. Shutting down gracefully...", sig)
		m.Mu.Lock()
		for id, cancel := range m.Cancellations {
			fmt.Printf("Cancelling download %s ...", id)
			cancel()
		}
		m.Mu.Unlock()
	}()
	
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				if saveErr := m.SaveState("download.json"); saveErr != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	m.Wg.Wait()
	close(m.Progress)
	<-done
	return nil
}
