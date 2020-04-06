package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cavaliercoder/grab"
)

func DownloadFile(url string, destinationDir string) {
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
