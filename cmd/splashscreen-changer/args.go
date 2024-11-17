package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// ヘルプメッセージを表示する関数
func printHelp() {
	fmt.Println("Usage: splashscreen-changer [options]")
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println("Environment Variables:")
	fmt.Printf("  %-20s %s\n", "CONFIG_PATH", "Path to the configuration file (default: data/config.yml)")

	// Config 構造体のフィールドから環境変数のキーを生成して表示
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
	// 取得の優先度は以下。
	// 1. 環境変数 SOURCE_PATH
	// 2. 設定ファイル source.path
	// 3. ユーザーフォルダの Pictures フォルダ内、VRChat フォルダ
	// エラー。

	// 1, 2 については、config.go にて実装済み。空値できた場合のみ、3 を行う
	// 3 については、ユーザーフォルダの Pictures フォルダ内に VRChat フォルダが存在するか確認し、存在する場合はそのパスを返す
	if config.Source.Path != "" {
		return config.Source.Path, nil
	}

	errorRequired := fmt.Errorf("source.path is required")

	// ユーザーフォルダの Pictures フォルダ内に VRChat フォルダが存在するか確認
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
	// 取得の優先度は以下。
	// 1. 環境変数 DESTINATION_PATH
	// 2. 設定ファイル destination.path
	// 3. Steam ライブラリフォルダから、VRChat のインストール先を取得
	// エラー。

	// 1, 2 については、config.go にて実装済み。空値できた場合のみ、3 を行う
	// 3 については、Steam ライブラリフォルダから、VRChat のインストール先フォルダを返す（EasyAntiCheatフォルダがあることを確認する）。見つからない場合はエラーを返す
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
