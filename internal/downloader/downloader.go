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

func Download(url string, opts DownloadOptions, wg *sync.WaitGroup, ctx context.Context) (totalSize int64, filename string, err error) {
	var currentSize int64
	var file *os.File
	var errFile error

	client := &http.Client{}
	headReq, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return
	}

	headResp, err := client.Do(headReq)
	if err != nil {
		return
	}
	filename = resolveFilename(url, opts, headResp)
	if info, err := os.Stat(filename); err == nil {
		currentSize = info.Size()
	}

	closeErr := headResp.Body.Close()
	if closeErr != nil {
		err = closeErr
		return
	}

	bodyGet, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	if currentSize > 0 {
		bodyGet.Header.Set("Range", fmt.Sprintf("bytes=%d-", currentSize))
	}

	bodyResp, err := client.Do(bodyGet)
	if err != nil {
		return
	}

	defer func() {
		closeErr := bodyResp.Body.Close()
		if closeErr != nil {
			err = closeErr
		}
	}()

	switch bodyResp.StatusCode {
	case http.StatusPartialContent:
		totalSize = currentSize + bodyResp.ContentLength
		file, errFile = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
		if errFile != nil {
			err = errFile
			return
		}

	case http.StatusOK:
		totalSize = bodyResp.ContentLength
		file, errFile = os.Create(filename)
		if errFile != nil {
			err = errFile
			return
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
		Total:        totalSize,
		Current:      currentSize,
		Destination:  file,
		StartTime:    time.Now(),
		ProgressChan: opts.Progress,
	}

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
			n, writeErr := bodyResp.Body.Read(buf)

			if writeErr == io.EOF {
				return
			} else if writeErr != nil {
				err = writeErr
				return
			}

			_, dataErr := pw.Write(buf[:n])
			if dataErr != nil {
				err = dataErr
				return
			}
		}
	}
}
