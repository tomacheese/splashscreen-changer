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

// listPNGFiles lists .png files under root (recursively if isRecursive, otherwise only directly under root).
func listPNGFiles(root string, isRecursive bool) ([]string, error) {
	var pngFiles []string

	if isRecursive {
		// Walk the directory tree recursively
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Only regular files with a .png extension
			if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".png") {
				pngFiles = append(pngFiles, path)
			}
			return nil
		})

		if err != nil {
			return nil, err
		}
	} else {
		// Only direct children of root
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

// pickRandomFile picks one file at random; it returns an error if files is empty.
func pickRandomFile(files []string) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no PNG files found")
	}

	rand.Seed(uint64(time.Now().UnixNano())) // seed with the current time
	randomIndex := rand.Intn(len(files))
	return files[randomIndex], nil
}

// cropToAspectRatio center-crops img to the aspect ratio of width x height and scales it to width x height.
func cropToAspectRatio(img image.Image, width, height int) image.Image {
	srcBounds := img.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	srcAspectRatio := float64(srcWidth) / float64(srcHeight)
	destAspectRatio := float64(width) / float64(height)

	var cropRect image.Rectangle
	if srcAspectRatio > destAspectRatio {
		// Source is wider: crop left and right
		newWidth := int(destAspectRatio * float64(srcHeight))
		x0 := (srcWidth - newWidth) / 2
		cropRect = image.Rect(x0, 0, x0+newWidth, srcHeight)
	} else {
		// Source is taller: crop top and bottom
		newHeight := int(float64(srcWidth) / destAspectRatio)
		y0 := (srcHeight - newHeight) / 2
		cropRect = image.Rect(0, y0, srcWidth, y0+newHeight)
	}

	// Crop to the computed rectangle
	croppedImg := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(cropRect)

	// Scale the cropped image to the target size
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Rect, croppedImg, croppedImg.Bounds(), draw.Over, nil)

	return dst
}

// resizePNGFile resizes the image at srcPath to width x height and saves it as a PNG at destPath.
// If the aspect ratio differs, the image is center-cropped first.
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

	// Adjust the aspect ratio
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
	// An explicit path (-config flag / CONFIG_PATH env var) is used as-is.
	// Otherwise the default is "data/config.yml" next to the executable,
	// or relative to the current directory under `go run`.
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
	// Parse command-line flags
	helpFlag := flag.Bool("help", false, "Show help message")
	versionFlag := flag.Bool("version", false, "Show version")
	configParamPath := flag.String("config", os.Getenv("CONFIG_PATH"), "Path to the configuration file")
	flag.Parse()

	// Show help
	if *helpFlag {
		printHelp()
		return
	}

	// Show version
	if *versionFlag {
		log.Println("splashscreen-changer")
		log.Println("|- Version", GetAppVersion())
		log.Println("|- Build date:", GetAppDate())
		return
	}

	// Load the config file
	configPath := getConfigPath(configParamPath)

	log.Println("Loading config file:", configPath)
	config, err := LoadConfig(configPath)
	if err != nil {
		log.Println("Failed to load configuration file:", err)
		return
	}

	// Open the log file
	path := getLogFilePath(&config.Log.Path)
	// Create the log file's parent directory if missing
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

	// Log the effective settings
	log.Printf("Source Path: %s\n", sourcePath)
	log.Printf("Source Recursive: %t\n", config.Source.Recursive)
	log.Printf("Destination Path: %s\n", destinationPath)
	log.Printf("Destination Width: %d\n", config.Destination.Width)
	log.Printf("Destination Height: %d\n", config.Destination.Height)

	// List PNG files in the source directory
	files, err := listPNGFiles(sourcePath, config.Source.Recursive)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	if len(files) == 0 {
		log.Println("No PNG files found")
		return
	}

	// Pick one file at random
	pickedFile, err := pickRandomFile(files)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	log.Println("Picked file:", pickedFile)

	// Resize and save into the EasyAntiCheat directory
	destFile := filepath.Join(destinationPath, "EasyAntiCheat", "SplashScreen.png")
	err = resizePNGFile(pickedFile, destFile, config.Destination.Width, config.Destination.Height)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	log.Println("Resized file saved to:", destFile)
}
