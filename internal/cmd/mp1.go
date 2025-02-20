package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
)

var (
	accessKey  string
	secretKey  string
	configFile = filepath.Join(os.Getenv("HOME"), ".megaport-cli-config.json")
)

type Config struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

func saveConfig(config Config) error {
	file, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(config)
}

func loadConfig() (Config, error) {
	var config Config
	file, err := os.Open(configFile)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	return config, err
}

func Login(ctx context.Context) (*megaport.Client, error) {
	httpClient := &http.Client{}

	config, err := loadConfig()
	if err != nil {
		fmt.Println("Please provide access key and secret key using the configure command")
		return nil, fmt.Errorf("access key and secret key are required")
	}

	megaportClient, err := megaport.New(httpClient, megaport.WithCredentials(config.AccessKey, config.SecretKey))
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

		config := Config{
			AccessKey: accessKey,
			SecretKey: secretKey,
		}

		// If either key is empty, print an error
		if config.AccessKey == "" || config.SecretKey == "" {
			fmt.Println("Error saving configuration: both access key and secret key are required")
			return
		}

		if err := saveConfig(config); err != nil {
			fmt.Println("Error saving configuration:", err)
			return
		}

		fmt.Println("Configuration saved")
	},
}

var showConfigCmd = &cobra.Command{
	Use:   "show-config",
	Short: "Show the current access key and secret key",
	Run: func(cmd *cobra.Command, args []string) {
		config, err := loadConfig()
		if err != nil {
			fmt.Println("Error loading configuration:", err)
			return
		}

		fmt.Printf("Access Key: %s\n", config.AccessKey)
		fmt.Printf("Secret Key: %s\n", config.SecretKey)
	},
}

func init() {
	configureCmd.Flags().StringVar(&accessKey, "access-key", "", "Your Megaport access key")
	configureCmd.Flags().StringVar(&secretKey, "secret-key", "", "Your Megaport secret key")
	rootCmd.AddCommand(configureCmd)
	rootCmd.AddCommand(showConfigCmd)
}
