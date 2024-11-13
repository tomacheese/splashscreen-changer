package main

import (
	"runtime/debug"
)

var version string
var commit string
var date string

func GetAppVersion() string {
	if version != "" {
		return version
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		return buildInfo.Main.Version
	}

	return "unknown"
}

func GetAppCommit() string {
	if commit != "" {
		return commit
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range buildInfo.Settings { if setting.Key == "vcs.revision" { return setting.Value } }
	}

	return "unknown"
}

func GetAppDate() string {
	if date != "" {
		return date
	}

	return "unknown"
}
