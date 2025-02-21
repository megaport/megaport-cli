package cmd

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"syscall"

	megaport "github.com/megaport/megaportgo"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/term"
)

// Configuration file path
var (
	configFile = filepath.Join(os.Getenv("HOME"), ".megaport-cli-config.json")
)

// Config struct stores environment and encrypted credentials
type Config struct {
	Environment     string `json:"environment"`
	EncryptedAccess string `json:"encrypted_access"`
	EncryptedSecret string `json:"encrypted_secret"`
	Salt            string `json:"salt"`
}

// Constants for encryption
const (
	keyLen     = 32
	saltLen    = 32
	iterations = 10000
)

// Encrypt credentials using password-derived key
func encrypt(password string, salt []byte, data []byte) ([]byte, error) {
	key := pbkdf2.Key([]byte(password), salt, iterations, keyLen, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// Decrypt credentials using password-derived key
func decrypt(password string, salt []byte, encrypted []byte) ([]byte, error) {
	key := pbkdf2.Key([]byte(password), salt, iterations, keyLen, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// writeConfigFile saves the config struct to disk
func writeConfigFile(cfg Config) error {
	f, err := os.Create(configFile)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(&cfg)
}

// loadConfig obtains the config from the config file
func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// PasswordReader interface for reading passwords
type PasswordReader interface {
	ReadPassword() (string, error)
}

// TerminalPasswordReader reads password from terminal
type TerminalPasswordReader struct{}

func (t *TerminalPasswordReader) ReadPassword() (string, error) {
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	return string(bytePassword), nil
}

// TestPasswordReader is used for testing password input
type TestPasswordReader struct {
	input string
}

func (t *TestPasswordReader) ReadPassword() (string, error) {
	return t.input, nil
}

// Global password reader, can be overridden for testing
var passwordReader PasswordReader = &TerminalPasswordReader{}

// promptPassword securely prompts the user for a password
func promptPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	password, err := passwordReader.ReadPassword()
	fmt.Println()
	return password, err
}

// Login mocks client auth (actual API calls are not performed in tests)
func Login(ctx context.Context) (*megaport.Client, error) {
	httpClient := &http.Client{}
	cfg, err := loadConfig()
	if err != nil || cfg.EncryptedAccess == "" || cfg.EncryptedSecret == "" {
		fmt.Println("Please provide access key and secret key using environment variables or the configure command")
		return nil, fmt.Errorf("access key and secret key are required")
	}

	password, err := promptPassword("Enter password to decrypt credentials: ")
	if err != nil {
		return nil, err
	}

	salt, err := base64.StdEncoding.DecodeString(cfg.Salt)
	if err != nil {
		return nil, err
	}

	encAccess, err := base64.StdEncoding.DecodeString(cfg.EncryptedAccess)
	if err != nil {
		return nil, err
	}

	encSecret, err := base64.StdEncoding.DecodeString(cfg.EncryptedSecret)
	if err != nil {
		return nil, err
	}

	accessKey, err := decrypt(password, salt, encAccess)
	if err != nil {
		return nil, err
	}

	secretKey, err := decrypt(password, salt, encSecret)
	if err != nil {
		return nil, err
	}

	var envOpt megaport.ClientOpt
	switch cfg.Environment {
	case "production":
		envOpt = megaport.WithEnvironment(megaport.EnvironmentProduction)
	case "staging":
		envOpt = megaport.WithEnvironment(megaport.EnvironmentStaging)
	case "development":
		envOpt = megaport.WithEnvironment(megaport.EnvironmentDevelopment)
	default:
		return nil, fmt.Errorf("unknown environment: %s", cfg.Environment)
	}

	megaportClient, err := megaport.New(httpClient, megaport.WithCredentials(string(accessKey), string(secretKey)), envOpt)
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
  MEGAPORT_ACCESS_KEY, MEGAPORT_SECRET_KEY, and MEGAPORT_ENVIRONMENT

Or through command line flags:
  --access-key, --secret-key, and --environment

You will be prompted to enter a password to encrypt your credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get environment from flag or default
		flagEnv, _ := cmd.Flags().GetString("environment")
		if flagEnv == "" {
			flagEnv = "staging"
		}

		// Validate environment
		switch flagEnv {
		case "production", "staging", "development":
			// valid
		default:
			return fmt.Errorf("invalid environment: %s (must be production, staging, or development)", flagEnv)
		}

		// Get access key and secret key from flags
		flagAccessKey, err := cmd.Flags().GetString("access-key")
		if err != nil {
			return fmt.Errorf("error getting access-key flag: %w", err)
		}
		flagSecretKey, err := cmd.Flags().GetString("secret-key")
		if err != nil {
			return fmt.Errorf("error getting secret-key flag: %w", err)
		}

		// If credentials from flags are missing, print message and return error
		if flagAccessKey == "" || flagSecretKey == "" {
			fmt.Println("Please provide credentials either through environment variables MEGAPORT_ACCESS_KEY, MEGAPORT_SECRET_KEY\nor through flags --access-key, --secret-key")
			return fmt.Errorf("no valid credentials provided")
		}

		// Prompt for password
		password, err := promptPassword("Enter password to encrypt credentials: ")
		if err != nil {
			return err
		}

		// Generate salt
		salt := make([]byte, saltLen)
		if _, err := rand.Read(salt); err != nil {
			return fmt.Errorf("error generating salt: %w", err)
		}

		// Encrypt credentials
		encAccess, err := encrypt(password, salt, []byte(flagAccessKey))
		if err != nil {
			return fmt.Errorf("error encrypting access key: %w", err)
		}

		encSecret, err := encrypt(password, salt, []byte(flagSecretKey))
		if err != nil {
			return fmt.Errorf("error encrypting secret key: %w", err)
		}

		cfg := Config{
			Environment:     flagEnv,
			EncryptedAccess: base64.StdEncoding.EncodeToString(encAccess),
			EncryptedSecret: base64.StdEncoding.EncodeToString(encSecret),
			Salt:            base64.StdEncoding.EncodeToString(salt),
		}

		if err := writeConfigFile(cfg); err != nil {
			return fmt.Errorf("error writing config: %v", err)
		}

		fmt.Printf("Environment (%s) saved successfully.\n", flagEnv)
		return nil
	},
}

func init() {
	configureCmd.Flags().String("access-key", "", "Your Megaport access key")
	configureCmd.Flags().String("secret-key", "", "Your Megaport secret key")
	configureCmd.Flags().String("environment", "staging", "API environment (production, staging, or development)")
	rootCmd.AddCommand(configureCmd)
}
