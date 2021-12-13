# TimeScaleDB Query Benchmarker #
A tool for benchmarking a certain TimeScaleDB query.

Benchmarker takes the number of workers and a csv file path as input and run the queries and benchmarkes their run time.
The csv file should have the following headers and format: "hostname, start_time, end_time". Queries are then assigned to workers based on a greedy algorithm with the condition that all queries for a specific host should be ran by the same worker.


# Build and Run #
To run the application, you need to clone the reposiroty and run the following command within the `TimeScaleDB_Benchmark` folder; Please note that any test file created needs to be in the same folder. First time the script is ran, it will set up the timescaleDB and app containers and runs the program. Following runs will run the app image as an executable.

```
sh run.sh <Number of Workers> <Path to the csv file name>
eg. sh run.sh 5 test.csv
```

When done with testing, run the following command within the project folder to clean up all the docker containers, volumes and images:
```
docker-compose down --rmi all --volumes --remove-orphans
```

