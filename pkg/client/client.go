package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
func (c *Client) SubmitOracleRequest(query string) (*models.OracleResult, error) {
	req := models.OracleRequest{
		ID:    utils.GenerateRequestID(),
		Query: query,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	log.Printf("📤 Submitting request %s: %s", req.ID, req.Query)

	resp, err := c.httpClient.Post(c.coordinatorURL+"/request", "application/json",
		bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to submit request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coordinator returned status %d", resp.StatusCode)
	}

	var result models.OracleResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

// SubmitOracleRequestWithContext submits a request with a custom context
func (c *Client) SubmitOracleRequestWithContext(ctx context.Context, query string) (*models.OracleResult, error) {
	req := models.OracleRequest{
		ID:    utils.GenerateRequestID(),
		Query: query,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.coordinatorURL+"/request",
		bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	log.Printf("📤 Submitting request %s: %s", req.ID, req.Query)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to submit request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coordinator returned status %d", resp.StatusCode)
	}

	var result models.OracleResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &result, nil
}

// SimulateClient submits a batch of requests and logs results
func SimulateClient(coordinatorURL string, queries []string) {
	client := NewClient(coordinatorURL)

	log.Printf("🚀 Starting client simulation with %d queries", len(queries))

	for i, query := range queries {
		log.Printf("\n--- Request %d ---", i+1)

		result, err := client.SubmitOracleRequest(query)
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
				fmt.Printf("  ✅ %s: $%.2f (took %v, reliable: %t)\n",
					workerResult.WorkerID, workerResult.Value, workerResult.ResponseTime, workerResult.Reliable)
			}
		}

		// Add delay between requests
		if i < len(queries)-1 {
			time.Sleep(2 * time.Second)
		}
	}

	log.Printf("✅ Client simulation completed")
}

// submitSingleRequest creates and submits one oracle request
func submitSingleRequest(client *Client, query string) *models.OracleResult {
	result, err := client.SubmitOracleRequest(query)
	if err != nil {
		log.Printf("❌ Request failed: %v", err)
		return nil
	}
	return result
}
