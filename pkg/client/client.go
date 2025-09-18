package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"distributed-worker-system/pkg/models"
	"distributed-worker-system/pkg/utils"
)

// Client represents a client that submits oracle requests
type Client struct {
	coordinatorURL string
	httpClient     *http.Client
}

// NewClient creates a new client instance
func NewClient(coordinatorURL string) *Client {
	return &Client{
		coordinatorURL: coordinatorURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SubmitOracleRequest submits a single oracle request to the coordinator
func (c *Client) SubmitOracleRequest(ctx context.Context, query string) (*models.OracleResult, error) {
	req := models.OracleRequest{
		ID:    utils.GenerateRequestID(),
		Query: query,
	}

	log.Printf("📤 Submitting request %s: %s", req.ID, req.Query)

	// Make HTTP request with context
	var result models.OracleResult
	err := c.makeRequest(ctx, "POST", "/request", req, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// SimulateClient submits a batch of requests and logs results
func SimulateClient(coordinatorURL string, queries []string) {
	client := NewClient(coordinatorURL)

	log.Printf("🚀 Starting client simulation with %d queries", len(queries))

	for i, query := range queries {
		log.Printf("\n--- Request %d ---", i+1)

		result, err := client.SubmitOracleRequest(context.Background(), query)
		if err != nil {
			log.Printf("❌ Request failed: %v", err)
			continue
		}

		// Display results
		fmt.Printf("\n🎯 Query: %s\n", query)
		fmt.Printf("💰 Final Value: $%.2f\n", result.FinalValue)
		fmt.Printf("📊 Workers: %d\n", len(result.WorkerResponses))
		fmt.Printf("⚠️  Note: %s\n", result.ReliabilityNote)

		fmt.Println("\n📋 Worker Details:")
		for _, workerResult := range result.WorkerResponses {
			if workerResult.Err != "" {
				fmt.Printf("  ❌ %s: %s (took %v)\n",
					workerResult.WorkerID, workerResult.Err, workerResult.ResponseTime)
			} else {
				fmt.Printf("  ✅ %s: $%.2f (took %v)\n",
					workerResult.WorkerID, workerResult.Value, workerResult.ResponseTime)
			}
		}

		// Add delay between requests
		if i < len(queries)-1 {
			time.Sleep(2 * time.Second)
		}
	}

	log.Printf("✅ Client simulation completed")
}

// makeRequest makes an HTTP request with context
func (c *Client) makeRequest(ctx context.Context, method string, path string, requestBody any, responseBody any) error {
	// Marshal request body
	reqBytes, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, method, c.coordinatorURL+path, bytes.NewBuffer(reqBytes))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to submit request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("coordinator returned status %d", resp.StatusCode)
	}

	// Read and unmarshal response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if err := json.Unmarshal(body, responseBody); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	return nil
}
