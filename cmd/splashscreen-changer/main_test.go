package main

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

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
