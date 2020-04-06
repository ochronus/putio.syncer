package main

import (
	"context"
	"fmt"

	"github.com/sherifabdlnaby/gpool"
)

func main() {
	client, remoteFolderId := Setup()
	pool := gpool.NewPool(DOWNLOAD_CONCURRENCY)
	defer pool.Stop()
	resultsChan := make(chan string)
	go func() {
		for url := range resultsChan {
			fmt.Println("Done downloading", url)
		}
	}()
	root, _ := client.Files.Get(context.Background(), remoteFolderId)
	var parents []string
	TraversePutioFolder(root, parents, client, pool, resultsChan)

}
