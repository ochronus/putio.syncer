package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"github.com/manifoldco/promptui"
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
		Label: "Select Day",
		Items: directories,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	fmt.Printf("You choose %q\n", result)
}
