#!/bin/bash -e
if [ $# -lt 2 ]; then
    echo "Usage: $0 <Number of workers> <csv file path>"
    echo
    echo "eg: $0 5 test.csv "
    exit 1
fi

WORKERS=$1
FILE=$2

DBIsRunning=$(docker-compose ps --services --filter status=running | grep timescaledb)
AppExists=$(docker-compose ps --services | grep app)

if ! [[ $DBIsRunning && $AppExists ]];
then
    chmod +x data/init-db.sh
    docker-compose down --rmi local --volumes --remove-orphans
    docker-compose up --detach
fi

docker-compose run --rm app  ${WORKERS} ${FILE}

