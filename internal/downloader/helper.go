package downloader

import (
	"fmt"
	"strings"
	"time"
)

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTP"[exp])
}

func (pw *ProgressWriter) FormatTime(d time.Duration) string {
	totaltime := int64(d.Seconds())
	hour := totaltime / 3600
	remain := totaltime % 3600
	minute := remain / 60
	second := remain % 60

	parts := []string{}

	if hour > 0 {
		hstr := fmt.Sprintf("%d Hours", hour)
		parts = append(parts, hstr)
	}

	if minute > 0 {
		mstr := fmt.Sprintf("%d Minutes", minute)
		parts = append(parts, mstr)
	}

	sstr := fmt.Sprintf("%d Seconds", second)
	parts = append(parts, sstr)

	result := strings.Join(parts, " ")

	return result
}
