#!/bin/bash

mkdir -p log_server_pb

protoc --go_out=./log_server_pb --go_opt=paths=source_relative \
    --go-grpc_out=./log_server_pb --go-grpc_opt=paths=source_relative \
    logs.proto