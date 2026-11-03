package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// printHelp prints usage, flags and the environment variables derived from Config.
func printHelp() {
	fmt.Println("Usage: splashscreen-changer [options]")
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println("Environment Variables:")
	fmt.Printf("  %-20s %s\n", "CONFIG_PATH", "Path to the configuration file (default: data/config.yml)")

	// Environment variable keys are SECTION_FIELD, generated from the Config struct fields.
	configType := reflect.TypeOf(Config{})
	for i := 0; i < configType.NumField(); i++ {
		section := configType.Field(i)
		sectionType := section.Type

		for j := 0; j < sectionType.NumField(); j++ {
			field := sectionType.Field(j)
			helpTag := field.Tag.Get("help")
			envKey := strings.ToUpper(section.Name + "_" + field.Name)
			defaultValue := field.Tag.Get("default")
			if defaultValue != "" {
				helpTag += fmt.Sprintf(" (default: %s)", defaultValue)
			}
			fmt.Printf("  %-20s %s\n", envKey, helpTag)
		}
	}

	fmt.Println()
	fmt.Println("GitHub: https://github.com/tomacheese/splashscreen-changer")
	fmt.Println("Booth: https://tomachi.booth.pm/items/6284870")
}

func getSourcePath(config *Config) (string, error) {
	// Resolution order:
	// 1. Environment variable SOURCE_PATH
	// 2. source.path in the config file
	// 3. The VRChat folder inside the user's Pictures folder
	// Otherwise an error is returned.
	//
	// 1 and 2 are already merged into config by config.go, so only step 3 is handled here.
	// Note: in step 3 the VRChat path is returned together with errorRequired even when
	// the folder exists, so the caller treats it as an error (auto-detection is effectively disabled).
	if config.Source.Path != "" {
		return config.Source.Path, nil
	}

	errorRequired := fmt.Errorf("source.path is required")

	// Check whether a VRChat folder exists in the Pictures folder (legacy location first)
	picturesLegacyPath, errLegacy := getPicturesLegacyPath()
	picturesNewPath, err := getPicturesPath()
	if errLegacy != nil && err != nil {
		return "", errorRequired
	}

	picturesPath := picturesLegacyPath
	if errLegacy != nil {
		picturesPath = picturesNewPath
	}

	vrchatPath := filepath.Join(picturesPath, "VRChat")
	if _, err := os.Stat(vrchatPath); err == nil {
		return vrchatPath, errorRequired
	}

	return "", errorRequired
}

func getDestinationPath(config *Config) (string, error) {
	// Resolution order:
	// 1. Environment variable DESTINATION_PATH
	// 2. destination.path in the config file
	// 3. The VRChat install folder found via the Steam library folders
	// Otherwise an error is returned.
	//
	// 1 and 2 are already merged into config by config.go, so only step 3 is handled here.
	// Step 3 requires an EasyAntiCheat folder inside the install folder.
	if config.Destination.Path != "" {
		return config.Destination.Path, nil
	}

	errorRequired := fmt.Errorf("destination.path is required")

	vrchatPath, err := findSteamGameDirectory("VRChat")
	if err != nil {
		return "", errorRequired
	}

	_, err = os.Stat(filepath.Join(vrchatPath, "EasyAntiCheat"))
	if err != nil {
		return "", fmt.Errorf("EasyAntiCheat folder not found in %s", vrchatPath)
	}

	return vrchatPath, nil
}
