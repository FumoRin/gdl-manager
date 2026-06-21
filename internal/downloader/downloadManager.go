package downloader

import (
	"context"
	"errors"
	"fmt"
)

// Start a download worker for each job
func (m *DownloadManager) StartWorker(count int) {
	for range count {
		go func() {
			for {
				job, ok := <-m.job
				if !ok {
					return
				}
				m.processJob(job)
			}
		}()
	}
}

// Start a Download  
func (m *DownloadManager) StartDownload(generatedID string, url string, filename string) {
	state, err := m.repo.GetDownload(generatedID)
	if err != nil || state == nil {
		state = &DownloadState{
			ID:        generatedID,
			URL:       url,
			Filename:  filename,
			TotalSize: 0,
		}
	}
	state.Status = StateQueue

	if err := m.repo.SaveDownload(state); err != nil {
		fmt.Printf("database error: %v\n", err)
		return
	}

	m.wg.Add(1)

	select {
	case <-m.ctx.Done():
		state.Status = StatePaused
		if err := m.repo.SaveDownload(state); err != nil {
			fmt.Printf("database error: %v\n", err)
		}
		m.wg.Done()
		return

	case m.job <- DownloadJob{ID: generatedID, URL: url, Filename: filename}:
		return
	}
}

func (m *DownloadManager) StopDownload(id string) {
	m.cancellationsMu.Lock()
	cancel, ok := m.cancellations[id]
	m.cancellationsMu.Unlock()

	if ok {
		cancel()
	}
}

func NewDownloadManager(repo DownloadRepository) *DownloadManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &DownloadManager{
		job:           make(chan DownloadJob, 1000),
		progress:      make(chan Progress),
		repo:          repo,
		cancellations: make(map[string]context.CancelFunc),
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (m *DownloadManager) processJob(job DownloadJob) {
	defer m.wg.Done()
	if m.ctx.Err() != nil {
		state, err := m.repo.GetDownload(job.ID)
		if err != nil {
			return
		}
		state.Status = StatePaused
		if err := m.repo.SaveDownload(state); err != nil {
			fmt.Printf("database error: %v\n", err)
		}
		return
	}

	// If it's downloading
	ctx, cancel := context.WithCancel(m.ctx)
	m.cancellationsMu.Lock()
	m.cancellations[job.ID] = cancel
	m.cancellationsMu.Unlock()

	opts := DownloadOptions{
		URL:      job.URL,
		Filename: job.Filename,
		Progress: m.progress,
	}

	totalSize, filename, downloadErr := Download(job.URL, opts, ctx)

	state, err := m.repo.GetDownload(job.ID)
	if err != nil {
		return
	}

	if filename != "" {
		state.Filename = filename
	}

	if totalSize > 0 {
		state.TotalSize = totalSize
	}

	if downloadErr == nil {
		state.Status = StateCompleted
	} else if errors.Is(downloadErr, context.Canceled) {
		state.Status = StatePaused
	} else {
		state.Status = StateError

		fmt.Printf("Download Failed for %s: %v\n", job.URL, downloadErr)
	}

	if err := m.repo.SaveDownload(state); err != nil {
		fmt.Printf("database error: %v\n", err)
		return
	}

	m.cancellationsMu.Lock()
	delete(m.cancellations, job.ID)
	m.cancellationsMu.Unlock()

	cancel()
}
