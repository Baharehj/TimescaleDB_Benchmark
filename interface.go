package main

import "time"

// Interface for picking query load for each worker.
// To compare different load balancing algorithms/heuristics, different types can be created to implement the below interface;
// eg. GreedyLoadBalancer or MultifitLoadBalancer. The concrete type can then be injected into QueryRunner
type ILoadBalancer interface {
	GetBalancedLoads(hostCounts map[string]int, numberOfWorkers int) map[string]int
}

// Interface for database operations.
// The program can be used for different Database types as long as this interface is implemented;
// eg. TimescaleDB or MysqlDB. The concrete type can then be injected into QueryRunner
type IDB interface {
	RunCPUUsageQuery(hostname string, startTime time.Time, endTime time.Time) error
}
