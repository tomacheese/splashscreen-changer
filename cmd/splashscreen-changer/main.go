package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
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

// PNGファイルリストからランダムに1つ選択する関数
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

// PNGファイルをリサイズする関数
// 同じアスペクト比でない場合、元の画像を中心を基準に切り取る
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
			fmt.Printf("  %-20s %s\n", envKey, helpTag)
		}
	}

	fmt.Println()
	fmt.Println("GitHub: https://github.com/tomacheese/splashscreen-changer")
}

func main() {
	// コマンドライン引数を解析する
	helpFlag := flag.Bool("help", false, "Show help message")
	versionFlag := flag.Bool("version", false, "Show version")
	configPath := flag.String("config", os.Getenv("CONFIG_PATH"), "Path to the configuration file")
	flag.Parse()

	// ヘルプメッセージを表示する
	if *helpFlag {
		printHelp()
		return
	}

	// バージョン情報を表示する
	if *versionFlag {
		fmt.Println("splashscreen-changer")
		fmt.Println("|- Version", GetAppVersion())
		fmt.Println("|- Build date:", GetAppDate())
		return
	}

	// 設定ファイルを読み込む。設定ファイルパスは環境変数 CONFIG_PATH または引数 -config で指定し、指定されていない場合は "data/config.yaml" とする。
	if *configPath == "" {
		*configPath = "data/config.yaml"
	}
	fmt.Println("Loading config file:", *configPath)
	config, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 設定値を表示する
	fmt.Printf("Source Path: %s\n", config.Source.Path)
	fmt.Printf("Source Recursive: %t\n", config.Source.Recursive)
	fmt.Printf("Destination Path: %s\n", config.Destination.Path)
	fmt.Printf("Destination Width: %d\n", config.Destination.Width)
	fmt.Printf("Destination Height: %d\n", config.Destination.Height)

	// ソースディレクトリ以下のPNGファイルをリストする
	files, err := listPNGFiles(config.Source.Path, config.Source.Recursive)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// ランダムで1つのファイルを選択する
	if len(files) == 0 {
		fmt.Println("No PNG files found")
		return
	}

	pickedFile, err := pickRandomFile(files)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Picked file:", pickedFile)

	// ファイルをリサイズして EasyAntiCheat ディレクトリに保存する
	destFile := filepath.Join(config.Destination.Path, "EasyAntiCheat", "SplashScreen.png")
	err = resizePNGFile(pickedFile, destFile, config.Destination.Width, config.Destination.Height)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Resized file saved to:", destFile)
}
