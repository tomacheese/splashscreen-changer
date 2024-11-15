//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"runtime"
)

func getPicturesLegacyPath() (string, error) {
	return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func getPicturesPath() (string, error) {
	return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}