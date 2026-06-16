package downloader

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func Download(url string, opts DownloadOptions, wg *sync.WaitGroup, ctx context.Context) (err error) {
	defer wg.Done()

	var currentSize int64
	var file *os.File
	var errFile error

	client := &http.Client{}
	headReq, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return err
	}

	headResp, err := client.Do(headReq)
	if err != nil {
		return err
	}
	filename := resolveFilename(url, opts, headResp)
	closeErr := headResp.Body.Close()
	if closeErr != nil {
		return closeErr
	}

	bodyGet, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	if currentSize > 0 {
		bodyGet.Header.Set("Range", fmt.Sprintf("bytes=%d-", currentSize))
	}

	bodyResp, err := client.Do(bodyGet)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := bodyResp.Body.Close()
		if closeErr != nil {
			err = closeErr
		}
	}()

	switch bodyResp.StatusCode {
	case http.StatusPartialContent:
		file, errFile = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
		if errFile != nil {
			return errFile
		}

	case http.StatusOK:
		file, errFile = os.Create(filename)
		if errFile != nil {
			return errFile
		}

	default:
		log.Fatalf("Unexpected status code: %d\n", bodyResp.StatusCode)
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			err = closeErr
		}
	}()

	buf := make([]byte, 32*1024)
	pw := &ProgressWriter{
		Filename:     filename,
		Total:        bodyResp.ContentLength,
		Destination:  file,
		StartTime:    time.Now(),
		ProgressChan: opts.Progress,
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			n, err := bodyResp.Body.Read(buf)

			if err == io.EOF {
				return nil
			} else if err != nil {
				return err
			}

			_, dataErr := pw.Write(buf[:n])
			if dataErr != nil {
				return dataErr
			}
		}
	}
}
