package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"
)

// Layout of the time given in the input to be parsed
const timeLayout = "2006-01-02 15:04:05"

type QueryInput struct {
	HostName  string
	StartTime time.Time
	EndTime   time.Time
}

func main() {
	// Get input arguments and validate
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println("Number of workers or test file path is missing")
		os.Exit(1)
	}

	// Get number of workers and file path from input
	workersCount, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Number of workers is not an integer")
		os.Exit(1)
	}
	csvFile := args[1]

	// Create a queryRunner
	queryRunner, err := createQueryRunner(workersCount)
	if err != nil {
		fmt.Println("failed to create the query runner, error: ", err.Error())
		os.Exit(1)
	}

	// Read csv and get number of queries per host
	queryInputs, err := readCSV(csvFile)
	if err != nil {
		fmt.Println("failed to read the csv file, error: ", err.Error())
		os.Exit(1)
	}

	// Run all the queries using the quey runner
	results, totalProcessingTime := queryRunner.Run(queryInputs)

	// Display the query time statistics
	err = displayBenchMarkStat(results, totalProcessingTime)
	if err != nil {
		fmt.Println("failed to display stat: ", err.Error())
		os.Exit(1)
	}
}

// readCSV reads a file with a given path that is relative to its current folder
// It returns a list of queries in the form of QueryInput struct
func readCSV(fileName string) ([]QueryInput, error) {
	inputFile, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close()

	reader := csv.NewReader(inputFile)
	reader.TrimLeadingSpace = true

	// Skip first line
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	// Read all lines in csv file
	csvLines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Iterate through the csv lines and create a QueryInput object for each line
	queryInputs := []QueryInput{}
	for _, line := range csvLines {
		// or each line, if startTime or endTime could not be parsed to date time,
		// ignore the line and continue processing the next lines
		hostName := line[0]
		startTime, err := time.Parse(timeLayout, line[1])
		if err != nil {
			continue
		}
		endTime, err := time.Parse(timeLayout, line[2])
		if err != nil {
			continue
		}

		query := QueryInput{
			HostName:  hostName,
			StartTime: startTime,
			EndTime:   endTime,
		}

		queryInputs = append(queryInputs, query)
	}

	return queryInputs, nil
}

// Create a query runner instance
// Query runner has load balancer, db and number of workers as dependencies
func createQueryRunner(workersCount int) (QueryRunner, error) {
	var queryRunner QueryRunner
	dbPort := os.Getenv("POSTGRES_PORT")
	dbName := os.Getenv("POSTGRES_DBNAME")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	db, err := NewTimeScaleDB(dbUser, dbPassword, dbPort, dbName, workersCount)
	if err != nil {
		return queryRunner, err
	}

	queryRunner = QueryRunner{
		LoadBalancer: GreedyLoadBalancer{},
		Db:           db,
		WorkersCount: workersCount,
	}

	return queryRunner, nil
}

// displayBenchMarkStat displays the statistics information
func displayBenchMarkStat(results []int64, totalProcessingTime int64) error {

	var median, min, max, mean, sum int64
	sort.Slice(results, func(i, j int) bool {
		return results[i] < results[j]
	})

	resultsCount := int64(len(results))
	sum = 0
	for _, result := range results {
		sum += result
	}

	if resultsCount%2 == 0 {
		median = (results[resultsCount/2-1] + results[resultsCount/2]) / 2
	} else {
		median = results[resultsCount/2]
	}

	min = results[0]
	max = results[resultsCount-1]
	mean = sum / resultsCount

	outputString := fmt.Sprintf(`
	Stats in Nanosecond: 

	Total number of queries ran: %s 
	Sum(total processing time): %s
	Mean: %s
	Median: %s
	Max: %s
	Min: %s
	`, formatInteger(resultsCount), formatInteger(totalProcessingTime), formatInteger(mean), formatInteger(median), formatInteger(max), formatInteger(min))

	fmt.Println(outputString)
	return nil
}

func formatInteger(num int64) string {
	str := fmt.Sprintf("%d", num)
	re := regexp.MustCompile("(\\d+)(\\d{3})")
	for n := ""; n != str; {
		n = str
		str = re.ReplaceAllString(str, "$1,$2")
	}
	return str
}
