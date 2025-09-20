package format

import (
	"testing"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "0B"},
		{"0", "0B"},
		{"1024", "1.0KB"},
		{"1536", "1.5KB"},
		{"1048576", "1.0MB"},
		{"1Ki", "1.0KB"},
		{"1Mi", "1.0MB"},
		{"invalid", "0B"},
	}

	for _, test := range tests {
		t.Run("Format_"+test.input, func(t *testing.T) {
			result := FormatBytes(test.input)
			if result != test.expected {
				t.Errorf("FormatBytes(%q) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}

func TestFormatBytesBinary(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "0B"},
		{"0", "0B"},
		{"1024", "1.0KiB"},
		{"1048576", "1.0MiB"},
		{"1Ki", "1.0KiB"},
		{"1Mi", "1.0MiB"},
		{"512Mi", "512.0MiB"},
		{"1024Mi", "1.0GiB"},
		{"2048Mi", "2.0GiB"},
	}

	for _, test := range tests {
		t.Run("FormatBinary_"+test.input, func(t *testing.T) {
			result := FormatBytesBinary(test.input)
			if result != test.expected {
				t.Errorf("FormatBytesBinary(%q) = %q, expected %q", test.input, result, test.expected)
			}
		})
	}
}

func TestParseByteInput(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue float64
		expectedUnit  string
	}{
		{"1024", 1024, ""},
		{"1.5Gi", 1.5, "GI"},
		{"512Mi", 512, "MI"},
		{"2GB", 2, "GB"},
		{"1.2TB", 1.2, "TB"},
		{"", 0, ""},
		{"invalid", 0, ""},
	}

	for _, test := range tests {
		t.Run("Parse_"+test.input, func(t *testing.T) {
			value, unit := parseByteInput(test.input)
			if value != test.expectedValue || unit != test.expectedUnit {
				t.Errorf("parseByteInput(%q) = (%v, %q), expected (%v, %q)",
					test.input, value, unit, test.expectedValue, test.expectedUnit)
			}
		})
	}
}

func TestConvertToBytes(t *testing.T) {
	tests := []struct {
		value    float64
		unit     string
		expected int64
	}{
		{1024, "", 1024},
		{1, "KB", 1000},
		{1, "MB", 1000000},
		{1, "GB", 1000000000},
		{1, "TB", 1000000000000},
		{1, "KI", 1024},
		{1, "MI", 1048576},
		{2.5, "GI", 2684354560},
		{512, "MI", 536870912},
	}

	for _, test := range tests {
		t.Run("Convert_"+test.unit, func(t *testing.T) {
			result := convertToBytes(test.value, test.unit)
			if result != test.expected {
				t.Errorf("convertToBytes(%v, %q) = %v, expected %v",
					test.value, test.unit, result, test.expected)
			}
		})
	}
}

func TestFormatBytesHuman(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1000000, "1.0MB"},
		{1000000000, "1.0GB"},
		{1000000000000, "1.0TB"},
		{5000000000000, "5.0TB"},
		{999, "999B"},
		{1000, "1.0KB"},
		{1000000, "1.0MB"},
	}

	for _, test := range tests {
		t.Run("Human_"+test.expected, func(t *testing.T) {
			result := formatBytesHuman(test.bytes)
			if result != test.expected {
				t.Errorf("formatBytesHuman(%v) = %q, expected %q", test.bytes, result, test.expected)
			}
		})
	}
}

func TestFormatBytesBinaryHuman(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0B"},
		{1024, "1.0KiB"},
		{1048576, "1.0MiB"},
		{2048, "2.0KiB"},
		{2097152, "2.0MiB"},
		{999, "999B"},
		{1024, "1.0KiB"},
	}

	for _, test := range tests {
		t.Run("Binary_"+test.expected, func(t *testing.T) {
			result := formatBytesBinaryHuman(test.bytes)
			if result != test.expected {
				t.Errorf("formatBytesBinaryHuman(%v) = %q, expected %q", test.bytes, result, test.expected)
			}
		})
	}
}
