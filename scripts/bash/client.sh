#!/bin/bash

current_dir=$(pwd)
wordkdir="../.."

cd $wordkdir
go run ./cmd/client/main.go ./cmd/client/config.json 
cd $current_dir
