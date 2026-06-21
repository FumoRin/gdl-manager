package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func Download(url string, opts DownloadOptions, ctx context.Context) (totalSize int64, filename string, err error) {
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
		return 0, "", fmt.Errorf("unexpected status code from server: %d", bodyResp.StatusCode)
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
		ByteAtStart:  currentSize,
		Destination:  file,
		StartTime:    time.Now(),
		ProgressChan: opts.Progress,
	}

	copyChan := make(chan error, 1)
	go func() {
		_, copyErr := io.CopyBuffer(pw, bodyResp.Body, buf)
		copyChan <- copyErr
	}()

	select {
	case <-ctx.Done():
		if bodyErr := bodyResp.Body.Close(); bodyErr != nil {
			fmt.Printf("error closing response body: %v", bodyErr)
			return
		} 
		err = ctx.Err()
		return
	case copyErr := <-copyChan:
		if copyErr != nil {
			err = copyErr
			return
		}
	}

	finalUpdate := Progress{
		Filename:    filename,
		CurrentSize: totalSize,
		TotalSize:   totalSize,
		Speed:       0,
		Percentage:  100.0,
		ETA:         0,
	}
	select {
	case opts.Progress <- finalUpdate:
	default:

	}

	return totalSize, filename, nil
}
