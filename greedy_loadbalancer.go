package main

import (
	"fmt"
	"sort"

	genericSlice "github.com/aclements/go-gg/generic/slice"
)

/*
I assume we want the task to be done as fast as possible so the workers would need a balanced load. 
In this specific case, we actually have all the queries at hand first. We can group the queries by host and 
get the number of queries per hosts and then balance the hosts across the workers so that workers run approximately the same number of queries.

	1. For simplicity, we can make the assumption that queries run in approximately the same amount of time. 
	   This would not necessarily be the case as the time interval for some queries could be very large or small and affect the query time. 

	2. Assuming we represent each host with an integer, we will have a list of integers that we want to divide between multiple buckets (number of our workers)
	 such that the largest sum of the numbers in each bucket is minimized. This is an NP-hard number-partitioning problem called Multi-Way Number Partitioning.
	  For the case of this assignment, I went for a simple greedy algorithm. 

	3. Worth mentioning that if the queries were being streamed to workers, the algorithm for assigning them could become completely different. 
	A simple greedy algorithm similar to what is implemented in the project would assign new hosts to workers with the least number of queued queries to run.
	A smarter heuristic could also take into account frequency of queries per host. For example, a worker that is expected to recieve new queries from a host already assigned to it
	 (based on the hosts historical trends) shouldn't necessarily take on a new host even if it has the least number of queries to run at the time.
*/

type GreedyLoadBalancer struct{}

type HostCount struct {
	HostName   string // Name of the host
	QueryCount int    // Count of the queries that will be ran by the host
}

// GetBalancedLoads Implements a basic greedy query load balancer.
// It returns a map of host names to the worker index they will be ran on
func (l GreedyLoadBalancer) GetBalancedLoads(hostCounts map[string]int, numWorkers int) map[string]int {

	// Create an array of type HostCount from hostCounts Map
	// The array will be sorted in descending order based on the count of queries for each host 
	var hostCountsArr []HostCount
	for k, v := range hostCounts {
		hostCountsArr = append(hostCountsArr, HostCount{HostName: k, QueryCount: v})
	}
	sort.Slice(hostCountsArr, func(i, j int) bool {
		return hostCountsArr[i].QueryCount > hostCountsArr[j].QueryCount
	})

	// sums array holds the total number of queries for each worker
	sums := make([]int, numWorkers)

	// At each iteration, one host is picked and it is assigned to the worker that currently has the minimum number of queries.
	// Number of queries for that host is then added to the sum of queries for that worker
	// and the loop continues to process and assign the next host.
	hostToWorker := make(map[string]int)
	var workerWithMinSum int
	for _, elem := range hostCountsArr {
		workerWithMinSum = genericSlice.ArgMin(sums)
		sums[workerWithMinSum] += elem.QueryCount
		hostToWorker[elem.HostName] = workerWithMinSum
	}

	fmt.Println("Number of queries assigned to each worker: ", sums)

	return hostToWorker
}
