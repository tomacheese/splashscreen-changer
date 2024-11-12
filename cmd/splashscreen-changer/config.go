package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Source struct {
		Path      string `yaml:"path" help:"Path to the source directory" required:"true"`
		Recursive bool   `yaml:"recursive" help:"Whether to search for PNG files recursively. Default is false" default:"false"`
	} `yaml:"source" required:"true"`
	Destination struct {
		Path   string `yaml:"path" help:"Path to the destination directory. The specified directory must have an EasyAntiCheat directory" required:"true"`
		Width  int    `yaml:"width" help:"Width of the destination image" default:"800"`
		Height int    `yaml:"height" help:"Height of the destination image" default:"450"`
	} `yaml:"destination" required:"true"`
}

// 設定ファイルを読み込む
func LoadConfig(filename string) (*Config, error) {
	var config Config

	// 設定ファイルが存在する場合のみ読み込む
	if _, err := os.Stat(filename); err == nil {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			return nil, err
		}
	}

	// 環境変数で設定を上書き
	overrideConfigWithEnv(&config)

	// デフォルト値を設定
	setDefaults(&config)

	// 設定ファイルの内容をチェック
	err := CheckConfig(&config)
	return &config, err
}

// 環境変数で設定を動的に取得する
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// 環境変数で設定を動的に取得する
func overrideConfigWithEnv(config *Config) {
	configValue := reflect.ValueOf(config).Elem()
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		section := configValue.Field(i)
		sectionType := section.Type()

		for j := 0; j < section.NumField(); j++ {
			field := section.Field(j)
			fieldType := sectionType.Field(j)
			envKey := strings.ToUpper(configType.Field(i).Name + "_" + fieldType.Name)

			if value := getEnv(envKey, ""); value != "" {
				switch field.Kind() {
				case reflect.String:
					field.SetString(value)
				case reflect.Bool:
					boolValue, _ := strconv.ParseBool(value)
					field.SetBool(boolValue)
				case reflect.Int:
					intValue, _ := strconv.Atoi(value)
					field.SetInt(int64(intValue))
				}
			}
		}
	}
}

// デフォルト値を設定する
func setDefaults(config *Config) {
	configValue := reflect.ValueOf(config).Elem()

	for i := 0; i < configValue.NumField(); i++ {
		section := configValue.Field(i)
		sectionType := section.Type()

		for j := 0; j < section.NumField(); j++ {
			field := section.Field(j)
			fieldType := sectionType.Field(j)

			if defaultValue, exists := fieldType.Tag.Lookup("default"); exists {
				switch field.Kind() {
				case reflect.String:
					if field.String() == "" {
						field.SetString(defaultValue)
					}
				case reflect.Bool:
					if !field.Bool() {
						boolValue, _ := strconv.ParseBool(defaultValue)
						field.SetBool(boolValue)
					}
				case reflect.Int:
					if field.Int() == 0 {
						intValue, _ := strconv.Atoi(defaultValue)
						field.SetInt(int64(intValue))
					}
				}
			}
		}
	}
}

// 設定ファイルの内容をチェックする
func CheckConfig(config *Config) error {
	// 各フィールドが空でないかチェック
	configValue := reflect.ValueOf(config).Elem()
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		section := configValue.Field(i)
		sectionType := section.Type()

		for j := 0; j < section.NumField(); j++ {
			field := section.Field(j)
			fieldType := sectionType.Field(j)
			fieldName := configType.Field(i).Name + "." + fieldType.Name

			// required タグが付いている場合、空の値が設定されていないかチェック
			if required, _ := fieldType.Tag.Lookup("required"); required == "true" {
				switch field.Kind() {
				case reflect.String:
					if field.String() == "" {
						return fmt.Errorf("%s is required", fieldName)
					}
				case reflect.Bool:
					if !field.Bool() {
						return fmt.Errorf("%s is required", fieldName)
					}
				case reflect.Int:
					if field.Int() == 0 {
						return fmt.Errorf("%s is required", fieldName)
					}
				}
			}
		}
	}

	// パスが存在するかチェック
	if _, err := os.Stat(config.Source.Path); err != nil {
		return fmt.Errorf("source path '%s' does not exist", config.Source.Path)
	}
	if _, err := os.Stat(config.Destination.Path); err != nil {
		return fmt.Errorf("destination path '%s' does not exist", config.Destination.Path)
	}

	// destination.path には "EasyAntiCheat" ディレクトリが存在すること
	eacPath := config.Destination.Path + "/EasyAntiCheat"
	if _, err := os.Stat(eacPath); err != nil {
		return fmt.Errorf("EasyAntiCheat directory not found in destination path '%s'", config.Destination.Path)
	}

	// destination.width が 0 より大きいこと
	if config.Destination.Width <= 0 {
		return fmt.Errorf("destination width must be greater than 0")
	}

	// destination.height が 0 より大きいこと
	if config.Destination.Height <= 0 {
		return fmt.Errorf("destination height must be greater than 0")
	}

	return nil
}
