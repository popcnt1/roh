package utils

import "fmt"

// HumanReadableBytes converts bytes to a human-readable format with units
func HumanReadableBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
		PB = TB * 1024
	)

	var result string
	switch {
	case bytes < KB:
		result = fmt.Sprintf("%d B", bytes)
	case bytes < MB:
		result = fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	case bytes < GB:
		result = fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes < TB:
		result = fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes < PB:
		result = fmt.Sprintf("%.2f TB", float64(bytes)/TB)
	default:
		result = fmt.Sprintf("%.2f PB", float64(bytes)/PB)
	}

	return result
}
