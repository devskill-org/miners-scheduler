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

func TestDownloadPublicationMarketData_Success(t *testing.T) {
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

	doc, err := client.DownloadPublicationMarketData(ctx, server.URL)

	if err != nil {
		t.Fatalf("DownloadPublicationMarketData() failed: %v", err)
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

func TestDownloadPublicationMarketData_EmptyURL(t *testing.T) {
	client := NewAPIClient()
	ctx := context.Background()

	_, err := client.DownloadPublicationMarketData(ctx, "")

	if err == nil {
		t.Fatal("Expected error for empty URL, got nil")
	}

	expectedError := "API URL cannot be empty"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestDownloadPublicationMarketData_HTTPError(t *testing.T) {
	// Create a test server that returns HTTP 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewAPIClient()
	ctx := context.Background()

	_, err := client.DownloadPublicationMarketData(ctx, server.URL)

	if err == nil {
		t.Fatal("Expected error for HTTP 500, got nil")
	}

	expectedErrorPrefix := "HTTP request failed with status 500"
	if !strings.HasPrefix(err.Error(), expectedErrorPrefix) {
		t.Errorf("Expected error starting with '%s', got '%s'", expectedErrorPrefix, err.Error())
	}
}

func TestDownloadPublicationMarketData_InvalidXML(t *testing.T) {
	// Create a test server that returns invalid XML
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<invalid><xml></invalid>"))
	}))
	defer server.Close()

	client := NewAPIClient()
	ctx := context.Background()

	_, err := client.DownloadPublicationMarketData(ctx, server.URL)

	if err == nil {
		t.Fatal("Expected error for invalid XML, got nil")
	}

	expectedErrorPrefix := "failed to decode XML response"
	if !strings.HasPrefix(err.Error(), expectedErrorPrefix) {
		t.Errorf("Expected error starting with '%s', got '%s'", expectedErrorPrefix, err.Error())
	}
}

func TestDownloadPublicationMarketData_ContextCancellation(t *testing.T) {
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

	_, err := client.DownloadPublicationMarketData(ctx, server.URL)

	if err == nil {
		t.Fatal("Expected error for context timeout, got nil")
	}

	// The exact error message depends on the context implementation
	if !strings.Contains(err.Error(), "context deadline exceeded") && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context cancellation error, got '%s'", err.Error())
	}
}

func TestDownloadPublicationMarketDataWithOptions_Success(t *testing.T) {
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

	doc, err := DownloadPublicationMarketDataWithOptions(ctx, server.URL, opts)

	if err != nil {
		t.Fatalf("DownloadPublicationMarketDataWithOptions() failed: %v", err)
	}

	if doc == nil {
		t.Fatal("Returned document is nil")
	}

	if doc.MRID != "1" {
		t.Errorf("Expected MRID '1', got '%s'", doc.MRID)
	}
}

func TestDownloadPublicationMarketDataWithOptions_EmptyURL(t *testing.T) {
	ctx := context.Background()
	opts := &DownloadOptions{}

	_, err := DownloadPublicationMarketDataWithOptions(ctx, "", opts)

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

// Unit tests for buildPublicationMarketDataURL
func TestBuildPublicationMarketDataURL(t *testing.T) {
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
			url := buildPublicationMarketDataURL(securityToken, urlFormat, tt.now)
			if url != tt.expected {
				t.Errorf("For %s, got url: %s, want: %s", tt.name, url, tt.expected)
			}
		})
	}
}

func TestDownloadPublicationMarketData_CustomUserAgent(t *testing.T) {
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

	_, err := client.DownloadPublicationMarketData(ctx, server.URL)

	if err != nil {
		t.Fatalf("DownloadPublicationMarketData() failed: %v", err)
	}
}

// Benchmark tests
func BenchmarkDownloadPublicationMarketData(b *testing.B) {
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
		_, err := client.DownloadPublicationMarketData(ctx, server.URL)
		if err != nil {
			b.Fatalf("DownloadPublicationMarketData() failed: %v", err)
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

// TestMergePublicationMarketData tests the merging of two PublicationMarketData objects
func TestMergePublicationMarketData(t *testing.T) {
	// Create first document
	doc1 := &PublicationMarketData{
		MRID:           "doc1",
		RevisionNumber: 1,
		Type:           "A44",
		PeriodTimeInterval: TimeInterval{
			Start: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC),
		},
		TimeSeries: []TimeSeries{
			{
				MRID:         "ts1",
				BusinessType: "A62",
				Period: Period{
					TimeInterval: TimeInterval{
						Start: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC),
					},
					Resolution: time.Hour,
					Points: []Point{
						{Position: 1, PriceAmount: 45.50},
						{Position: 2, PriceAmount: 42.30},
					},
				},
			},
		},
	}

	// Create second document
	doc2 := &PublicationMarketData{
		MRID:           "doc2",
		RevisionNumber: 1,
		Type:           "A44",
		PeriodTimeInterval: TimeInterval{
			Start: time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC),
		},
		TimeSeries: []TimeSeries{
			{
				MRID:         "ts2",
				BusinessType: "A62",
				Period: Period{
					TimeInterval: TimeInterval{
						Start: time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC),
						End:   time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC),
					},
					Resolution: time.Hour,
					Points: []Point{
						{Position: 1, PriceAmount: 50.00},
						{Position: 2, PriceAmount: 48.75},
					},
				},
			},
		},
	}

	// Test merging
	merged := mergePublicationMarketData(doc1, doc2)

	// Verify the merged document contains TimeSeries from both documents
	if len(merged.TimeSeries) != 2 {
		t.Errorf("Expected 2 TimeSeries in merged document, got %d", len(merged.TimeSeries))
	}

	// Verify first TimeSeries is from doc1
	if merged.TimeSeries[0].MRID != "ts1" {
		t.Errorf("Expected first TimeSeries MRID 'ts1', got '%s'", merged.TimeSeries[0].MRID)
	}

	// Verify second TimeSeries is from doc2
	if merged.TimeSeries[1].MRID != "ts2" {
		t.Errorf("Expected second TimeSeries MRID 'ts2', got '%s'", merged.TimeSeries[1].MRID)
	}

	// Verify the period time interval is updated to span both documents
	expectedEnd := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	if !merged.PeriodTimeInterval.End.Equal(expectedEnd) {
		t.Errorf("Expected merged PeriodTimeInterval.End to be %v, got %v", expectedEnd, merged.PeriodTimeInterval.End)
	}

	// Verify the start time is preserved from the first document
	expectedStart := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	if !merged.PeriodTimeInterval.Start.Equal(expectedStart) {
		t.Errorf("Expected merged PeriodTimeInterval.Start to be %v, got %v", expectedStart, merged.PeriodTimeInterval.Start)
	}

	// Verify that the original document is not modified
	if len(doc1.TimeSeries) != 1 {
		t.Errorf("Original doc1 should still have 1 TimeSeries, got %d", len(doc1.TimeSeries))
	}
}

// TestMergePublicationMarketData_NilInputs tests merging with nil inputs
func TestMergePublicationMarketData_NilInputs(t *testing.T) {
	doc := &PublicationMarketData{
		MRID: "doc1",
		TimeSeries: []TimeSeries{
			{MRID: "ts1"},
		},
	}

	// Test first parameter nil
	result := mergePublicationMarketData(nil, doc)
	if result != doc {
		t.Error("Expected merging nil with doc to return doc")
	}

	// Test second parameter nil
	result = mergePublicationMarketData(doc, nil)
	if result != doc {
		t.Error("Expected merging doc with nil to return doc")
	}

	// Test both parameters nil
	result = mergePublicationMarketData(nil, nil)
	if result != nil {
		t.Error("Expected merging nil with nil to return nil")
	}
}

// TestMergePublicationMarketData_EndTimeNotExtended tests that end time is not updated if second doc has earlier end time
func TestMergePublicationMarketData_EndTimeNotExtended(t *testing.T) {
	doc1 := &PublicationMarketData{
		MRID: "doc1",
		PeriodTimeInterval: TimeInterval{
			Start: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC),
		},
		TimeSeries: []TimeSeries{
			{MRID: "ts1"},
		},
	}

	doc2 := &PublicationMarketData{
		MRID: "doc2",
		PeriodTimeInterval: TimeInterval{
			Start: time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC),
			End:   time.Date(2024, 6, 2, 12, 0, 0, 0, time.UTC),
		},
		TimeSeries: []TimeSeries{
			{MRID: "ts2"},
		},
	}

	merged := mergePublicationMarketData(doc1, doc2)

	// The end time should remain from doc1 since doc2's end time is earlier
	expectedEnd := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	if !merged.PeriodTimeInterval.End.Equal(expectedEnd) {
		t.Errorf("Expected merged PeriodTimeInterval.End to remain %v, got %v", expectedEnd, merged.PeriodTimeInterval.End)
	}
}

// Example test showing how to use the API client
func ExampleAPIClient_DownloadPublicationMarketData() {
	client := NewAPIClient()
	ctx := context.Background()

	// This would be a real ENTSO-E API URL in practice
	apiURL := "https://web-api.tp.entsoe.eu/api?documentType=A44&in_Domain=10YCZ-CEPS-----N&out_Domain=10YCZ-CEPS-----N&periodStart=202509052300&periodEnd=202509062300&securityToken=YOUR_TOKEN"

	doc, err := client.DownloadPublicationMarketData(ctx, apiURL)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Document MRID: %s\n", doc.MRID)
	fmt.Printf("Number of TimeSeries: %d\n", len(doc.TimeSeries))
}
