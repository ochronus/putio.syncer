package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cavaliercoder/grab"
	"github.com/manifoldco/promptui"
	"github.com/putdotio/go-putio"
	"github.com/sherifabdlnaby/gpool"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type PutioWalkFunc func(file putio.File, parents []string, client *putio.Client, pool *gpool.Pool, resultsChan chan<- string)

func walkPutIo(root putio.File, parents []string, client *putio.Client, walkFn PutioWalkFunc, pool *gpool.Pool, resultsChan chan<- string) error {
	if !root.IsDir() {
		return nil
	}
	parents = append(parents, root.Name)
	files, _, err := client.Files.List(context.Background(), root.ID)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			walkPutIo(file, parents, client, walkFn, pool, resultsChan)
		} else {
			walkFn(file, parents, client, pool, resultsChan)
		}
	}
	return nil
}

func worker(url string, destinationDir string) {
	mkdirErr := os.MkdirAll(destinationDir, os.ModePerm)
	if mkdirErr != nil {
		fmt.Printf("Error creating path: %v\n", mkdirErr)
	}
	fmt.Println("Downloading", url, "to", destinationDir)
	_, err := grab.Get(destinationDir, url)
	if err != nil {
		log.Fatal(err)
	}
}

func printFile(file putio.File, parents []string, client *putio.Client, pool *gpool.Pool, resultsChan chan<- string) {
	currDir, err := os.Getwd()
	if err != nil {
		panic("Cannot get current directory")
	}
	destinationDir := fmt.Sprintf("%s/%s", currDir, strings.Join(parents, "/"))
	downloadUrl, err := client.Files.URL(context.Background(), file.ID, true)
	if err != nil {
		fmt.Printf("Cannot get download url for %s\n", file.Name)
	} else {
		job := func() {
			worker(downloadUrl, destinationDir)
			resultsChan <- file.Name
		}
		err := pool.Enqueue(context.Background(), job)
		if err != nil {
			fmt.Printf("Error queueing: %v", err)
		}
	}
}

func main() {
	configLoadErr := LoadConfig()
	if configLoadErr != nil {
		panic("Cannot load config")
	}
	token := viper.GetString("PUTIO_TOKEN")
	if token == "" {
		tokenPrompt := promptui.Prompt{
			Label: "Please input your put.io token",
		}

		result, err := tokenPrompt.Run()

		if err != nil {
			panic(fmt.Sprintf("Prompt failed: %v", err))
		}
		viper.Set("PUTIO_TOKEN", result)
		token = result
		err = viper.WriteConfig()
		if err != nil {
			panic(fmt.Sprintf("Error saving config: %v", err))
		}

	}

	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	oauthClient := oauth2.NewClient(context.TODO(), tokenSource)
	client := putio.NewClient(oauthClient)

	remoteFolderId := viper.GetInt64("PUTIO_REMOTE_FOLDER_ID")
	if remoteFolderId == 0 {
		// get root directory
		files, _, err := client.Files.List(context.TODO(), 0)
		if err != nil {
			log.Fatal(err)
		}
		var directories []string
		for _, file := range files {
			if file.IsDir() {
				directories = append(directories, file.Name)
			}
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
					panic(fmt.Sprintf("Error saving config: %v", err))
				}
			}
		}
	}
	pool := gpool.NewPool(2)
	defer pool.Stop()
	resultsChan := make(chan string)
	go func() {
		for url := range resultsChan {
			fmt.Println(url)
		}
	}()
	root, _ := client.Files.Get(context.TODO(), remoteFolderId)
	var parents []string
	walkPutIo(root, parents, client, printFile, pool, resultsChan)

}
