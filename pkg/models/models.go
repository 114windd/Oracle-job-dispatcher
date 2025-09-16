package models

import "time"

// OracleRequest represents a request to fetch data from oracles
type OracleRequest struct {
	ID    string `json:"id"`
	Query string `json:"query"`
}

// WorkerResult represents the response from a worker
type WorkerResult struct {
	WorkerID     string        `json:"worker_id"`
	RequestID    string        `json:"request_id"`
	Value        float64       `json:"value"`
	Err          string        `json:"err,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
	Reliable     bool          `json:"reliable"`
}

// OracleResult represents the final aggregated response
type OracleResult struct {
	RequestID       string         `json:"request_id"`
	FinalValue      float64        `json:"final_value"`
	WorkerResponses []WorkerResult `json:"worker_responses"`
	ReliabilityNote string         `json:"reliability_note"`
}

// WorkerInfo represents information about a registered worker
type WorkerInfo struct {
	ID       string    `json:"id"`
	Endpoint string    `json:"endpoint"`
	LastSeen time.Time `json:"last_seen"`
	Reliable bool      `json:"reliable"`
}

// RegisterRequest represents a worker registration request
type RegisterRequest struct {
	ID       string `json:"id"`
	Endpoint string `json:"endpoint"`
}

// RegisterResponse represents the response to a worker registration
type RegisterResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
