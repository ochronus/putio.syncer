package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/putdotio/go-putio"
	"github.com/spf13/viper"
)

const DOWNLOAD_CONCURRENCY = 2

func loadConfig() error {
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

func getTokenFromUser() string {
	tokenPrompt := promptui.Prompt{
		Label: "Please enter your put.io token",
	}

	result, err := tokenPrompt.Run()

	if err != nil {
		log.Fatal("Prompt failed", err)
	}
	viper.Set("PUTIO_TOKEN", result)
	err = viper.WriteConfig()
	if err != nil {
		log.Fatal("Error saving config", err)
	}
	return result
}

func getRemoteFolderFromUser(client *putio.Client) (remoteFolderId int64) {
	files, directories, err := ListToplevelFolders(client)
	if err != nil {
		log.Fatal("Cannot list top level folders", err)
	}
	prompt := promptui.Select{
		Label: "Select the put.io folder for sync",
		Items: directories,
		Size:  20,
	}

	_, result, promptErr := prompt.Run()

	if promptErr != nil {
		fmt.Printf("Prompt failed %v\n", promptErr)
		return
	}
	for _, file := range files {
		if file.Name == result {
			viper.Set("PUTIO_REMOTE_FOLDER_ID", file.ID)
			remoteFolderId = file.ID
			err = viper.WriteConfig()
			if err != nil {
				log.Fatal("Error saving config", err)
			}
		}
	}
	return remoteFolderId
}

func Setup() (client *putio.Client, remoteFolderId int64) {
	configLoadErr := loadConfig()
	if configLoadErr != nil {
		log.Fatal("Cannot load config")
	}
	token := viper.GetString("PUTIO_TOKEN")
	if token == "" {
		token = getTokenFromUser()
	}

	client = GetPutioClient(token)

	remoteFolderId = viper.GetInt64("PUTIO_REMOTE_FOLDER_ID")
	if remoteFolderId == 0 {
		remoteFolderId = getRemoteFolderFromUser(client)
	}
	return client, remoteFolderId
}
