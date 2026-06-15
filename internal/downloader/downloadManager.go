package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

func (m *DownloadManager) StartWorker(count int ) {
	for range count {
		go func() {
			for job := range m.Job {
				ctx, cancel := context.WithCancel(context.Background())

				m.Mu.Lock()
				m.Cancellations[job.ID] = cancel
				m.State[job.ID].Status = StateDownloading
				m.Mu.Unlock()

				opts := DownloadOptions{
					URL:      job.URL,
					Filename: job.Filename,
					Progress: m.Progress,
				}

				err := Download(job.URL, opts, &m.Wg, ctx)
				if err == nil {
					m.Mu.Lock()
					m.State[job.ID].Status = StateCompleted
					m.Mu.Unlock()
				} else {
					m.Mu.Lock()
					m.State[job.ID].Status = StateError
					m.Mu.Unlock()

					fmt.Printf("Download Failed for %s: %v\n", job.URL, err)
				}

				m.Mu.Lock()
				delete(m.Cancellations, job.ID)
				m.Mu.Unlock()
				cancel()
			}
		}()
	}
}

func (m *DownloadManager) StartDownload(generatedID string, url string, filename string) {
	m.Wg.Add(1)

	m.Mu.Lock()
	m.State[generatedID] = &DownloadState{
		ID:       generatedID,
		URL:      url,
		Filename: filename,
		Status:   StateQueue,
	}
	m.Mu.Unlock()

	m.Job <- DownloadJob{ID: generatedID, URL: url, Filename: filename}
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
	return  &DownloadManager{
		Job: make(chan DownloadJob),
		Progress: make(chan Progress),
		State: make(map[string]*DownloadState),
		Cancellations: make(map[string]context.CancelFunc),
	}
}
