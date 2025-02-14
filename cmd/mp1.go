package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	megaport "github.com/megaport/megaportgo"
)

func Login() (*megaport.Client, error) {
	httpClient := &http.Client{}

	accessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey := os.Getenv("MEGAPORT_SECRET_KEY")
	if accessKey == "" || secretKey == "" {
		fmt.Println("Please set MEGAPORT_ACCESS_KEY and MEGAPORT_SECRET_KEY environment variables")
		os.Exit(1)
	}
	var err error
	megaportClient, err := megaport.New(httpClient, megaport.WithCredentials(accessKey, secretKey))
	if err != nil {
		return nil, err
	}
	_, err = megaportClient.Authorize(context.Background())
	if err != nil {
		return nil, err
	}
	return megaportClient, nil
}
