package main

import (
	"os"
	"path/filepath"
	"time"
)

func getLogFilePath(logParamPath *string) string {
	// An explicit log path (config log.path / LOG_PATH) is used as-is.
	// Otherwise the log file is "logs/yyyy-mm-dd.log" next to the executable,
	// or relative to the current directory under `go run`.
	if *logParamPath != "" {
		return *logParamPath
	}

	date := time.Now().Format("2006-01-02")

	exePath, err := os.Executable()
	if err != nil {
		return filepath.Join("logs", date+".log")
	}

	if isGoRun() {
		return filepath.Join("logs", date+".log")
	}

	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "logs", date+".log")
}
