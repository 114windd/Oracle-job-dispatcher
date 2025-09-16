package coordinator

import (
	"sort"

	"distributed-worker-system/pkg/models"
)

// aggregateAverage computes the average of successful worker results
func aggregateAverage(results []models.WorkerResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	var sum float64
	var count int

	for _, result := range results {
		if result.Err == "" { // Only count successful results
			sum += result.Value
			count++
		}
	}

	if count == 0 {
		return 0.0
	}

	return sum / float64(count)
}

// aggregateMedian computes the median result
func aggregateMedian(results []models.WorkerResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	// Filter successful results
	var values []float64
	for _, result := range results {
		if result.Err == "" {
			values = append(values, result.Value)
		}
	}

	if len(values) == 0 {
		return 0.0
	}

	// Sort values
	sort.Float64s(values)

	// Calculate median
	n := len(values)
	if n%2 == 0 {
		return (values[n/2-1] + values[n/2]) / 2.0
	}
	return values[n/2]
}

// aggregateMajorityVote computes majority vote (rounding values, useful for categorical data)
func aggregateMajorityVote(results []models.WorkerResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	// Filter successful results and round to nearest integer
	voteCounts := make(map[int]int)
	var values []float64

	for _, result := range results {
		if result.Err == "" {
			rounded := int(result.Value + 0.5) // Round to nearest integer
			voteCounts[rounded]++
			values = append(values, result.Value)
		}
	}

	if len(values) == 0 {
		return 0.0
	}

	// Find the value with the most votes
	var majorityValue int
	var maxVotes int

	for value, votes := range voteCounts {
		if votes > maxVotes {
			majorityValue = value
			maxVotes = votes
		}
	}

	return float64(majorityValue)
}

// AggregateResults aggregates worker results using the specified strategy
func AggregateResults(results []models.WorkerResult, strategy string) float64 {
	switch strategy {
	case "median":
		return aggregateMedian(results)
	case "majority":
		return aggregateMajorityVote(results)
	default: // "average"
		return aggregateAverage(results)
	}
}
