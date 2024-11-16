package main

import (
	"os"
	"path/filepath"
	"time"
)

func getLogFilePath(logParamPath *string) string {
	// ログフォルダパスは環境変数 LOG_PATH または引数 -log で指定し、指定されていない場合は "logs/" とする。
	// "logs/" の場所は、実行ファイルと同じディレクトリにあるものとする。go runで実行する場合は、カレントディレクトリにあるものとする。
	// ログファイルのファイル名は、yyyy-mm-dd.log とする。
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
