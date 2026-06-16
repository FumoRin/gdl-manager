package downloader

import (
	"context"
	"io"
	"sync"
	"time"
)

type DownloadStatus int

const (
	StateQueue DownloadStatus = iota
	StateDownloading
	StatePaused
	StateCompleted
	StateError
)

type DownloadOptions struct {
	URL      string
	Filename string
	Size     int64
	Progress chan Progress
}

type DownloadJob struct {
	ID       string
	URL      string
	Filename string
}

type DownloadManager struct {
	Wg            sync.WaitGroup
	Progress      chan Progress
	Job           chan DownloadJob
	State         map[string]*DownloadState
	Mu            sync.Mutex
	Cancellations map[string]context.CancelFunc
	Ctx           context.Context
	Cancel        context.CancelFunc
}

type DownloadState struct {
	ID        string
	URL       string
	Filename  string
	TotalSize int64
	Status    DownloadStatus
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

func (s DownloadStatus) String() string {
	return [...]string{"Queued", "Downloading", "Paused", "Completed", "Error"}[s]
}
