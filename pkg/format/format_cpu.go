package format

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// FormatCPU formats CPU values to be more readable
// Input examples: "24431948n", "100m", "1", "0.5"
// Output examples: "25m", "100m", "1000m", "500m"
func FormatCPU(input string) string {
	if input == "" {
		return "0m"
	}

	// Parse the input value and unit
	value, unit := parseCPUInput(input)

	// Convert everything to millicores for consistent display
	millicores := convertToMillicores(value, unit)

	return formatCPUMillicores(millicores)
}

func parseCPUInput(input string) (float64, string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return 0, ""
	}

	// Handle cases like "24431948n" or "100m" or "1" or "0.5"
	re := regexp.MustCompile(`^(\d+(?:\.\d+)?)\s*([a-zA-Z]*)$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) == 3 {
		value, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			return 0, ""
		}
		unit := strings.ToLower(matches[2])
		return value, unit
	}

	// Try parsing as plain number
	value, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, ""
	}
	return value, ""
}

func convertToMillicores(value float64, unit string) float64 {
	switch unit {
	case "", "cores", "core":
		// Whole cores - convert to millicores
		return value * 1000
	case "m":
		// Already in millicores
		return value
	case "n", "nano":
		// Nanocores - convert to millicores
		return value / 1000000
	case "u", "micro":
		// Microcores - convert to millicores
		return value / 1000
	default:
		// Assume cores if no unit
		return value * 1000
	}
}

func formatCPUMillicores(millicores float64) string {
	if millicores == 0 {
		return "0m"
	}

	// Use millicores for values less than 1000m (1 core)
	if millicores < 1000 {
		if millicores < 0.001 {
			// For extremely small values, show as 0m
			return "0m"
		}
		if millicores < 1 {
			// For very small values, show with appropriate precision
			return fmt.Sprintf("%.3fm", millicores)
		}
		// Round to nearest millicore
		return fmt.Sprintf("%.0fm", math.Round(millicores))
	}

	// For values >= 1000m, show as cores
	cores := millicores / 1000
	if cores < 10 {
		return fmt.Sprintf("%.1f", cores)
	}
	return fmt.Sprintf("%.0f", cores)
}

// FormatCPUWithUnit formats CPU values with appropriate units
// This is an alternative that shows units dynamically (m, cores)
func FormatCPUWithUnit(input string) string {
	if input == "" {
		return "0m"
	}

	value, unit := parseCPUInput(input)
	millicores := convertToMillicores(value, unit)

	if millicores < 1000 {
		if millicores < 1 {
			return fmt.Sprintf("%.3fm", millicores)
		}
		return fmt.Sprintf("%.0fm", millicores)
	}

	cores := millicores / 1000
	if cores < 10 {
		return fmt.Sprintf("%.1f cores", cores)
	}
	return fmt.Sprintf("%.0f cores", cores)
}
