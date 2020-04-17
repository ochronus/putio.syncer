package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/putdotio/go-putio"
	"github.com/sherifabdlnaby/gpool"
	"golang.org/x/oauth2"
)

func ListToplevelFolders(client *putio.Client) (files []putio.File, directories []string, err error) {
	files, _, err = client.Files.List(context.Background(), 0)
	if err != nil {
		return nil, nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			directories = append(directories, file.Name)
		}
	}
	return files, directories, nil
}

func GetPutioClient(token string) *putio.Client {
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	oauthClient := oauth2.NewClient(context.Background(), tokenSource)
	return putio.NewClient(oauthClient)
}

func TraversePutioFolder(root putio.File, parents []string, client *putio.Client, pool *gpool.Pool, resultsChan chan<- string, localFolder string) error {
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
			TraversePutioFolder(file, parents, client, pool, resultsChan, localFolder)
		} else {
			processFile(file, parents, client, pool, resultsChan, localFolder)
		}
	}
	return nil
}

func processFile(file putio.File, parents []string, client *putio.Client, pool *gpool.Pool, resultsChan chan<- string, localFolder string) {
	// TODO: make path handling OS independent
	destinationDir := fmt.Sprintf("%s/%s", localFolder, strings.Join(parents, "/"))
	downloadUrl, err := client.Files.URL(context.Background(), file.ID, true)
	if err != nil {
		fmt.Printf("Cannot get download url for %s\n", file.Name)
	} else {
		job := func() {
			DownloadFile(downloadUrl, destinationDir)
			resultsChan <- file.Name
		}
		err := pool.Enqueue(context.Background(), job)
		if err != nil {
			log.Fatal("Error queueing", err)
		}
	}
}
