package entsoe

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/devskill-org/miners-scheduler/utils"
)

// APIClient represents an HTTP client for the ENTSO-E API
type APIClient struct {
	httpClient *http.Client
	userAgent  string
}

// NewAPIClient creates a new ENTSO-E API client with default settings
func NewAPIClient() *APIClient {
	return &APIClient{
		httpClient: &http.Client{},
		userAgent:  "entsoe-go-client/1.0",
	}
}

// SetUserAgent sets a custom user agent for the API client
func (c *APIClient) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

// DownloadPublicationMarketDocument downloads and decodes a PublicationMarketDocument from the given API URL
func (c *APIClient) DownloadPublicationMarketDocument(ctx context.Context, apiURL string) (*PublicationMarketDocument, error) {
	opts := &DownloadOptions{
		UserAgent: c.userAgent,
	}

	return DownloadPublicationMarketDocumentWithOptions(ctx, apiURL, opts)
}

// DownloadPublicationMarketDocumentWithOptions downloads and decodes a PublicationMarketDocument with additional options
type DownloadOptions struct {
	UserAgent string
	Headers   map[string]string
}

func DownloadPublicationMarketDocument(ctx context.Context, securityToken string, urlFormat string, locationStr string) (*PublicationMarketDocument, error) {
	location, err := time.LoadLocation(locationStr)
	if err != nil {
		return nil, err
	}

	now := time.Now().In(location)
	url := buildPublicationMarketDocumentURL(securityToken, urlFormat, now)

	client := NewAPIClient()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	marketDocument, err := client.DownloadPublicationMarketDocument(ctx, url)
	if err != nil {
		return nil, err
	}
	return marketDocument, nil
}

// buildPublicationMarketDocumentURL extracts the URL assignment logic for DownloadPublicationMarketDocument.
func buildPublicationMarketDocumentURL(securityToken string, urlFormat string, now time.Time) string {
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	periodStart := utils.GetUTCString(start)
	periodEnd := utils.GetUTCString(start.AddDate(0, 0, 1))

	return fmt.Sprintf(urlFormat, periodStart, periodEnd, securityToken)
}

// DownloadPublicationMarketDocumentWithOptions downloads and decodes a PublicationMarketDocument with custom options
func DownloadPublicationMarketDocumentWithOptions(ctx context.Context, apiURL string, opts *DownloadOptions) (*PublicationMarketDocument, error) {
	if apiURL == "" {
		return nil, fmt.Errorf("API URL cannot be empty")
	}

	client := &http.Client{}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set default headers
	userAgent := "entsoe-go-client/1.0"
	if opts.UserAgent != "" {
		userAgent = opts.UserAgent
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/xml, text/xml")

	// Set custom headers
	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Decode the XML response using the existing decoder
	doc, err := DecodeEnergyPricesXML(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode XML response: %w", err)
	}

	return doc, nil
}

// ValidateAPIURL performs basic validation on the API URL
func ValidateAPIURL(apiURL string) error {
	if apiURL == "" {
		return fmt.Errorf("API URL cannot be empty")
	}

	// Basic URL validation - in production you might want more sophisticated validation
	if len(apiURL) < 7 { // minimum: http://
		return fmt.Errorf("API URL appears to be too short")
	}

	if apiURL[:7] != "http://" && apiURL[:8] != "https://" {
		return fmt.Errorf("API URL must start with http:// or https://")
	}

	return nil
}
