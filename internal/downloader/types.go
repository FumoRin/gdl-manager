package downloader

import (
	"context"
	"fmt"
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
	wg              sync.WaitGroup
	progress        chan Progress
	job             chan DownloadJob
	repo            DownloadRepository
	cancellationsMu sync.Mutex
	cancellations   map[string]context.CancelFunc
	ctx             context.Context
	cancel          context.CancelFunc
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
	ByteAtStart  int64
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
	names := [...]string{"Queued", "Downloading", "Paused", "Completed", "Error"}
	if int(s) < 0 || int(s) >= len(names) {
		return fmt.Sprintf("Undefined(%d)", s)
	}
	return names[s]
}

type PartState struct {
	ID          string
	DownloadID  string
	StartByte   int64
	EndByte     int64
	CurrentByte int64
	WorkerID    string
}

type DownloadRepository interface {
	SaveDownload(state *DownloadState) error
	GetDownload(id string) (*DownloadState, error)
	GetIncompleteDownload() ([]*DownloadState, error)
	GetAllDownloads() ([]*DownloadState, error)
	DeleteDownload(id string) error
	UpdatePartsProgress(partID string, currentByte int64) error
	GetParts(downloadID string) ([]*PartState, error)
	CreatePart(part *PartState) error
	UpdateFilename(id string, newFilename string) error
}
