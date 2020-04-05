package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/putdotio/go-putio"
	"golang.org/x/oauth2"
)

func main() {
	_ = godotenv.Load()
	token := os.Getenv("PUTIO_TOKEN")
	if len(token) == 0 {
		panic("PUTIO_TOKEN is empty")
	}
	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	oauthClient := oauth2.NewClient(context.TODO(), tokenSource)
	client := putio.NewClient(oauthClient)

	// get root directory
	root, err := client.Files.Get(context.TODO(), 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name of root folder is: %s\n", root.Name)
}
