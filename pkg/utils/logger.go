package utils

import (
	"fmt"
	"log"
	"time"

	"distributed-worker-system/pkg/models"

	"github.com/google/uuid"
)

// LogWorkerResult logs worker responses with metadata
func LogWorkerResult(res models.WorkerResult) {
	if res.Err != "" {
		log.Printf("❌ Worker %s failed: %s (took %v)", res.WorkerID, res.Err, res.ResponseTime)
	} else {
		log.Printf("✅ Worker %s: value=%.2f, time=%v, reliable=%t",
			res.WorkerID, res.Value, res.ResponseTime, res.Reliable)
	}
}

// LogOracleResult logs final aggregated result
func LogOracleResult(result models.OracleResult) {
	log.Printf("🎯 Final Result: value=%.2f, workers=%d, note='%s'",
		result.FinalValue, len(result.WorkerResponses), result.ReliabilityNote)

	for _, res := range result.WorkerResponses {
		LogWorkerResult(res)
	}
}

// GenerateRequestID creates unique IDs for oracle requests
func GenerateRequestID() string {
	return fmt.Sprintf("req-%s", uuid.New().String()[:8])
}

// GenerateWorkerID creates unique IDs for workers
func GenerateWorkerID() string {
	return fmt.Sprintf("worker-%s", uuid.New().String()[:8])
}

// CalculateReliability calculates worker reliability based on success rate
func CalculateReliability(successCount, totalCount int) bool {
	if totalCount == 0 {
		return false
	}
	successRate := float64(successCount) / float64(totalCount)
	return successRate >= 0.7 // 70% success rate threshold
}

// FormatDuration formats duration for logging
func FormatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fμs", float64(d.Nanoseconds())/1000)
	} else if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Nanoseconds())/1e6)
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
