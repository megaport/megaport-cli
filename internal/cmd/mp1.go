package cmd

import (
	"context"
	"fmt"
	"net/http"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

var (
	accessKey string
	secretKey string
)

func Login(ctx context.Context) (*megaport.Client, error) {
	httpClient := &http.Client{}

	if accessKey == "" || secretKey == "" {
		fmt.Println("Please provide access key and secret key using the configure command")
		return nil, fmt.Errorf("access key and secret key are required")
	}

	megaportClient, err := megaport.New(httpClient, megaport.WithCredentials(accessKey, secretKey))
	if err != nil {
		return nil, err
	}
	_, err = megaportClient.Authorize(ctx)
	if err != nil {
		return nil, err
	}
	return megaportClient, nil
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure the CLI with your credentials",
	Run: func(cmd *cobra.Command, args []string) {
		accessKey, _ = cmd.Flags().GetString("access-key")
		secretKey, _ = cmd.Flags().GetString("secret-key")
		fmt.Println("Configuration saved")
	},
}

func init() {
	configureCmd.Flags().StringVar(&accessKey, "access-key", "", "Your Megaport access key")
	configureCmd.Flags().StringVar(&secretKey, "secret-key", "", "Your Megaport secret key")
	rootCmd.AddCommand(configureCmd)
}
