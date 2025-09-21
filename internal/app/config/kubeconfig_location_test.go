package config

import (
	"os"
	"testing"
)

func TestGetKubeconfigsLocations(t *testing.T) {
	t.Run("Normal case", func(t *testing.T) {
		locations := GetKubeconfigsLocations()
		if len(locations) == 0 {
			t.Error("Expected at least one location")
		}
		if len(locations) > 0 {
			home, err := os.UserHomeDir()
			if err == nil {
				expected := home + "/.kube"
				if locations[0] != expected {
					t.Errorf("Expected %s, got %s", expected, locations[0])
				}
			}
		}
	})
}

func TestGetKubeconfigsLocationsWithMockHome(t *testing.T) {
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	os.Unsetenv("HOME")
	locations := GetKubeconfigsLocations()

	if len(locations) != 1 {
		t.Error("Expected exactly one location in fallback case")
	}
	if locations[0] != "." {
		t.Errorf("Expected fallback location to be '.', got %s", locations[0])
	}
}
