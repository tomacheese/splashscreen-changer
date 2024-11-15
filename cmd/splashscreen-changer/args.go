package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"unsafe"
)

// ヘルプメッセージを表示する関数
func printHelp() {
	fmt.Println("Usage: splashscreen-changer [options]")
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println("Environment Variables:")
	fmt.Printf("  %-20s %s\n", "CONFIG_PATH", "Path to the configuration file (default: data/config.yaml)")

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
}

// SHGetKnownFolderPathを使うための設定
var (
	modShell32               = syscall.NewLazyDLL("shell32.dll")
	procSHGetKnownFolderPath = modShell32.NewProc("SHGetKnownFolderPath")
	modOle32                 = syscall.NewLazyDLL("ole32.dll")
	procCoTaskMemFree        = modOle32.NewProc("CoTaskMemFree")
	FOLDERID_PicturesLegacy  = syscall.GUID{Data1: 0x0DDD015D, Data2: 0xB06C, Data3: 0x45D5, Data4: [8]byte{0x8C, 0x4C, 0xF5, 0x97, 0x13, 0x85, 0x46, 0x39}}
	FOLDERID_Pictures        = syscall.GUID{Data1: 0x33E28130, Data2: 0x4E1E, Data3: 0x4676, Data4: [8]byte{0x83, 0x5A, 0x98, 0x5A, 0x76, 0x87, 0x67, 0x4D}}
)

// getKnownFolderPath は、指定されたKnown Folderのパスを取得します。
func getKnownFolderPath(folderID *syscall.GUID) (string, error) {
	var pathPtr uintptr
	// SHGetKnownFolderPathを呼び出してフォルダパスを取得
	ret, _, _ := procSHGetKnownFolderPath.Call(
		uintptr(unsafe.Pointer(folderID)),
		0,
		0,
		uintptr(unsafe.Pointer(&pathPtr)),
	)
	if ret != 0 {
		return "", fmt.Errorf("failed to get folder path, error code: %d", ret)
	}
	defer procCoTaskMemFree.Call(pathPtr) // メモリ解放

	// ポインタから文字列を取得
	return syscall.UTF16ToString((*[1 << 16]uint16)(unsafe.Pointer(pathPtr))[:]), nil
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
	picturesLegacyPath, errLegacy := getKnownFolderPath(&FOLDERID_PicturesLegacy)
	picturesNewPath, err := getKnownFolderPath(&FOLDERID_Pictures)
	if errLegacy != nil && err != nil {
		return "", errorRequired
	}

	picturesPath := picturesLegacyPath
	if errLegacy != nil {
		picturesPath = picturesNewPath
	}

	vrchatPath := filepath.Join(picturesPath, "VRChat")
	if _, err := os.Stat(vrchatPath); err == nil {
		return vrchatPath, nil
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
	if err == nil {
		fmt.Println("findSteamGameDirectory", err)
		return vrchatPath, nil
	}

	_, err = os.Stat(filepath.Join(vrchatPath, "EasyAntiCheat"))
	if err != nil {
		return "", nil
	}

	return "", errorRequired
}
