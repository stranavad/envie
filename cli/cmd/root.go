package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	token   string
	project string
	apiURL  string

	// Version info (set at build time via ldflags)
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "envie",
	Short: "Envie CLI - Secure secrets management",
	Long: `Envie CLI allows you to securely fetch and manage environment variables
from your Envie projects. Secrets are end-to-end encrypted and decrypted
locally using your CLI identity token.

Usage in CI/CD:
  export ENVIE_TOKEN=envie_xxxxx
  envie export --project my-project > .env

Usage in Docker:
  ARG ENVIE_TOKEN
  RUN envie export --project my-project --format dotenv > .env`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global persistent flags (available to all commands)
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "CLI identity token (or set ENVIE_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&project, "project", "", "Project ID or name")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "https://api.envie.sh", "Envie API URL")
}

// getToken returns the token from flag or environment variable
func getToken() (string, error) {
	if token != "" {
		return token, nil
	}
	if envToken := os.Getenv("ENVIE_TOKEN"); envToken != "" {
		return envToken, nil
	}
	return "", fmt.Errorf("no token provided: use --token flag or set ENVIE_TOKEN environment variable")
}

// getProject returns the project from flag or environment variable
func getProject() (string, error) {
	if project != "" {
		return project, nil
	}
	if envProject := os.Getenv("ENVIE_PROJECT"); envProject != "" {
		return envProject, nil
	}
	return "", fmt.Errorf("no project provided: use --project flag or set ENVIE_PROJECT environment variable")
}
