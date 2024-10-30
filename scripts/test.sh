#! /usr/bin/env bash


set -eu

case $CLIENT_COUNT in
    ''|*[!0-9]*) echo "ERROR: CLIENT_COUNT must be an integer"
                 exit 1;;
    *) ;;
esac

echo "Generating pandas output"
set -x
./query_solver.sh -o "pandas_output.txt" $GAMES_PATH $REVIEWS_PATH

echo "Initializing docker"
sudo docker compose -f ../docker-compose.yaml up --detach
echo "Starting ${CLIENT_COUNT}"
seq $CLIENT_COUNT | parallel -j${CLIENT_COUNT} "go run ../cmd/client ../cmd/client/config.json > output_{}.txt"
PYTHON_BIN=$(command -v python || command -v python3 || { echo "python not found"; exit 1; })
echo "Computing diffs..."
seq $CLIENT_COUNT | parallel -j${CLIENT_COUNT} "${PYTHON_BIN} ./diff/main.py --first pandas_output.txt --second output_{}.txt > diff_{}.txt"
echo "Closing"
sudo docker compose -f ../docker-compose.yaml stop 
sudo docker compose -f ../docker-compose.yaml down 
set +x
