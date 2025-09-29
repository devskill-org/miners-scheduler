package entsoe

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Sample XML response for testing
const sampleXMLResponse = `<?xml version="1.0" encoding="UTF-8"?>
<Publication_MarketDocument xmlns="urn:iec62325.351:tc57wg16:451-3:publicationdocument:7:0">
    <mRID>1</mRID>
    <revisionNumber>1</revisionNumber>
    <type>A44</type>
    <sender_MarketParticipant.mRID codingScheme="A01">10X1001A1001A450</sender_MarketParticipant.mRID>
    <sender_MarketParticipant.marketRole.type>A32</sender_MarketParticipant.marketRole.type>
    <receiver_MarketParticipant.mRID codingScheme="A01">10X1001A1001A450</receiver_MarketParticipant.mRID>
    <receiver_MarketParticipant.marketRole.type>A33</receiver_MarketParticipant.marketRole.type>
    <createdDateTime>2025-09-05T21:00:00Z</createdDateTime>
    <period.timeInterval>
        <start>2025-09-05T22:00Z</start>
        <end>2025-09-06T21:00Z</end>
    </period.timeInterval>
    <TimeSeries>
        <mRID>1</mRID>
        <businessType>A62</businessType>
        <in_Domain.mRID codingScheme="A01">10Y1001A1001A83F</in_Domain.mRID>
        <out_Domain.mRID codingScheme="A01">10Y1001A1001A83F</out_Domain.mRID>
        <currency_Unit.name>EUR</currency_Unit.name>
        <price_Measure_Unit.name>MWH</price_Measure_Unit.name>
        <curveType>A01</curveType>
        <Period>
            <timeInterval>
                <start>2025-09-05T22:00Z</start>
                <end>2025-09-06T21:00Z</end>
            </timeInterval>
            <resolution>PT1H</resolution>
            <Point>
                <position>1</position>
                <price.amount>45.50</price.amount>
            </Point>
            <Point>
                <position>2</position>
                <price.amount>42.30</price.amount>
            </Point>
        </Period>
    </TimeSeries>
</Publication_MarketDocument>`

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient()

	if client == nil {
		t.Fatal("NewAPIClient() returned nil")
	}

	if client.httpClient == nil {
		t.Error("httpClient is nil")
	}

	if client.userAgent != "entsoe-go-client/1.0" {
		t.Errorf("Expected userAgent 'entsoe-go-client/1.0', got '%s'", client.userAgent)
	}

}

func TestSetUserAgent(t *testing.T) {
	client := NewAPIClient()
	customUserAgent := "my-custom-agent/2.0"

	client.SetUserAgent(customUserAgent)

	if client.userAgent != customUserAgent {
		t.Errorf("Expected userAgent '%s', got '%s'", customUserAgent, client.userAgent)
	}
}

func TestDownloadPublicationMarketDocument_Success(t *testing.T) {
	// Create a test server that returns sample XML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("User-Agent") != "entsoe-go-client/1.0" {
			t.Errorf("Expected User-Agent 'entsoe-go-client/1.0', got '%s'", r.Header.Get("User-Agent"))
		}

		expectedAccept := "application/xml, text/xml"
		if r.Header.Get("Accept") != expectedAccept {
			t.Errorf("Expected Accept '%s', got '%s'", expectedAccept, r.Header.Get("Accept"))
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sampleXMLResponse))
	}))
	defer server.Close()

	client := NewAPIClient()
	ctx := context.Background()

	doc, err := client.DownloadPublicationMarketDocument(ctx, server.URL)

	if err != nil {
		t.Fatalf("DownloadPublicationMarketDocument() failed: %v", err)
	}

	if doc == nil {
		t.Fatal("Returned document is nil")
	}

	// Verify some basic properties of the parsed document
	if doc.MRID != "1" {
		t.Errorf("Expected MRID '1', got '%s'", doc.MRID)
	}

	if len(doc.TimeSeries) != 1 {
		t.Errorf("Expected 1 TimeSeries, got %d", len(doc.TimeSeries))
	}

	if len(doc.TimeSeries[0].Period.Points) != 2 {
		t.Errorf("Expected 2 Points, got %d", len(doc.TimeSeries[0].Period.Points))
	}
}

func TestDownloadPublicationMarketDocument_EmptyURL(t *testing.T) {
	client := NewAPIClient()
	ctx := context.Background()

	_, err := client.DownloadPublicationMarketDocument(ctx, "")

	if err == nil {
		t.Fatal("Expected error for empty URL, got nil")
	}

	expectedError := "API URL cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestDownloadPublicationMarketDocument_HTTPError(t *testing.T) {
	// Create a test server that returns HTTP 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewAPIClient()
	ctx := context.Background()

	_, err := client.DownloadPublicationMarketDocument(ctx, server.URL)

	if err == nil {
		t.Fatal("Expected error for HTTP 500, got nil")
	}

	expectedErrorPrefix := "HTTP request failed with status 500"
	if !strings.HasPrefix(err.Error(), expectedErrorPrefix) {
		t.Errorf("Expected error starting with '%s', got '%s'", expectedErrorPrefix, err.Error())
	}
}

func TestDownloadPublicationMarketDocument_InvalidXML(t *testing.T) {
	// Create a test server that returns invalid XML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<invalid><xml></invalid>"))
	}))
	defer server.Close()

	client := NewAPIClient()
	ctx := context.Background()

	_, err := client.DownloadPublicationMarketDocument(ctx, server.URL)

	if err == nil {
		t.Fatal("Expected error for invalid XML, got nil")
	}

	expectedErrorPrefix := "failed to decode XML response"
	if !strings.HasPrefix(err.Error(), expectedErrorPrefix) {
		t.Errorf("Expected error starting with '%s', got '%s'", expectedErrorPrefix, err.Error())
	}
}

func TestDownloadPublicationMarketDocument_ContextCancellation(t *testing.T) {
	// Create a test server with a delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sampleXMLResponse))
	}))
	defer server.Close()

	client := NewAPIClient()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.DownloadPublicationMarketDocument(ctx, server.URL)

	if err == nil {
		t.Fatal("Expected error for context timeout, got nil")
	}

	// The exact error message depends on the context implementation
	if !strings.Contains(err.Error(), "context deadline exceeded") && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context cancellation error, got '%s'", err.Error())
	}
}

func TestDownloadPublicationMarketDocumentWithOptions_Success(t *testing.T) {
	customUserAgent := "test-agent/1.0"
	customHeader := "test-value"

	// Create a test server that verifies custom options
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != customUserAgent {
			t.Errorf("Expected User-Agent '%s', got '%s'", customUserAgent, r.Header.Get("User-Agent"))
		}

		if r.Header.Get("X-Custom-Header") != customHeader {
			t.Errorf("Expected X-Custom-Header '%s', got '%s'", customHeader, r.Header.Get("X-Custom-Header"))
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sampleXMLResponse))
	}))
	defer server.Close()

	ctx := context.Background()
	opts := &DownloadOptions{
		UserAgent: customUserAgent,
		Headers: map[string]string{
			"X-Custom-Header": customHeader,
		},
	}

	doc, err := DownloadPublicationMarketDocumentWithOptions(ctx, server.URL, opts)

	if err != nil {
		t.Fatalf("DownloadPublicationMarketDocumentWithOptions() failed: %v", err)
	}

	if doc == nil {
		t.Fatal("Returned document is nil")
	}

	if doc.MRID != "1" {
		t.Errorf("Expected MRID '1', got '%s'", doc.MRID)
	}
}

func TestDownloadPublicationMarketDocumentWithOptions_EmptyURL(t *testing.T) {
	ctx := context.Background()
	opts := &DownloadOptions{}

	_, err := DownloadPublicationMarketDocumentWithOptions(ctx, "", opts)

	if err == nil {
		t.Fatal("Expected error for empty URL, got nil")
	}

	expectedError := "API URL cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestValidateAPIURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantError bool
	}{
		{
			name:      "valid HTTPS URL",
			url:       "https://web-api.tp.entsoe.eu/api",
			wantError: false,
		},
		{
			name:      "valid HTTP URL",
			url:       "http://example.com/api",
			wantError: false,
		},
		{
			name:      "empty URL",
			url:       "",
			wantError: true,
		},
		{
			name:      "too short URL",
			url:       "http",
			wantError: true,
		},
		{
			name:      "invalid protocol",
			url:       "ftp://example.com",
			wantError: true,
		},
		{
			name:      "no protocol",
			url:       "example.com/api",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAPIURL(tt.url)
			if tt.wantError && err == nil {
				t.Errorf("ValidateAPIURL() expected error for URL '%s', got nil", tt.url)
			}
			if !tt.wantError && err != nil {
				t.Errorf("ValidateAPIURL() unexpected error for URL '%s': %v", tt.url, err)
			}
		})
	}
}

// Unit tests for buildPublicationMarketDocumentURL
func TestBuildPublicationMarketDocumentURL(t *testing.T) {
	securityToken := "test-token"
	urlFormat := "https://example.com?start=%s&end=%s&token=%s"

	location, err := time.LoadLocation("CET")
	if err != nil {
		t.Fatalf("Failed to load EET location: %v", err)
	}

	tests := []struct {
		name     string
		now      time.Time
		expected string
	}{
		{
			name:     "22:00",
			now:      time.Date(2024, 6, 1, 22, 0, 0, 0, location),
			expected: "https://example.com?start=202405312200&end=202406012200&token=test-token",
		},
		{
			name:     "23:00",
			now:      time.Date(2024, 6, 1, 23, 0, 0, 0, location),
			expected: "https://example.com?start=202405312200&end=202406012200&token=test-token",
		},
		{
			name:     "00:00",
			now:      time.Date(2024, 6, 2, 0, 0, 0, 0, location),
			expected: "https://example.com?start=202406012200&end=202406022200&token=test-token",
		},
		{
			name:     "00:01",
			now:      time.Date(2024, 6, 2, 0, 1, 0, 0, location),
			expected: "https://example.com?start=202406012200&end=202406022200&token=test-token",
		},
		{
			name:     "01:00",
			now:      time.Date(2024, 6, 2, 1, 0, 0, 0, location),
			expected: "https://example.com?start=202406012200&end=202406022200&token=test-token",
		},
		{
			name:     "02:00",
			now:      time.Date(2024, 6, 2, 2, 0, 0, 0, location),
			expected: "https://example.com?start=202406012200&end=202406022200&token=test-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := buildPublicationMarketDocumentURL(securityToken, urlFormat, tt.now)
			if url != tt.expected {
				t.Errorf("For %s, got url: %s, want: %s", tt.name, url, tt.expected)
			}
		})
	}
}

func TestDownloadPublicationMarketDocument_CustomUserAgent(t *testing.T) {
	customUserAgent := "my-test-agent/3.0"

	// Create a test server that checks the user agent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != customUserAgent {
			t.Errorf("Expected User-Agent '%s', got '%s'", customUserAgent, r.Header.Get("User-Agent"))
		}

		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sampleXMLResponse))
	}))
	defer server.Close()

	client := NewAPIClient()
	client.SetUserAgent(customUserAgent)
	ctx := context.Background()

	_, err := client.DownloadPublicationMarketDocument(ctx, server.URL)

	if err != nil {
		t.Fatalf("DownloadPublicationMarketDocument() failed: %v", err)
	}
}

// Benchmark tests
func BenchmarkDownloadPublicationMarketDocument(b *testing.B) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sampleXMLResponse))
	}))
	defer server.Close()

	client := NewAPIClient()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.DownloadPublicationMarketDocument(ctx, server.URL)
		if err != nil {
			b.Fatalf("DownloadPublicationMarketDocument() failed: %v", err)
		}
	}
}

func BenchmarkValidateAPIURL(b *testing.B) {
	url := "https://web-api.tp.entsoe.eu/api"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateAPIURL(url)
	}
}

// Example test showing how to use the API client
func ExampleAPIClient_DownloadPublicationMarketDocument() {
	client := NewAPIClient()
	ctx := context.Background()

	// This would be a real ENTSO-E API URL in practice
	apiURL := "https://web-api.tp.entsoe.eu/api?documentType=A44&in_Domain=10YCZ-CEPS-----N&out_Domain=10YCZ-CEPS-----N&periodStart=202509052300&periodEnd=202509062300&securityToken=YOUR_TOKEN"

	doc, err := client.DownloadPublicationMarketDocument(ctx, apiURL)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Document MRID: %s\n", doc.MRID)
	fmt.Printf("Number of TimeSeries: %d\n", len(doc.TimeSeries))
}
