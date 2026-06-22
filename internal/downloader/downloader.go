package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func Download(id string, url string, opts DownloadOptions, ctx context.Context) (totalSize int64, filename string, err error) {
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
	tmpFilename := filename + "." + id + ".tmp"

	if info, err := os.Stat(tmpFilename); err == nil {
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
		if bodyResp.ContentLength >= 0 {
			totalSize = currentSize + bodyResp.ContentLength
		} else {
			totalSize = 0 // Unknown
		}

		file, errFile = os.OpenFile(tmpFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
		if errFile != nil {
			err = errFile
			return
		}

	case http.StatusOK:
		if bodyResp.ContentLength >= 0 {
			totalSize =  bodyResp.ContentLength
		} else {
			totalSize = 0 // Unknown
		}

		file, errFile = os.Create(tmpFilename)
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

	// Finalize the file
	finalFilename := GetUniqueFilename(filename)
	if err := os.Rename(tmpFilename, finalFilename); err != nil {
		return totalSize, filename, fmt.Errorf("failed to rename temp file to final: %v", err)
	}
	filename = finalFilename

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
