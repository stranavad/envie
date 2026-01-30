package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/stranavad/envie/cli/internal/api"
	"github.com/stranavad/envie/cli/internal/config"
	"github.com/stranavad/envie/cli/internal/crypto"
	"github.com/spf13/cobra"
)

var authToken string

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Envie",
	Long: `Authenticate the CLI with a project token.

Project tokens are created in the Envie desktop app:
  1. Open the Envie app
  2. Go to a project's Settings > Tokens
  3. Click "Create Token"
  4. Copy the generated token

Note: Each token is tied to a specific project and is read-only.

Usage:
  envie auth --token envie_xxxxx
  envie auth  # Interactive prompt`,
	RunE: runAuth,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	Long:  `Remove the stored CLI identity token from your machine.`,
	RunE:  runLogout,
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current identity",
	Long:  `Display information about the currently authenticated CLI identity.`,
	RunE:  runWhoami,
}

func init() {
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)

	authCmd.Flags().StringVar(&authToken, "token", "", "CLI identity token")
}

func runAuth(cmd *cobra.Command, args []string) error {
	tokenValue := authToken

	// If no token provided via flag, prompt for it
	if tokenValue == "" {
		fmt.Println("To authenticate, create a Project Token in the Envie app:")
		fmt.Println()
		fmt.Println("  1. Open the Envie desktop app")
		fmt.Println("  2. Go to your project's Settings > Tokens")
		fmt.Println("  3. Click 'Create Token'")
		fmt.Println("  4. Copy the generated token")
		fmt.Println()
		fmt.Print("Paste your token here: ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		tokenValue = strings.TrimSpace(input)

		if tokenValue == "" {
			return fmt.Errorf("no token provided")
		}
	}

	// Validate token format
	identity, err := crypto.ParseToken(tokenValue)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// Verify with server
	fmt.Print("Verifying token... ")
	client := api.NewClient(apiURL, identity.IdentityID)
	info, err := client.VerifyIdentity()
	if err != nil {
		fmt.Println("failed")
		return fmt.Errorf("authentication failed: %w", err)
	}
	fmt.Println("ok")

	// Store credentials
	creds := &config.Credentials{
		Token: tokenValue,
	}
	if err := config.StoreCredentials(creds); err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	credsPath, _ := config.GetCredentialsPath()
	fmt.Println()
	fmt.Printf("✓ Authenticated for project: %s\n", info.ProjectName)
	fmt.Printf("  Token name: %s\n", info.TokenName)
	if info.ExpiresAt != nil {
		fmt.Printf("  Expires: %s\n", *info.ExpiresAt)
	} else {
		fmt.Printf("  Expires: never\n")
	}
	fmt.Printf("  Credentials saved to: %s\n", credsPath)

	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	if err := config.ClearCredentials(); err != nil {
		return err
	}
	fmt.Println("✓ Logged out. Credentials removed.")
	return nil
}

func runWhoami(cmd *cobra.Command, args []string) error {
	// Get token
	tokenValue, err := getToken()
	if err != nil {
		// Try loading from credentials file
		creds, err := config.LoadCredentials()
		if err != nil {
			return fmt.Errorf("not authenticated: run 'envie auth' first")
		}
		tokenValue = creds.Token
	}

	// Parse and verify
	identity, err := crypto.ParseToken(tokenValue)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	client := api.NewClient(apiURL, identity.IdentityID)
	info, err := client.VerifyIdentity()
	if err != nil {
		return fmt.Errorf("failed to verify identity: %w", err)
	}

	fmt.Printf("Project:    %s\n", info.ProjectName)
	fmt.Printf("Project ID: %s\n", info.ProjectID)
	fmt.Printf("Token:      %s\n", info.TokenName)
	if info.ExpiresAt != nil {
		fmt.Printf("Expires:    %s\n", *info.ExpiresAt)
	} else {
		fmt.Printf("Expires:    never\n")
	}

	return nil
}
