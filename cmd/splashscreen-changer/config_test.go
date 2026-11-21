package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create necessary directories
	tmpDir := os.TempDir()
	sourceDirPath := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(sourceDirPath, os.ModePerm); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(sourceDirPath); err != nil {
			t.Errorf("Failed to remove source directory: %v", err)
		}
	}()
	destinationDirPath := filepath.Join(tmpDir, "destination")
	if err := os.MkdirAll(filepath.Join(destinationDirPath, "EasyAntiCheat"), os.ModePerm); err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(destinationDirPath); err != nil {
			t.Fatalf("Failed to remove destination directory: %v", err)
		}
	}()

	// Create a temporary config file
	configContent := `
source:
	path: {{ .Source.Path }}
	recursive: true
destination:
	path: {{ .Destination.Path }}
	width: 1024
	height: 768
`
	configContent = strings.TrimSpace(configContent)
	configContent = strings.ReplaceAll(configContent, "\t", "  ")
	configContent = strings.ReplaceAll(configContent, "{{ .Source.Path }}", sourceDirPath)
	configContent = strings.ReplaceAll(configContent, "{{ .Destination.Path }}", destinationDirPath)

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Load the config
	config, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check the loaded config values
	if config.Source.Path != filepath.Join(tmpDir, "source") {
		t.Errorf("Expected source path to be '%s', got '%s'", filepath.Join(tmpDir, "source"), config.Source.Path)
	}
	if !config.Source.Recursive {
		t.Errorf("Expected source recursive to be true, got false")
	}
	if config.Destination.Path != filepath.Join(tmpDir, "destination") {
		t.Errorf("Expected destination path to be '%s', got '%s'", filepath.Join(tmpDir, "destination"), config.Destination.Path)
	}
	if config.Destination.Width != 1024 {
		t.Errorf("Expected destination width to be 1024, got %d", config.Destination.Width)
	}
	if config.Destination.Height != 768 {
		t.Errorf("Expected destination height to be 768, got %d", config.Destination.Height)
	}
}

func TestLoadConfigWithDefaults(t *testing.T) {
	// Create necessary directories
	tmpDir := os.TempDir()
	sourceDirPath := filepath.Join(tmpDir, "source")
	os.MkdirAll(sourceDirPath, os.ModePerm)
	defer os.RemoveAll(sourceDirPath)
	destinationDirPath := filepath.Join(tmpDir, "destination")
	os.MkdirAll(filepath.Join(destinationDirPath, "EasyAntiCheat"), os.ModePerm)
	defer os.RemoveAll(destinationDirPath)

	// Create a temporary config file
	configContent := `
source:
	path: {{ .Source.Path }}
destination:
	path: {{ .Destination.Path }}
`
	configContent = strings.TrimSpace(configContent)
	configContent = strings.ReplaceAll(configContent, "\t", "  ")
	configContent = strings.ReplaceAll(configContent, "{{ .Source.Path }}", sourceDirPath)
	configContent = strings.ReplaceAll(configContent, "{{ .Destination.Path }}", destinationDirPath)

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Load the config
	config, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check the loaded config values
	if config.Source.Path != filepath.Join(tmpDir, "source") {
		t.Errorf("Expected source path to be '%s', got '%s'", filepath.Join(tmpDir, "source"), config.Source.Path)
	}
	if !config.Source.Recursive {
		t.Errorf("Expected source recursive to be true got false")
	}
	if config.Destination.Path != filepath.Join(tmpDir, "destination") {
		t.Errorf("Expected destination path to be '%s', got '%s'", filepath.Join(tmpDir, "destination"), config.Destination.Path)
	}
	if config.Destination.Width != 800 {
		t.Errorf("Expected destination width to be 800, got %d", config.Destination.Width)
	}
	if config.Destination.Height != 450 {
		t.Errorf("Expected destination height to be 450, got %d", config.Destination.Height)
	}
}

func TestLoadConfigWithEnvOverrides(t *testing.T) {
	// Create necessary directories
	tmpDir := os.TempDir()
	envSourceDirPath := filepath.Join(tmpDir, "env", "source")
	os.MkdirAll(envSourceDirPath, os.ModePerm)
	defer os.RemoveAll(envSourceDirPath)
	envDestinationDirPath := filepath.Join(tmpDir, "env", "destination")
	os.MkdirAll(filepath.Join(envDestinationDirPath, "EasyAntiCheat"), os.ModePerm)
	defer os.RemoveAll(envDestinationDirPath)
	configSourceDirPath := filepath.Join(tmpDir, "source")
	os.MkdirAll(configSourceDirPath, os.ModePerm)
	defer os.RemoveAll(configSourceDirPath)
	configDestinationDirPath := filepath.Join(tmpDir, "destination")
	os.MkdirAll(filepath.Join(configDestinationDirPath, "EasyAntiCheat"), os.ModePerm)
	defer os.RemoveAll(configDestinationDirPath)

	// Set environment variables
	customEnv := map[string]string{
		"SOURCE_PATH":        envSourceDirPath,
		"SOURCE_RECURSIVE":   "true",
		"DESTINATION_PATH":   envDestinationDirPath,
		"DESTINATION_WIDTH":  "1280",
		"DESTINATION_HEIGHT": "720",
	}
	for key, value := range customEnv {
		os.Setenv(key, value)
	}
	defer func() {
		for key := range customEnv {
			os.Unsetenv(key)
		}
	}()

	// Create a temporary config file
	configContent := `
source:
	path: {{ .Source.Path }}
	recursive: false
destination:
	path: {{ .Destination.Path }}
	width: 1024
	height: 768
`
	configContent = strings.TrimSpace(configContent)
	configContent = strings.ReplaceAll(configContent, "\t", "  ")
	configContent = strings.ReplaceAll(configContent, "{{ .Source.Path }}", configSourceDirPath)
	configContent = strings.ReplaceAll(configContent, "{{ .Destination.Path }}", configDestinationDirPath)
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(configContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Load the config
	config, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Check the loaded config values
	if config.Source.Path != envSourceDirPath {
		t.Errorf("Expected source path to be '%s', got '%s'", envSourceDirPath, config.Source.Path)
	}
	if !config.Source.Recursive {
		t.Errorf("Expected source recursive to be true, got false")
	}
	if config.Destination.Path != envDestinationDirPath {
		t.Errorf("Expected destination path to be '%s', got '%s'", envDestinationDirPath, config.Destination.Path)
	}
	if config.Destination.Width != 1280 {
		t.Errorf("Expected destination width to be 1280, got %d", config.Destination.Width)
	}
	if config.Destination.Height != 720 {
		t.Errorf("Expected destination height to be 720, got %d", config.Destination.Height)
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"SOURCE_RECURSIVE", "DESTINATION_WIDTH", "DESTINATION_HEIGHT", "SOURCE_PATH", "DESTINATION_PATH"} {
		t.Setenv(key, "")
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
	return path
}

func newValidDirsConfig(t *testing.T, extraSource string) string {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	destination := filepath.Join(root, "destination")
	if err := os.MkdirAll(source, os.ModePerm); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(destination, "EasyAntiCheat"), os.ModePerm); err != nil {
		t.Fatalf("Failed to create destination directory: %v", err)
	}
	content := "source:\n  path: " + source + "\n" + extraSource + "destination:\n  path: " + destination + "\n"
	return writeTempConfig(t, content)
}

func TestLoadConfigRecursiveFalseFromFile(t *testing.T) {
	clearConfigEnv(t)
	configPath := newValidDirsConfig(t, "  recursive: false\n")

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config.Source.Recursive {
		t.Errorf("Expected source recursive to be false, got true")
	}
}

func TestLoadConfigRecursiveFalseFromEnv(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("SOURCE_RECURSIVE", "false")
	configPath := newValidDirsConfig(t, "")

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if config.Source.Recursive {
		t.Errorf("Expected source recursive to be false, got true")
	}
}

func TestLoadConfigExplicitZeroWidth(t *testing.T) {
	clearConfigEnv(t)
	configPath := writeTempConfig(t, "destination:\n  width: 0\n")

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatalf("Expected an error for destination.width 0, got nil")
	}
	if err.Error() != "destination width must be greater than 0" {
		t.Errorf("Unexpected error: %q", err.Error())
	}
}
