package format

import (
	"testing"
)

func TestFormatCPU(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Nanocores to millicores",
			input:    "24431948n",
			expected: "24m",
		},
		{
			name:     "Millicores unchanged",
			input:    "100m",
			expected: "100m",
		},
		{
			name:     "Whole cores to millicores",
			input:    "1",
			expected: "1.0",
		},
		{
			name:     "Fractional cores to millicores",
			input:    "0.5",
			expected: "500m",
		},
		{
			name:     "Multiple cores to cores",
			input:    "2.5",
			expected: "2.5",
		},
		{
			name:     "Large number of cores",
			input:    "10",
			expected: "10",
		},
		{
			name:     "Microcores to millicores",
			input:    "500u",
			expected: "0.500m",
		},
		{
			name:     "Very small nanocores",
			input:    "100n",
			expected: "0m",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "0m",
		},
		{
			name:     "Explicit cores unit",
			input:    "2cores",
			expected: "2.0",
		},
		{
			name:     "Explicit core unit",
			input:    "1core",
			expected: "1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatCPU(tt.input)
			if result != tt.expected {
				t.Errorf("FormatCPU(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseCPUInput(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedVal  float64
		expectedUnit string
	}{
		{
			name:         "Nanocores",
			input:        "24431948n",
			expectedVal:  24431948,
			expectedUnit: "n",
		},
		{
			name:         "Millicores",
			input:        "100m",
			expectedVal:  100,
			expectedUnit: "m",
		},
		{
			name:         "Whole number",
			input:        "1",
			expectedVal:  1,
			expectedUnit: "",
		},
		{
			name:         "Fractional number",
			input:        "0.5",
			expectedVal:  0.5,
			expectedUnit: "",
		},
		{
			name:         "With spaces",
			input:        " 100m ",
			expectedVal:  100,
			expectedUnit: "m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, unit := parseCPUInput(tt.input)
			if val != tt.expectedVal {
				t.Errorf("parseCPUInput(%q) value = %v, want %v", tt.input, val, tt.expectedVal)
			}
			if unit != tt.expectedUnit {
				t.Errorf("parseCPUInput(%q) unit = %q, want %q", tt.input, unit, tt.expectedUnit)
			}
		})
	}
}

func TestConvertToMillicores(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		unit     string
		expected float64
	}{
		{
			name:     "Cores to millicores",
			value:    1,
			unit:     "",
			expected: 1000,
		},
		{
			name:     "Explicit cores to millicores",
			value:    2,
			unit:     "cores",
			expected: 2000,
		},
		{
			name:     "Millicores unchanged",
			value:    500,
			unit:     "m",
			expected: 500,
		},
		{
			name:     "Nanocores to millicores",
			value:    1000000,
			unit:     "n",
			expected: 1,
		},
		{
			name:     "Microcores to millicores",
			value:    1000,
			unit:     "u",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToMillicores(tt.value, tt.unit)
			if result != tt.expected {
				t.Errorf("convertToMillicores(%v, %q) = %v, want %v", tt.value, tt.unit, result, tt.expected)
			}
		})
	}
}

func TestFormatCPUMillicores(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{
			name:     "Zero",
			input:    0,
			expected: "0m",
		},
		{
			name:     "Small millicores",
			input:    0.5,
			expected: "0.500m",
		},
		{
			name:     "Very small millicores",
			input:    0.123,
			expected: "0.123m",
		},
		{
			name:     "Normal millicores",
			input:    100,
			expected: "100m",
		},
		{
			name:     "Just under 1 core",
			input:    999,
			expected: "999m",
		},
		{
			name:     "Exactly 1 core",
			input:    1000,
			expected: "1.0",
		},
		{
			name:     "Fractional cores",
			input:    1500,
			expected: "1.5",
		},
		{
			name:     "Multiple cores",
			input:    2500,
			expected: "2.5",
		},
		{
			name:     "Large cores",
			input:    10000,
			expected: "10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCPUMillicores(tt.input)
			if result != tt.expected {
				t.Errorf("formatCPUMillicores(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
