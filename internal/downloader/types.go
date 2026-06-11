package downloader

import (
	"io"
	"time"
)

type DownloadOptions struct {
	URL      string
	Filename string
	Size     int64
}

type ProgressWriter struct {
	Total int64
	Current int64
	Destination io.Writer

	StartTime time.Time
}
