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
	configFile = filepath.Join(os.Getenv("HOME"), ".megaport-cli-config.json")
)

type Config struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
}

// loadConfig obtains the config from environment variables or the config file.
func loadConfig() (*Config, error) {
	envAccessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	envSecretKey := os.Getenv("MEGAPORT_SECRET_KEY")
	if envAccessKey != "" && envSecretKey != "" {
		return &Config{
			AccessKey: envAccessKey,
			SecretKey: envSecretKey,
		}, nil
	}

	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Login mocks client auth (actual API calls are not performed in tests).
func Login(ctx context.Context) (*megaport.Client, error) {
	httpClient := &http.Client{}
	cfg, err := loadConfig()
	if err != nil || cfg.AccessKey == "" || cfg.SecretKey == "" {
		fmt.Println("Please provide access key and secret key using environment variables or the configure command")
		return nil, fmt.Errorf("access key and secret key are required")
	}

	megaportClient, err := megaport.New(httpClient, megaport.WithCredentials(cfg.AccessKey, cfg.SecretKey))
	if err != nil {
		return nil, err
	}
	if _, err := megaportClient.Authorize(ctx); err != nil {
		return nil, err
	}
	return megaportClient, nil
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure the CLI with your credentials",
	Long: `Configure the CLI with your Megaport API credentials.

You can provide credentials either through environment variables:
  MEGAPORT_ACCESS_KEY and MEGAPORT_SECRET_KEY

Or through command line flags:
  --access-key and --secret-key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Try environment variables first
		envAccessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
		envSecretKey := os.Getenv("MEGAPORT_SECRET_KEY")

		// If environment variables are present, use them
		if envAccessKey != "" && envSecretKey != "" {
			config := Config{
				AccessKey: envAccessKey,
				SecretKey: envSecretKey,
			}
			if err := writeConfigFile(config); err != nil {
				return fmt.Errorf("error writing config from environment: %v", err)
			}
			fmt.Println("Credentials from environment saved successfully.")
			return nil
		}

		// If no environment variables, check flags
		flagAccessKey, err := cmd.Flags().GetString("access-key")
		if err != nil {
			return fmt.Errorf("error getting access-key flag: %w", err)
		}
		flagSecretKey, err := cmd.Flags().GetString("secret-key")
		if err != nil {
			return fmt.Errorf("error getting secret-key flag: %w", err)
		}

		// If flags are missing, return an error
		if flagAccessKey == "" || flagSecretKey == "" {
			fmt.Println("Please provide credentials either through environment variables MEGAPORT_ACCESS_KEY and MEGAPORT_SECRET_KEY\nor through flags --access-key and --secret-key")
			return fmt.Errorf("no valid credentials provided")
		}

		// If flags are present, use them
		config := Config{
			AccessKey: flagAccessKey,
			SecretKey: flagSecretKey,
		}

		if err := writeConfigFile(config); err != nil {
			return fmt.Errorf("error writing config from flags: %v", err)
		}

		fmt.Println("Credentials from flags saved successfully.")
		return nil
	},
}

// writeConfigFile saves the config struct to disk as JSON.
func writeConfigFile(cfg Config) error {
	f, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(&cfg); err != nil {
		return err
	}
	return nil
}

func init() {
	configureCmd.Flags().String("access-key", "", "Your Megaport access key")
	configureCmd.Flags().String("secret-key", "", "Your Megaport secret key")
	rootCmd.AddCommand(configureCmd)
}
