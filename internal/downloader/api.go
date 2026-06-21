package downloader

// Handling SIGTERM by doing a Graceful shutdown
func (m *DownloadManager) Shutdown() {
	m.cancel()
}

// Replacement for closing a m.Job channel
func (m *DownloadManager) Close() {
	close(m.job)
}

// Replacement for m.Wg.Wait()
func (m *DownloadManager) Wait() {
	m.wg.Wait()
	close(m.progress)
}

// Passing ProgressChan for Read-Only Access on Progress channel
func (m *DownloadManager) ProgressChan() <- chan Progress {
	return m.progress
}

// Repo access delegation for Get Download
func (m *DownloadManager) GetDownload(id string) (*DownloadState, error) {
	return m.repo.GetDownload(id)
}

// Repo access delegation for Get Incomplete Download
func (m *DownloadManager) GetIncompleteDownload() ([]*DownloadState, error) {
	return m.repo.GetIncompleteDownload()
}

// Repo access delegation for Get Incomplete Download
func (m *DownloadManager) GetAllDownload() ([]*DownloadState, error) {
	return m.repo.GetAllDownloads()
}
