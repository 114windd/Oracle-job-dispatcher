package coordinator

import (

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


// AggregateResults aggregates worker results using the specified strategy
func AggregateResults(results []models.WorkerResult, strategy string) float64 {
	
	return aggregateAverage(results)

}
