package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/fumorin/gdl-manager/internal/downloader"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func RunDownloadSession(m *downloader.DownloadManager) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	p := mpb.New()

	var bars sync.Map
	var progressMap sync.Map

	progressDone := make(chan struct{})

	// Print Download Progress
	go func() {
		for prog := range m.ProgressChan() {
			// Store a copy of the progress to avoid race conditions
			pCopy := prog
			progressMap.Store(pCopy.Filename, &pCopy)

			if val, ok := bars.Load(pCopy.Filename); ok {
				bar := val.(*mpb.Bar)
				bar.SetCurrent(pCopy.CurrentSize)
			} else {
				filename := pCopy.Filename
				bar := p.AddBar(pCopy.TotalSize,
					mpb.PrependDecorators(
						decor.Name(downloader.Truncate(filename, 20), decor.WC{W: 20, C: decor.DindentRight}),
						decor.Percentage(decor.WC{W: 6}),
					),
					mpb.AppendDecorators(
						decor.Any(func(st decor.Statistics) string {
							if v, ok := progressMap.Load(filename); ok {
								currProg := v.(*downloader.Progress)
								return fmt.Sprintf(" %s/%s",
									downloader.FormatBytes(currProg.CurrentSize),
									downloader.FormatBytes(currProg.TotalSize),
								)
							}
							return ""
						}, decor.WC{W: 25, C: decor.DindentRight}),
						decor.Any(func(st decor.Statistics) string {
							if v, ok := progressMap.Load(filename); ok {
								currProg := v.(*downloader.Progress)
								return fmt.Sprintf("  %s/s", downloader.FormatBytes(int64(currProg.Speed)))
							}
							return ""
						}, decor.WC{W: 15, C: decor.DindentRight}),
						decor.Any(func(st decor.Statistics) string {
							if v, ok := progressMap.Load(filename); ok {
								currProg := v.(*downloader.Progress)
								return fmt.Sprintf("  ETA %s", downloader.FormatTime(currProg.ETA))
							}
							return ""
						}, decor.WC{W: 30, C: decor.DindentRight}),
					),
				)
				bars.Store(filename, bar)
				bar.SetCurrent(pCopy.CurrentSize)
			}
		}

		p.Wait()
		close(progressDone)
	}()

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal %v. Shutting down gracefully...\n", sig)
		m.Shutdown()
		p.Shutdown()
	}()

	m.Wait()

	<-progressDone

	return nil
}
