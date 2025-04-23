package finder

import (
	"fmt"
	"regexp"

	"github.com/alewtschuk/pfmt"
)

// Strips any ANSI color codes from input string
func StripColor(s string) string {
	var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiEscape.ReplaceAllString(s, "")
}

// formatSize takes a size in bytes and returns a human-readable string
func FormatSize(size int64) string {
	const (
		_          = iota
		KB float64 = 1 << (10 * iota)
		MB
		GB
		TB
	)

	floatSize := float64(size)

	switch {
	case floatSize >= TB:
		return pfmt.ApplyColor(fmt.Sprintf("%.1f TB", floatSize/TB), 196)
	case floatSize >= GB:
		return pfmt.ApplyColor(fmt.Sprintf("%.1f GB", floatSize/GB), 6)
	case floatSize >= MB:
		return pfmt.ApplyColor(fmt.Sprintf("%.1f MB", floatSize/MB), 150)
	case floatSize >= KB:
		return pfmt.ApplyColor(fmt.Sprintf("%.1f KB", floatSize/KB), 70)
	default:
		return pfmt.ApplyColor(fmt.Sprintf("%d B", size), 205)
	}
}
