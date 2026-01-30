package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the Envie API client
type Client struct {
	baseURL    string
	identityID string
	httpClient *http.Client
}

// ConfigItem represents an encrypted config item from the API
type ConfigItem struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	EncryptedValue string  `json:"encryptedValue"`
	Description    *string `json:"description,omitempty"`
	ExpiresAt      *string `json:"expiresAt,omitempty"`
}

// ProjectConfigResponse is the response from the config endpoint
type ProjectConfigResponse struct {
	ProjectID           string       `json:"projectId"`
	ProjectName         string       `json:"projectName"`
	EncryptedProjectKey string       `json:"encryptedProjectKey"`
	Items               []ConfigItem `json:"items"`
	ConfigChecksum      string       `json:"configChecksum"`
}

// IdentityInfo contains information about the CLI token
type IdentityInfo struct {
	TokenID     string  `json:"tokenId"`
	TokenName   string  `json:"tokenName"`
	ProjectID   string  `json:"projectId"`
	ProjectName string  `json:"projectName"`
	ExpiresAt   *string `json:"expiresAt,omitempty"`
}

// ErrorResponse represents an API error
type ErrorResponse struct {
	Error string `json:"error"`
}

// NewClient creates a new API client with CLI identity authentication
func NewClient(baseURL, identityID string) *Client {
	return &Client{
		baseURL:    baseURL,
		identityID: identityID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetProjectConfig fetches the encrypted config for a project
func (c *Client) GetProjectConfig(projectID string) (*ProjectConfigResponse, error) {
	url := fmt.Sprintf("%s/v1/projects/%s/config", c.baseURL, projectID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleError(resp)
	}

	var configResp ProjectConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&configResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &configResp, nil
}

// VerifyIdentity verifies the CLI identity and returns identity info
func (c *Client) VerifyIdentity() (*IdentityInfo, error) {
	url := fmt.Sprintf("%s/v1/cli/verify", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleError(resp)
	}

	var info IdentityInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &info, nil
}

// setHeaders sets common headers for API requests
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("X-CLI-Identity", c.identityID)
	req.Header.Set("User-Agent", "envie-cli/1.0")
	req.Header.Set("Accept", "application/json")
}

// handleError parses and returns an appropriate error from the response
func (c *Client) handleError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	var errResp ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return fmt.Errorf("%s (status %d)", errResp.Error, resp.StatusCode)
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized: invalid or expired token")
	case http.StatusForbidden:
		return fmt.Errorf("forbidden: no access to this project")
	case http.StatusNotFound:
		return fmt.Errorf("not found: project does not exist")
	default:
		return fmt.Errorf("API error: status %d", resp.StatusCode)
	}
}
