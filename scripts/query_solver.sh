#!/bin/bash

output_file=""
while getopts "o:" opt; do
    case $opt in
        o) output_file="$OPTARG" ;;
        \?) echo "Usage: $0 -o <output_file> <games_file> <reviews_file>"
            exit 1 ;;
    esac
done
shift $((OPTIND - 1))

if [ -z "$output_file" ] || [ $# -ne 2 ]; then
    echo "Usage: $0 -o <output_file> <games_file> <reviews_file>"
    exit 1
fi

poetry run --directory ./query_solver python ./query_solver/results.py $1 $2 > ./$output_file
