package downloader

import (
	"fmt"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"
)

func FormatBytes(b int64) string {
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

func FormatTime(d time.Duration) string {
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

func Truncate(s string, max int) string {
    if len(s) > max {
        return s[:max-3] + "..."
    }
    return s
}

func resolveFilename(url string, opts DownloadOptions, resp *http.Response) string {
	if opts.Filename != "" {
		return opts.Filename
	}

	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil && params["filename"] != "" {
			return params["filename"]
		}
	}

	return path.Base(url)
}
