package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/exp/rand"
	"golang.org/x/image/draw"
)

// 指定されたディレクトリ以下のすべてのPNGファイルをリストする関数
func listPNGFiles(root string, isRecursive bool) ([]string, error) {
	var pngFiles []string

	if isRecursive {
		// Walk関数で指定されたディレクトリ以下を再帰的に探索
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// ファイルで拡張子が.pngのものだけをリストに追加
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".png") {
				pngFiles = append(pngFiles, path)
			}
			return nil
		})

		if err != nil {
			return nil, err
		}
	} else {
		// 指定されたディレクトリの直下のみを探索
		files, err := os.ReadDir(root)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".png") {
				pngFiles = append(pngFiles, filepath.Join(root, file.Name()))
			}
		}
	}

	return pngFiles, nil
}

// PNGファイルリストからラダムに1つ選択する関数
func pickRandomFile(files []string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no PNG files found")
	}

	rand.Seed(uint64(time.Now().UnixNano())) // 現在時刻をシードにして乱数を初期化
	randomIndex := rand.Intn(len(files))
	return files[randomIndex], nil
}

// 画像を指定されたアスペクト比に切り取る関数
func cropToAspectRatio(img image.Image, width, height int) image.Image {
	srcBounds := img.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	srcAspectRatio := float64(srcWidth) / float64(srcHeight)
	destAspectRatio := float64(width) / float64(height)

	var cropRect image.Rectangle
	if srcAspectRatio > destAspectRatio {
		// 横長の場合、左右を切り取る
		newWidth := int(destAspectRatio * float64(srcHeight))
		x0 := (srcWidth - newWidth) / 2
		cropRect = image.Rect(x0, 0, x0+newWidth, srcHeight)
	} else {
		// 縦長の場合、上下を切り取る
		newHeight := int(float64(srcWidth) / destAspectRatio)
		y0 := (srcHeight - newHeight) / 2
		cropRect = image.Rect(0, y0, srcWidth, y0+newHeight)
	}

	// 指定された範囲を切り取る
	croppedImg := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(cropRect)

	// 切り取った画像を指定のサイズにリサイズ
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Rect, croppedImg, croppedImg.Bounds(), draw.Over, nil)

	return dst
}

// resizePNGFileは、指定されたPNG画像を指定の幅と高さにリサイズし、保存します。
// リサイズの際、元の画像のアスペクト比が異なる場合は、中央を基準にクロップ（切り取り）します。
// - srcPath: 元のPNGファイルのパス
// - destPath: リサイズ後のPNGファイルの保存先パス
// - width: リサイズ後の画像の幅
// - height: リサイズ後の画像の高さ
func resizePNGFile(srcPath, destPath string, width, height int) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcImage, _, err := image.Decode(srcFile)
	if err != nil {
		return err
	}

	// アスペクト比を調整
	srcImage = cropToAspectRatio(srcImage, width, height)

	destImage := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(destImage, destImage.Rect, srcImage, srcImage.Bounds(), draw.Over, nil)

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	err = png.Encode(destFile, destImage)
	if err != nil {
		return err
	}

	return nil
}

func isGoRun() bool {
	executable, err := os.Executable()
	if err != nil {
		return false
	}

	goTmpDir := os.Getenv("GOTMPDIR")
	if goTmpDir != "" {
		return strings.HasPrefix(executable, goTmpDir)
	}

	return strings.HasPrefix(executable, os.TempDir())
}

func getConfigPath(configParamPath *string) string {
	// 設定ファイルパスは環境変数 CONFIG_PATH または引数 -config で指定し、指定されていない場合は "data/config.yml" とする。
	// "data/config.yml" の場所は、実行ファイルと同じディレクトリにあるものとする。go runで実行する場合は、カレントディレクトリにあるものとする。
	if *configParamPath != "" {
		return *configParamPath
	}

	exePath, err := os.Executable()
	if err != nil {
		return filepath.Join("data", "config.yml")
	}

	if isGoRun() {
		return filepath.Join("data", "config.yml")
	}

	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, "data", "config.yml")
}

func main() {
	// コマンドライン引数を解析する
	helpFlag := flag.Bool("help", false, "Show help message")
	versionFlag := flag.Bool("version", false, "Show version")
	configParamPath := flag.String("config", os.Getenv("CONFIG_PATH"), "Path to the configuration file")
	flag.Parse()

	// ヘルプメッセージを表示する
	if *helpFlag {
		printHelp()
		return
	}

	// バージョン情報を表示する
	if *versionFlag {
		log.Println("splashscreen-changer")
		log.Println("|- Version", GetAppVersion())
		log.Println("|- Build date:", GetAppDate())
		return
	}

	// 設定ファイルを読み込む。
	configPath := getConfigPath(configParamPath)

	log.Println("Loading config file:", configPath)
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Println("Failed to load configuration file:", err)
		return
	}

	// ログファイルを開く
	path := getLogFilePath(&config.Log.Path)
	// ログファイルの親ディレクトリが存在しない場合は作成する
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		log.Fatal(err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	mw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(mw)

	sourcePath, err := getSourcePath(config)
	if err != nil {
		log.Println("Failed to obtain source path")
		log.Println()
		log.Println("The following steps are used to obtain the source paths. This error occurs because the following steps could not be taken to obtain the source path.")
		log.Println("1. Environment variable SOURCE_PATH. If this is not set, the following steps are taken.")
		log.Println("2. source.path in Configuration file. If this is not set, the following steps are taken.")
		log.Println("3. Check if the VRChat folder exists in the Pictures folder in the user folder.")
		log.Println("If the VRChat folder exists, the path to the VRChat folder is used as the source path.")
		return
	}

	destinationPath, err := getDestinationPath(config)
	if err != nil {
		log.Println("Failed to obtain destination path")
		log.Println()
		log.Println("The following steps are used to obtain the destination paths. This error occurs because the following steps could not be taken to obtain the destination path.")
		log.Println("1. Environment variable DESTINATION_PATH. If this is not set, the following steps are taken.")
		log.Println("2. destination.path in Configuration file. If this is not set, the following steps are taken.")
		log.Println("3. Get the installation destination folder of VRChat from the Steam library folder.")
		log.Println("If the EasyAntiCheat folder exists in the VRChat folder, the path to the VRChat folder is used as the destination path.")
		return
	}

	// 設定値を表示する
	log.Printf("Source Path: %s\n", sourcePath)
	log.Printf("Source Recursive: %t\n", config.Source.Recursive)
	log.Printf("Destination Path: %s\n", destinationPath)
	log.Printf("Destination Width: %d\n", config.Destination.Width)
	log.Printf("Destination Height: %d\n", config.Destination.Height)

	// ソースディレクトリ以下のPNGファイルをリストする
	files, err := listPNGFiles(sourcePath, config.Source.Recursive)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	// ランダムで1つのファイルを選択する
	if len(files) == 0 {
		log.Println("No PNG files found")
		return
	}

	pickedFile, err := pickRandomFile(files)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	log.Println("Picked file:", pickedFile)

	// ファイルをリサイズして EasyAntiCheat ディレクトリに保存する
	destFile := filepath.Join(destinationPath, "EasyAntiCheat", "SplashScreen.png")
	err = resizePNGFile(pickedFile, destFile, config.Destination.Width, config.Destination.Height)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	log.Println("Resized file saved to:", destFile)
}
