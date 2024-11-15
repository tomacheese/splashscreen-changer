//go:build windows
// +build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

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

func getPicturesLegacyPath() (string, error) {
	return getKnownFolderPath(&FOLDERID_PicturesLegacy)
}

func getPicturesPath() (string, error) {
	return getKnownFolderPath(&FOLDERID_Pictures)
}
