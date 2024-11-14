package main

import (
	"testing"
)

func TestGetAppVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    interface{}
	}{
		{"Version set without v", "1.0.0", "1.0.0"},
		{"Version set with v", "v1.0.0", "1.0.0"},
		{"Version not set, build info available", "", []string{"(devel)", "unknown"}}, // Assuming no build info available in test environment
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version = tt.version
			got := GetAppVersion()
			switch want := tt.want.(type) {
			case string:
				if got != want {
					t.Errorf("GetAppVersion() = %v, want %v", got, want)
				}
			case [2]string:
				if got != want[0] && got != want[1] {
					t.Errorf("GetAppVersion() = %v, want %v or %v", got, want[0], want[1])
				}
			}
		})
	}
}

func TestGetAppDate(t *testing.T) {
	tests := []struct {
		name string
		date string
		want string
	}{
		{"Date set", "2023-10-01", "2023-10-01"},
		{"Date not set", "", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date = tt.date
			if got := GetAppDate(); got != tt.want {
				t.Errorf("GetAppDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
