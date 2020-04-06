package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

func LoadConfig() error {
	homeDirPath, _ := os.UserHomeDir()
	configPath := filepath.FromSlash(fmt.Sprintf("%s/.putio.syncer", homeDirPath))
	configFileName := "config"
	configFileType := "yaml"
	configFilePath := filepath.FromSlash(fmt.Sprintf("%s/%s.%s", configPath, configFileName, configFileType))
	mkdirErr := os.MkdirAll(configPath, os.ModePerm)
	if mkdirErr != nil {
		fmt.Printf("Error creating path: %v\n", mkdirErr)
	}
	_, fileErr := os.OpenFile(configFilePath, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if fileErr != nil {
		fmt.Printf("Error 'touching' file %s: %v\n", configFilePath, fileErr)
	}
	viper.SetConfigName(configFileName) // name of config file (without extension)
	viper.SetConfigType(configFileType) // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(configPath)     // call multiple times to add many search paths
	err := viper.ReadInConfig()         // Find and read the config file
	return err
}
