package downloader

import (
	"io"
	"sync"
	"time"
)

type DownloadOptions struct {
	URL      string
	Filename string
	Size     int64
	Progress chan Progress
}

type DownloadJob struct {
	URL      string
	Filename string
}

type DownloadManager struct {
	Wg       sync.WaitGroup
	Progress chan Progress
	Job      chan DownloadJob
}

type ProgressWriter struct {
	Filename     string
	Total        int64
	Current      int64
	Destination  io.Writer
	ProgressChan chan Progress

	StartTime  time.Time
	LastUpdate time.Time
}

type Progress struct {
	Filename    string
	Percentage  float64
	CurrentSize int64
	TotalSize   int64
	Speed       float64
	ETA         time.Duration
}
