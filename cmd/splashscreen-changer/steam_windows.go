//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/andygrunwald/vdf"
	"golang.org/x/sys/windows/registry"
)

func GetSteamInstallFolder() (string, error) {
	// Open the key for reading
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Wow6432Node\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	// Read the value of the key
	installPath, _, err := k.GetStringValue("InstallPath")
	if err != nil {
		return "", err
	}

	return installPath, nil
}

func getSteamLibraryFolders(steamInstallPath string) ([]string, error) {
	// Open the file for reading
	steamLibraryFoldersPath := filepath.Join(steamInstallPath, "steamapps", "libraryfolders.vdf")
	f, err := os.Open(steamLibraryFoldersPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Parse the VDF file
	p := vdf.NewParser(f)
	vdf, err := p.Parse()
	if err != nil {
		return nil, err
	}

	// Get the LibraryFolders section
	libraryFolders, ok := vdf["libraryfolders"]
	if !ok {
		return nil, fmt.Errorf("LibraryFolders not found in %s", steamLibraryFoldersPath)
	}

	// Iterate over the LibraryFolders and get the paths
	paths := []string{}
	for key, value := range libraryFolders.(map[string]any) {
		// The first path is the Steam installation folder
		if key == "0" {
			paths = append(paths, steamInstallPath)
			continue
		}

		// Get the path
		path, ok := value.(map[string]interface{})["path"]
		if !ok {
			return nil, fmt.Errorf("path not found in LibraryFolders[%s]", key)
		}

		// Convert the path to a string
		pathStr, ok := path.(string)
		if !ok {
			return nil, fmt.Errorf("path is not a string in LibraryFolders[%s]", key)
		}

		// Append the path to the list of paths
		paths = append(paths, pathStr)
	}

	// Return the list of paths
	return paths, nil
}

func findSteamGameDirectory(gameName string) (string, error) {
	steamInstallPath, err := GetSteamInstallFolder()
	if err != nil {
		return "", err
	}

	steamLibraryFolders, err := getSteamLibraryFolders(steamInstallPath)
	if err != nil {
		return "", err
	}

	// Iterate over the Steam Library Folders
	for _, steamLibraryFolder := range steamLibraryFolders {
		// The Steam Library Folder contains the game directory
		//  e.g. C:\Program Files (x86)\Steam\steamapps\common\Portal 2
		gameDirectory := filepath.Join(steamLibraryFolder, "steamapps", "common", gameName)

		// Check if the game directory exists
		if _, err := os.Stat(gameDirectory); err == nil {
			return gameDirectory, nil
		}
	}

	// Return an error if the game directory was not found
	return "", fmt.Errorf("game directory not found for %s", gameName)
}
