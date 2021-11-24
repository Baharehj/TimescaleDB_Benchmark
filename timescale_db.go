package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type TimeScaleDB struct {
	// Connection pools are concurrency safe and using them increases the performance
	// as the cost of creating new connections is high. 
	connectionPool *pgxpool.Pool
}

// NewTimeScaleDB creates a new TimeScaleDB instance
// It returns a timeScaleDB db client
func NewTimeScaleDB(userName string, password string, port string, dbName string, connectionPoolCount int) (*TimeScaleDB, error) {
	/*
	We can pass the maximum number of connections in a connection pool as a parameter
	An optimal number of connections would likely depend on a variety of factors such as
	how the load is distributed among workers for that particular dataset, toal number of workers and queries, etc.
	A higher count does not always equal a faster total run time. For this assignment, setting the maxiumum number to the
	number of workers seem like a good heuristic to adopt.
	*/
	connStr := fmt.Sprintf("postgres://%s:%s@postgres:%s/%s?pool_max_conns=%d", userName, password, port, dbName, connectionPoolCount)

	dbpool, err := pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	return &TimeScaleDB{connectionPool: dbpool}, nil
}

// RunCPUUsageQuery Runs a query that measures cpu usage statistics for a given host
// in the time between a given start and end time.
// It returns any error that may happen during running of the query.
func (db TimeScaleDB) RunCPUUsageQuery(hostName string, startTime time.Time, endTime time.Time) error {

	// Origin parameter shifts the alignment of the bucket to the start given by the Start Time
	query := `Select time_bucket('1 minute', ts, origin=>$1) as minute, 
	min(usage) as min_usage, max(usage) as max_usage from cpu_usage where 
	ts >= $1 and ts <= $2 and host = $3 group by minute;`

	recs, err := db.connectionPool.Query(context.Background(), query, startTime, endTime, hostName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to execute query %v\n", err)
		return err
	}

	recs.Close()

	return nil
}
