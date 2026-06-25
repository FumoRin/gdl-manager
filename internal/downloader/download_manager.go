package downloader

import (
	"context"
	"errors"
	"fmt"
	"os"
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

func (m *DownloadManager) RenameDownload(id string, newFilename string) error {
	state, err := m.repo.GetDownload(id)
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("download not found")
	}

	m.StopDownload(id)

	// Handle physical file rename
	oldTmp := state.Filename + "." + id + ".tmp"
	newTmp := newFilename + "." + id + ".tmp"

	if _, err := os.Stat(oldTmp); err == nil {
		if err := os.Rename(oldTmp, newTmp); err != nil {
			return fmt.Errorf("failed to rename temp file: %w", err)
		}
	}

	// If it's completed, we should also rename the final file.
	// Since we don't store the exact final filename (GetUniqueFilename adds suffixes),
	// this part is tricky. But we can try to rename state.Filename if it exists.
	if state.Status == StateCompleted {
		if _, err := os.Stat(state.Filename); err == nil {
			if err := os.Rename(state.Filename, newFilename); err != nil {
				// We ignore this error because GetUniqueFilename might have renamed it to something else
				fmt.Printf("warning: could not rename final file %s: %v\n", state.Filename, err)
			}
		}
	}

	if err := m.repo.UpdateFilename(id, newFilename); err != nil {
		return err
	}

	return nil
}

func (m *DownloadManager) DeleteDownload(id string, deleteFiles bool) error {
	state, err := m.repo.GetDownload(id)
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("download not found")
	}

	m.StopDownload(id)

	if deleteFiles {
		// Delete temp file
		tmpFile := state.Filename + "." + id + ".tmp"
		_ = os.Remove(tmpFile)
		// Delete final file
		_ = os.Remove(state.Filename)
	}

	if err := m.repo.DeleteDownload(id); err != nil {
		return err
	}

	return nil
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

	totalSize, filename, downloadErr := Download(job.ID, job.URL, opts, ctx)

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
