package main

import (
	"errors"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainConfigLoadFailureExitCode(t *testing.T) {
	if os.Getenv("SPLASH_TEST_RUN_MAIN") == "1" {
		os.Args = []string{"splashscreen-changer", "-config", os.Getenv("SPLASH_TEST_CONFIG")}
		os.Unsetenv("CONFIG_PATH")
		main()
		return
	}

	cases := map[string]string{
		"invalid YAML": "source: [unclosed",
		"zero width":   "destination:\n  width: 0\n",
	}
	for name, content := range cases {
		t.Run(name, func(t *testing.T) {
			configPath := filepath.Join(t.TempDir(), "config.yml")
			if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			cmd := exec.Command(os.Args[0], "-test.run=^TestMainConfigLoadFailureExitCode$")
			cmd.Env = append(os.Environ(),
				"SPLASH_TEST_RUN_MAIN=1",
				"SPLASH_TEST_CONFIG="+configPath,
				"DESTINATION_WIDTH=", "DESTINATION_HEIGHT=", "SOURCE_PATH=", "DESTINATION_PATH=",
			)
			output, err := cmd.CombinedOutput()

			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) {
				t.Fatalf("Expected an *exec.ExitError, got %v\noutput: %s", err, output)
			}
			if exitErr.ExitCode() != 1 {
				t.Errorf("Expected exit code 1, got %d\noutput: %s", exitErr.ExitCode(), output)
			}
			if !strings.Contains(string(output), "Failed to load configuration file:") {
				t.Errorf("Expected output to contain the failure message, got: %s", output)
			}
		})
	}
}

// Test listPNGFiles function
func TestListPNGFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir := os.TempDir()
	tmpDir = filepath.Join(tmpDir, "listPNGFiles")
	os.Mkdir(tmpDir, os.ModePerm)

	// Create some test PNG files
	file1, _ := os.Create(filepath.Join(tmpDir, "test1.png"))
	file1.Close()
	file2, _ := os.Create(filepath.Join(tmpDir, "test2.png"))
	file2.Close()

	// Create a non-PNG file
	file3, _ := os.Create(filepath.Join(tmpDir, "test3.txt"))
	file3.Close()

	// Test non-recursive listing
	files, err := listPNGFiles(tmpDir, false)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("Expected 2 PNG files, got %d", len(files))
	}

	// Test recursive listing
	subDir := filepath.Join(tmpDir, "subdir")
	os.Mkdir(subDir, 0755)
	file4, _ := os.Create(filepath.Join(subDir, "test4.png"))
	file4.Close()

	files, err = listPNGFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("Expected 3 PNG files, got %d", len(files))
	}
}

// Test pickRandomFile function
func TestPickRandomFile(t *testing.T) {
	files := []string{"file1.png", "file2.png", "file3.png"}
	pickedFile, err := pickRandomFile(files)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if pickedFile == "" {
		t.Fatalf("Expected a file to be picked, got an empty string")
	}

	// Test with empty list
	files = []string{}
	_, err = pickRandomFile(files)
	if err == nil {
		t.Fatalf("Expected an error, got nil")
	}
}

// Test cropToAspectRatio function
func TestCropToAspectRatio(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 200))
	croppedImg := cropToAspectRatio(img, 50, 50)

	if croppedImg.Bounds().Dx() != 50 || croppedImg.Bounds().Dy() != 50 {
		t.Fatalf("Expected cropped image to be 50x50, got %dx%d", croppedImg.Bounds().Dx(), croppedImg.Bounds().Dy())
	}
}

// Test resizePNGFile function
func TestResizePNGFile(t *testing.T) {
	// Create a temporary directory
	tempDir := os.TempDir()

	// Create a test PNG file
	srcPath := filepath.Join(tempDir, "src.png")
	destPath := filepath.Join(tempDir, "dest.png")

	img := image.NewRGBA(image.Rect(0, 0, 100, 200))
	f, _ := os.Create(srcPath)
	png.Encode(f, img)
	f.Close()

	err := resizePNGFile(srcPath, destPath, 50, 50)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check if the destination file exists
	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Fatalf("Expected destination file to exist, but it does not")
	}

	// Check if the destination file has the correct dimensions
	destFile, _ := os.Open(destPath)
	defer destFile.Close()
	destImg, _, _ := image.Decode(destFile)
	if destImg.Bounds().Dx() != 50 || destImg.Bounds().Dy() != 50 {
		t.Fatalf("Expected resized image to be 50x50, got %dx%d", destImg.Bounds().Dx(), destImg.Bounds().Dy())
	}
}
