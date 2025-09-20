package utils

import (
	"testing"
	"time"
)

func TestFormatAge(t *testing.T) {
	now := time.Now()

	t.Run("Seconds", func(t *testing.T) {
		past := now.Add(-30 * time.Second)
		result := FormatAge(past)
		if result != "30s" {
			t.Errorf("Expected '30s', got '%s'", result)
		}
	})

	t.Run("Minutes and seconds", func(t *testing.T) {
		past := now.Add(-5*time.Minute - 30*time.Second)
		result := FormatAge(past)
		if result != "5m30s" {
			t.Errorf("Expected '5m30s', got '%s'", result)
		}
	})

	t.Run("Hours and minutes", func(t *testing.T) {
		past := now.Add(-2*time.Hour - 15*time.Minute)
		result := FormatAge(past)
		if result != "2h15m" {
			t.Errorf("Expected '2h15m', got '%s'", result)
		}
	})

	t.Run("Days and hours", func(t *testing.T) {
		past := now.Add(-3*24*time.Hour - 4*time.Hour)
		result := FormatAge(past)
		if result != "3d4h" {
			t.Errorf("Expected '3d4h', got '%s'", result)
		}
	})

	t.Run("Years", func(t *testing.T) {
		past := now.Add(-2 * 365 * 24 * time.Hour)
		result := FormatAge(past)
		if result != "2y" {
			t.Errorf("Expected '2y', got '%s'", result)
		}
	})

	t.Run("Years and months", func(t *testing.T) {
		past := now.Add(-2*365*24*time.Hour - 60*24*time.Hour)
		result := FormatAge(past)
		if result != "2y2mo" {
			t.Errorf("Expected '2y2mo', got '%s'", result)
		}
	})

	t.Run("Edge case: exactly 1 minute", func(t *testing.T) {
		past := now.Add(-1 * time.Minute)
		result := FormatAge(past)
		if result != "1m0s" {
			t.Errorf("Expected '1m0s', got '%s'", result)
		}
	})

	t.Run("Edge case: exactly 1 hour", func(t *testing.T) {
		past := now.Add(-1 * time.Hour)
		result := FormatAge(past)
		if result != "1h0m" {
			t.Errorf("Expected '1h0m', got '%s'", result)
		}
	})
}
