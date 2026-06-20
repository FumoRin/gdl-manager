package downloader

import (
	"context"
	"errors"
	"fmt"
	"log"
)

func (m *DownloadManager) StartWorker(count int) {
	for range count {
		go func() {
			for {
				job, ok := <-m.Job
				if !ok {
					return
				}
				m.processJob(job)
			}
		}()
	}
}

func (m *DownloadManager) StartDownload(generatedID string, url string, filename string) {
	state, err := m.Repo.GetDownload(generatedID)
	if err != nil || state == nil {
		state = &DownloadState{
			ID:        generatedID,
			URL:       url,
			Filename:  filename,
			TotalSize: 0,
		}
	}
	state.Status = StateQueue

	if err := m.Repo.SaveDownload(state); err != nil {
		log.Printf("database error: %v\n", err)
		return
	}

	select {
	case <-m.Ctx.Done():
		state.Status = StatePaused
		if err := m.Repo.SaveDownload(state); err != nil {
			log.Printf("database error: %v\n", err)
		}
		m.Wg.Done()
		return

	case m.Job <- DownloadJob{ID: generatedID, URL: url, Filename: filename}:
		return
	}
}

func (m *DownloadManager) StopDownload(id string) {
	if cancel, ok := m.Cancellations[id]; ok {
		cancel()
	}
}

func NewDownloadManager(repo DownloadRepository) *DownloadManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &DownloadManager{
		Job:           make(chan DownloadJob),
		Progress:      make(chan Progress),
		Repo:          repo,
		Cancellations: make(map[string]context.CancelFunc),
		Ctx:           ctx,
		Cancel:        cancel,
	}
}

func (m *DownloadManager) processJob(job DownloadJob) {
	defer m.Wg.Done()
	if m.Ctx.Err() != nil {
		state, err := m.Repo.GetDownload(job.ID)
		if err != nil {
			return
		}
		state.Status = StatePaused
		if err := m.Repo.SaveDownload(state); err != nil {
			log.Printf("database error: %v\n", err)
		}
		return
	}

	// If it's downloading
	ctx, cancel := context.WithCancel(m.Ctx)
	m.Cancellations[job.ID] = cancel

	opts := DownloadOptions{
		URL:      job.URL,
		Filename: job.Filename,
		Progress: m.Progress,
	}

	totalSize, filename, downloadErr := Download(job.URL, opts, &m.Wg, ctx)

	state, err := m.Repo.GetDownload(job.ID)
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

	if err := m.Repo.SaveDownload(state); err != nil {
		fmt.Printf("database error: %v\n", err)
		return
	}
	cancel()
}
