package downloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

func (m *DownloadManager) StartWorker(count int) {
	for range count {
		go func() {
			for {
				select {
				case <-m.Ctx.Done():
					return

				case job, ok := <-m.Job:
					if !ok {
						return
					}

					if m.Ctx.Err() != nil {
						m.Mu.Lock()
						m.State[job.ID].Status = StatePaused
						m.Mu.Unlock()
						m.Wg.Done()
						return
					}

					ctx, cancel := context.WithCancel(m.Ctx)
					m.Mu.Lock()
					m.Cancellations[job.ID] = cancel
					m.State[job.ID].Status = StateDownloading
					m.Mu.Unlock()

					opts := DownloadOptions{
						URL:      job.URL,
						Filename: job.Filename,
						Progress: m.Progress,
					}

					totalSize, filename, err := Download(job.URL, opts, &m.Wg, ctx)

					m.Mu.Lock()

					if filename != "" {
						m.State[job.ID].Filename = filename
					}

					if totalSize > 0 {
						m.State[job.ID].TotalSize = totalSize
					}

					if err == nil {
						m.State[job.ID].Status = StateCompleted
					} else if errors.Is(err, context.Canceled) {
						m.State[job.ID].Status = StatePaused
					} else {
						m.State[job.ID].Status = StateError

						fmt.Printf("Download Failed for %s: %v\n", job.URL, err)
					}


					delete(m.Cancellations, job.ID)
					m.Mu.Unlock()

					cancel()

				}
			}
		}()
	}
}

func (m *DownloadManager) StartDownload(generatedID string, url string, filename string) {
	m.Mu.Lock()
	m.State[generatedID] = &DownloadState{
		ID:       generatedID,
		URL:      url,
		Filename: filename,
		Status:   StateQueue,
	}
	m.Mu.Unlock()

	select {
	case <-m.Ctx.Done():
		m.Mu.Lock()
		m.State[generatedID].Status = StatePaused
		m.Mu.Unlock()
		m.Wg.Done()
		return

	case m.Job <- DownloadJob{ID: generatedID, URL: url, Filename: filename}:
		return
	}
}

func (m *DownloadManager) StopDownload(id string) {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	if cancel, ok := m.Cancellations[id]; ok {
		cancel()
	}
}

func (m *DownloadManager) SaveState(filepath string) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	data, err := json.MarshalIndent(m.State, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath, data, 0o644)
}

func (m *DownloadManager) LoadState(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	m.Mu.Lock()
	defer m.Mu.Unlock()

	return json.Unmarshal(data, &m.State)
}

func NewDownloadManager() *DownloadManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &DownloadManager{
		Job:           make(chan DownloadJob),
		Progress:      make(chan Progress),
		State:         make(map[string]*DownloadState),
		Cancellations: make(map[string]context.CancelFunc),
		Ctx:           ctx,
		Cancel:        cancel,
	}
}
