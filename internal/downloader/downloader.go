package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"
)

func Download(url string, opts DownloadOptions) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			err = closeErr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"server returned %d",
			resp.StatusCode,
		)
	}

	filename := opts.Filename
	if filename == "" {
		filename = path.Base(url)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			err = closeErr
		}
	}()

	size := formatBytes(resp.ContentLength)

	fmt.Printf("Downloading %s\n", filename)
	fmt.Printf("Size: %s bytes\n", size)

	pw := &ProgressWriter{
		Total:       resp.ContentLength,
		Destination: file,
		StartTime: time.Now(),
	}

	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
				case <-ticker.C:
					percentage :=
						(float64(pw.Current)/float64(pw.Total)) * 100

					fmt.Printf(
						"\r\033[K%.2f%% | %s/%s | %s/s | ETA %s",
						percentage,
						formatBytes(pw.Current),
						formatBytes(pw.Total),
						formatBytes(int64(pw.Speed())),
						pw.FormatTime(pw.Eta()),
					)
				case <-done:
					return
			}
		}
	}()

	_, err = io.Copy(pw, resp.Body)
	if err != nil {
		return err
	}

	close(done)

	fmt.Println()
	fmt.Println("Download Complete")

	return nil
}
