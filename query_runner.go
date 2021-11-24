package main

import (
	"sync"
	"time"
)

type QueryRunner struct {
	LoadBalancer ILoadBalancer
	Db           IDB
	WorkersCount int
}

// Run runs an array of queries.
// It returns an array of query run time results in Nanoseconds, total Processing time and any error that may have occured.
func (b QueryRunner) Run(queryInputs []QueryInput) ([]int64, int64) {

	// Extract query count per host and create a map of host to their query counts
	hostQueryCount := createHostCountMap(queryInputs)

	// Get a map of host names to the index of the worker that will run them
	hostToWorker := b.LoadBalancer.GetBalancedLoads(hostQueryCount, b.WorkersCount)

	// Create an input channel for each worker and one shared outputChannel
	inputChans := make([]chan QueryInput, b.WorkersCount)
	for i := range inputChans {
		inputChans[i] = make(chan QueryInput)
	}
	outputChannel := make(chan []int64, b.WorkersCount)

	// Create a wait group.
	// Wait group is needed to ensure all workers are finished before aggregating the results and leaving the function
	wg := &sync.WaitGroup{}

	// Create workers for each input channel
	for i, _ := range inputChans {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			var results []int64
			for queryInput := range inputChans[idx] {
				// Errors while running a query are ignored and the program keeps running.
				// If we wanted to provid statistics for errors, we could add errors to a different output channel and
				// return the aggregated errors or error statistics from this function
				start := time.Now()
				_ = b.Db.RunCPUUsageQuery(queryInput.HostName, queryInput.StartTime, queryInput.EndTime)
				elapsed := time.Since(start)
				results = append(results, int64(elapsed/time.Nanosecond))
			}
			outputChannel <- results
		}(i)
	}

	// Record the start of query processing.
	// Individual query run times and the total processing time will be measured in nanoseconds.
	processingStartTime := time.Now()
	var totalProcessingTime int64

	// Populate workers' input channels by the results taken from loadBalancer
	for _, query := range queryInputs {
		workerIndex := hostToWorker[query.HostName]
		inputChans[workerIndex] <- query
	}

	// Close all the input channels
	for _, inputChan := range inputChans {
		close(inputChan)
	}

	// A seperate goroutine created to wait on the workers to finish and close the output channel
	go func() {
		wg.Wait()
		totalProcessingTime = int64(time.Since(processingStartTime) / time.Nanosecond)
		close(outputChannel)
	}()

	// Aggregate all results into outputChannel
	var results []int64
	for s := range outputChannel {
		for _, v := range s {
			results = append(results, v)
		}
	}

	return results, totalProcessingTime
}

// createHostCountMap creates a map of host to their query counts
func createHostCountMap(queryInputs []QueryInput) map[string]int {
	hostQueryCount := make(map[string]int)
	for _, query := range queryInputs {
		if _, ok := hostQueryCount[query.HostName]; ok {
			hostQueryCount[query.HostName] += 1
		} else {
			hostQueryCount[query.HostName] = 1
		}
	}

	return hostQueryCount
}
