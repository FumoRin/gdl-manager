package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

func Download(url string, opts DownloadOptions, wg *sync.WaitGroup) (err error) {
	defer wg.Done()
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

	filename := resolveFilename(url, opts, resp)

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

	pw := &ProgressWriter{
		Filename:     filename,
		Total:        resp.ContentLength,
		Destination:  file,
		StartTime:    time.Now(),
		ProgressChan: opts.Progress,
	}

	_, err = io.Copy(pw, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
