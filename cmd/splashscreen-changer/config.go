package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Source struct {
		Path      string `yaml:"path" help:"Path to the source directory. If not specified, the VRChat folder in the user's Pictures folder is searched and used if available. If not, an error is returned."`
		Recursive bool   `yaml:"recursive" help:"Whether to search for PNG files recursively" default:"true"`
	} `yaml:"source" required:"true"`
	Destination struct {
		Path   string `yaml:"path" help:"Path to the destination directory. The specified directory must have an EasyAntiCheat directory. If not specified, the VRChat folder is searched based on the Steam library folder and used if available. If not, an error is returned."`
		Width  int    `yaml:"width" help:"Width of the destination image" default:"800"`
		Height int    `yaml:"height" help:"Height of the destination image" default:"450"`
	} `yaml:"destination" required:"true"`
	Log struct {
		Path string `yaml:"path" help:"Path to the log file" default:""`
	} `yaml:"log"`
}

// LoadConfig loads the config file (if it exists), applies environment variable
// overrides and defaults, then validates the result.
func LoadConfig(filename string) (*Config, error) {
	var config Config

	// Read the file only if it exists
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

	// Override with environment variables
	overrideConfigWithEnv(&config)

	// Apply defaults
	setDefaults(&config)

	// Validate
	err := checkConfig(&config)
	return &config, err
}

// getEnv returns the environment variable value, or defaultValue if it is not set.
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// overrideConfigWithEnv overrides non-empty SECTION_FIELD environment variables onto config, using reflection.
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
					if boolValue, err := strconv.ParseBool(value); err == nil {
						field.SetBool(boolValue)
					} else {
						fmt.Printf("Error parsing bool for %s: %v\n", envKey, err)
					}
				case reflect.Int:
					if intValue, err := strconv.Atoi(value); err == nil {
						field.SetInt(int64(intValue))
					} else {
						fmt.Printf("Error parsing int for %s: %v\n", envKey, err)
					}
				}
			}
		}
	}
}

// setDefaults fills zero-valued fields from their `default` struct tags.
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

// checkConfig validates the loaded config.
func checkConfig(config *Config) error {
	configValue := reflect.ValueOf(config).Elem()
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		section := configValue.Field(i)
		sectionType := section.Type()

		for j := 0; j < section.NumField(); j++ {
			field := section.Field(j)
			fieldType := sectionType.Field(j)
			fieldName := configType.Field(i).Name + "." + fieldType.Name

			// Fields tagged required:"true" must not be empty
			if required, _ := fieldType.Tag.Lookup("required"); required == "true" {
				switch field.Kind() {
				case reflect.String:
					if field.String() == "" {
						return fmt.Errorf("%s is required but was empty", strings.ToLower(fieldName))
					}
				case reflect.Bool:
					if !field.Bool() {
						return fmt.Errorf("%s is required", strings.ToLower(fieldName))
					}
				case reflect.Int:
					if field.Int() == 0 {
						return fmt.Errorf("%s is required", strings.ToLower(fieldName))
					}
				}
			}
		}
	}

	// Paths must exist
	if config.Source.Path != "" {
		if _, err := os.Stat(config.Source.Path); err != nil {
			return fmt.Errorf("source path '%s' does not exist", config.Source.Path)
		}
	}

	if config.Destination.Path != "" {
		if _, err := os.Stat(config.Destination.Path); err != nil {
			return fmt.Errorf("destination path '%s' does not exist", config.Destination.Path)
		}

		// destination.path must contain an "EasyAntiCheat" directory
		eacPath := config.Destination.Path + "/EasyAntiCheat"
		if _, err := os.Stat(eacPath); err != nil {
			return fmt.Errorf("EasyAntiCheat directory not found in destination path '%s'", config.Destination.Path)
		}
	}

	// destination.width must be greater than 0
	if config.Destination.Width <= 0 {
		return fmt.Errorf("destination width must be greater than 0")
	}

	// destination.height must be greater than 0
	if config.Destination.Height <= 0 {
		return fmt.Errorf("destination height must be greater than 0")
	}

	return nil
}
