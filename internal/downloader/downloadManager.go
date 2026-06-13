package downloader

import (
	"fmt"
)

func (m* DownloadManager) StartWorker(count int) {
	for range count{
		go func () {
			for job := range m.Job {
				opts := DownloadOptions{
					URL: job.URL,
					Filename: job.Filename,
					Progress: m.Progress,
				}

				err := Download(job.URL, opts, &m.Wg)
				if err != nil {
					fmt.Printf("Download Failed for %s: %v\n", job.URL, err)
				}
			}
		} ()
	}
}

func (m *DownloadManager) StartDownload (url string, filename string) {
	m.Wg.Add(1)

	m.Job <- DownloadJob{URL: url, Filename: filename}
}
