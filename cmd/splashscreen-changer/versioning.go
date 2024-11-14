package main

import (
	"runtime/debug"
)

var version string
var date string

// GetAppVersion returns the current application version as a string.
// If the version variable is set and starts with 'v', the 'v' is removed before returning the version.
// If the version variable is not set, it attempts to read the build information.
// If the build information's version starts with 'v', the 'v' is removed before returning the version.
// If neither the version variable nor the build information is available, it returns "unknown".
func GetAppVersion() string {
	if version != "" {
		// vから始まる場合は、vを削除して返す
		if len(version) > 0 && version[0] == 'v' {
			return version[1:]
		}
		return version
	}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		// vから始まる場合は、vを削除して返す
		if len(buildInfo.Main.Version) > 0 && buildInfo.Main.Version[0] == 'v' {
			return buildInfo.Main.Version[1:]
		}
		return buildInfo.Main.Version
	}

	return "unknown"
}

// GetAppDate returns the application date if it is set, otherwise it returns "unknown".
func GetAppDate() string {
	if date != "" {
		return date
	}

	return "unknown"
}
