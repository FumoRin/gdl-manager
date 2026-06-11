package downloader

import "time"

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Destination.Write(p)
	if err != nil {
		return n, err
	}

	pw.Current += int64(n)

	return n, err
}

func (pw *ProgressWriter) Speed() float64 {
	elapsed := time.Since(pw.StartTime).Seconds()

	if elapsed == 0 {
		return 0
	}

	return float64(pw.Current) / elapsed
}

func (pw *ProgressWriter) Eta() time.Duration {
	remaining := pw.Total - pw.Current
	speed := pw.Speed()

	eta := float64(remaining) / speed

	// Return ETA in nanoseconds
	return time.Duration(eta * 1000000000)
}
