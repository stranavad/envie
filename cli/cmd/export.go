package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/stranavad/envie/cli/internal/api"
	"github.com/stranavad/envie/cli/internal/crypto"
	"github.com/spf13/cobra"
)

var (
	exportFormat string
	exportOutput string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export project secrets",
	Long: `Export secrets from an Envie project in various formats.

Secrets are fetched from the server and decrypted locally using your CLI identity.
The decryption happens entirely on your machine - the server never sees plaintext values.

Examples:
  # Export as shell commands (for eval)
  eval $(envie export --project my-api)

  # Export as .env file
  envie export --project my-api --format dotenv > .env

  # Export as JSON
  envie export --project my-api --format json

  # Use environment variable for token
  export ENVIE_TOKEN=envie_xxxxx
  envie export --project my-api`,
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "shell", "Output format: shell, dotenv, json")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Write to file instead of stdout")
}

func runExport(cmd *cobra.Command, args []string) error {
	// 1. Get token
	tokenValue, err := getToken()
	if err != nil {
		return err
	}

	// 2. Get project
	projectID, err := getProject()
	if err != nil {
		return err
	}

	// 3. Parse token and derive keys
	identity, err := crypto.ParseToken(tokenValue)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// 4. Create API client and fetch config
	client := api.NewClient(apiURL, identity.IdentityID)
	configResp, err := client.GetProjectConfig(projectID)
	if err != nil {
		return fmt.Errorf("failed to fetch config: %w", err)
	}

	// 5. Decrypt project key using CLI identity's private key
	projectKey, err := crypto.DecryptWithPrivateKeyBase64(identity.PrivateKey, configResp.EncryptedProjectKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt project key: %w", err)
	}

	// 6. Decrypt each config value
	secrets := make(map[string]string)
	for _, item := range configResp.Items {
		decrypted, err := crypto.DecryptConfigValueBase64(projectKey, item.EncryptedValue)
		if err != nil {
			return fmt.Errorf("failed to decrypt '%s': %w", item.Name, err)
		}
		secrets[item.Name] = string(decrypted)
	}

	// 7. Format output
	output, err := formatSecrets(secrets, exportFormat)
	if err != nil {
		return err
	}

	// 8. Write output
	if exportOutput != "" {
		if err := os.WriteFile(exportOutput, []byte(output), 0600); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Wrote %d secrets to %s\n", len(secrets), exportOutput)
	} else {
		fmt.Print(output)
	}

	return nil
}

// formatSecrets formats the secrets map according to the specified format
func formatSecrets(secrets map[string]string, format string) (string, error) {
	// Sort keys for consistent output
	keys := make([]string, 0, len(secrets))
	for k := range secrets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	switch format {
	case "shell":
		return formatShell(keys, secrets), nil
	case "dotenv":
		return formatDotenv(keys, secrets), nil
	case "json":
		return formatJSON(secrets)
	default:
		return "", fmt.Errorf("unknown format: %s (use shell, dotenv, or json)", format)
	}
}

// formatShell formats secrets as shell export commands
func formatShell(keys []string, secrets map[string]string) string {
	var sb strings.Builder
	for _, key := range keys {
		value := secrets[key]
		// Escape single quotes in value
		escaped := strings.ReplaceAll(value, "'", "'\"'\"'")
		sb.WriteString(fmt.Sprintf("export %s='%s'\n", key, escaped))
	}
	return sb.String()
}

// formatDotenv formats secrets as a .env file
func formatDotenv(keys []string, secrets map[string]string) string {
	var sb strings.Builder
	for _, key := range keys {
		value := secrets[key]
		// Quote values that contain special characters
		if needsQuoting(value) {
			// Escape double quotes and backslashes
			escaped := strings.ReplaceAll(value, "\\", "\\\\")
			escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
			escaped = strings.ReplaceAll(escaped, "\n", "\\n")
			sb.WriteString(fmt.Sprintf("%s=\"%s\"\n", key, escaped))
		} else {
			sb.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		}
	}
	return sb.String()
}

// formatJSON formats secrets as JSON
func formatJSON(secrets map[string]string) (string, error) {
	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data) + "\n", nil
}

// needsQuoting returns true if the value needs to be quoted in .env format
func needsQuoting(value string) bool {
	if value == "" {
		return true
	}
	for _, c := range value {
		switch c {
		case ' ', '"', '\'', '\\', '\n', '\r', '\t', '#', '$', '!', '`':
			return true
		}
	}
	return false
}
