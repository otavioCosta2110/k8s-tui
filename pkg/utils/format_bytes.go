package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func FormatBytes(input string) string {
	if input == "" {
		return "0B"
	}

	value, unit := parseByteInput(input)

	bytes := convertToBytes(value, unit)

	return formatBytesHuman(bytes)
}

func parseByteInput(input string) (float64, string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, ""
	}

	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([A-Za-z]*)$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) == 3 {
		value, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return 0, ""
		}
		unit := strings.ToUpper(matches[2])
		if unit == "" {
			return value, ""
		}
		switch unit {
		case "K", "KI":
			return value, "KI"
		case "M", "MI":
			return value, "MI"
		case "G", "GI":
			return value, "GI"
		case "T", "TI":
			return value, "TI"
		case "KB":
			return value, "KB"
		case "MB":
			return value, "MB"
		case "GB":
			return value, "GB"
		case "TB":
			return value, "TB"
		default:
			return 0, ""
		}
	}

	value, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, ""
	}
	return value, ""
}

func convertToBytes(value float64, unit string) int64 {
	switch unit {
	case "B", "":
		return int64(value)
	case "KB":
		return int64(value * 1000)
	case "MB":
		return int64(value * 1000 * 1000)
	case "GB":
		return int64(value * 1000 * 1000 * 1000)
	case "TB":
		return int64(value * 1000 * 1000 * 1000 * 1000)
	case "KI":
		return int64(value * 1024)
	case "MI":
		return int64(value * 1024 * 1024)
	case "GI":
		return int64(value * 1024 * 1024 * 1024)
	case "TI":
		return int64(value * 1024 * 1024 * 1024 * 1024)
	default:
		return int64(value)
	}
}

func formatBytesHuman(bytes int64) string {
	if bytes == 0 {
		return "0B"
	}

	const (
		KB = 1000
		MB = KB * 1000
		GB = MB * 1000
		TB = GB * 1000
	)

	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1fTB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

func FormatBytesBinary(input string) string {
	if input == "" {
		return "0B"
	}

	value, unit := parseByteInput(input)

	bytes := convertToBytesBinary(value, unit)

	return formatBytesBinaryHuman(bytes)
}

func convertToBytesBinary(value float64, unit string) int64 {
	switch unit {
	case "B", "":
		return int64(value)
	case "KI":
		return int64(value * 1024)
	case "MI":
		return int64(value * 1024 * 1024)
	case "GI":
		return int64(value * 1024 * 1024 * 1024)
	case "TI":
		return int64(value * 1024 * 1024 * 1024 * 1024)
	default:
		return int64(value)
	}
}

func formatBytesBinaryHuman(bytes int64) string {
	if bytes == 0 {
		return "0B"
	}

	const (
		KiB = 1024
		MiB = KiB * 1024
		GiB = MiB * 1024
		TiB = GiB * 1024
	)

	switch {
	case bytes >= TiB:
		return fmt.Sprintf("%.1fTiB", float64(bytes)/float64(TiB))
	case bytes >= GiB:
		return fmt.Sprintf("%.1fGiB", float64(bytes)/float64(GiB))
	case bytes >= MiB:
		return fmt.Sprintf("%.1fMiB", float64(bytes)/float64(MiB))
	case bytes >= KiB:
		return fmt.Sprintf("%.1fKiB", float64(bytes)/float64(KiB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
