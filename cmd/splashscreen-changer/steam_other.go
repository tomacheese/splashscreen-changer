//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"runtime"
)

func GetSteamInstallFolder() (string, error) {
	return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}

func findSteamGameDirectory(_ string) (string, error) {
	return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
}
