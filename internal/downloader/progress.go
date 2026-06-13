package downloader

import "time"

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.Destination.Write(p)
	if err != nil {
		return n, err
	}
	pw.Current += int64(n)

	var percentage float64

	if pw.Total > 0 {
		percentage = (float64(pw.Current)/float64(pw.Total)) * 100
	}

	if time.Since(pw.LastUpdate) > 200*time.Millisecond {
		pw.LastUpdate = time.Now()
		update := Progress{
			Filename: pw.Filename,
			CurrentSize: pw.Current,
			TotalSize: pw.Total,
			Speed: float64(pw.Speed()),
			Percentage: percentage,
			ETA: pw.Eta(),
		}

		pw.ProgressChan <- update
	}

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
	var result time.Duration
	remaining := pw.Total - pw.Current
	speed := pw.Speed()
	if speed <= 0 {
		return 0
	}

	eta := float64(remaining) / speed
	if eta > 0 {
		result = time.Duration(eta * 1000000000)
	} else {
		result = 0
	}

	return result
}
